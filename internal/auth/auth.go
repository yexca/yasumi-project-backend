package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

var ErrUnauthenticated = errors.New("auth: unauthenticated")

type User struct {
	ID          string
	Username    string
	Email       string
	DisplayName string
}

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (User, error)
}

type Clock interface {
	Now() time.Time
}

type SystemClock struct{}

func (SystemClock) Now() time.Time {
	return time.Now().UTC()
}

type DevBearerAuthenticator struct {
	token string
	user  User
}

func NewDevBearerAuthenticator(cfg config.AuthConfig) *DevBearerAuthenticator {
	return &DevBearerAuthenticator{
		token: cfg.DevToken,
		user: User{
			ID:          cfg.DevUserID,
			Username:    "local-dev",
			Email:       "local-dev@local.invalid",
			DisplayName: cfg.DevDisplayName,
		},
	}
}

func (a *DevBearerAuthenticator) Authenticate(_ context.Context, token string) (User, error) {
	if strings.TrimSpace(token) == "" {
		return User{}, ErrUnauthenticated
	}
	if subtle.ConstantTimeCompare([]byte(token), []byte(a.token)) != 1 {
		return User{}, ErrUnauthenticated
	}
	return a.user, nil
}

type AccountRepository interface {
	InTx(ctx context.Context, fn func(context.Context, accountTx) error) error
}

type accountTx interface {
	CreateAccount(ctx context.Context, user repository.UserRecord, credential repository.CredentialRecord, settings repository.UserSettingsRecord, onboardingItems []repository.ItemRecord) error
	FindAccountByIdentifier(ctx context.Context, identifier string) (repository.AccountWithCredentialRecord, error)
	GetCredentialForUpdate(ctx context.Context, userID string) (repository.CredentialRecord, error)
	GetUserForUpdate(ctx context.Context, userID string) (repository.UserRecord, error)
	CreateSession(ctx context.Context, session repository.SessionRecord) error
	GetActiveSessionWithUserForUpdate(ctx context.Context, sessionID string, now time.Time) (repository.SessionWithUserRecord, error)
	FindActiveSessionByRefreshHashForUpdate(ctx context.Context, refreshTokenHash string, now time.Time) (repository.SessionWithUserRecord, error)
	RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time, replacedBySessionID *string) error
	UpdateCredential(ctx context.Context, credential repository.CredentialRecord) error
	UpdateUserProfile(ctx context.Context, userID string, displayName *string, updatedAt time.Time) (repository.UserRecord, error)
}

type RepositoryAdapter struct {
	repo *repository.Repository
}

func NewRepositoryAdapter(repo *repository.Repository) RepositoryAdapter {
	return RepositoryAdapter{repo: repo}
}

func (a RepositoryAdapter) InTx(ctx context.Context, fn func(context.Context, accountTx) error) error {
	return a.repo.InTx(ctx, func(ctx context.Context, tx *repository.Tx) error {
		return fn(ctx, tx)
	})
}

type AccountService struct {
	repo              AccountRepository
	clock             Clock
	accessTokenSecret []byte
	accessTokenTTL    time.Duration
	sessionTTL        time.Duration
	hasher            PasswordHasher
}

type PasswordHasher struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

type RegisterRequest struct {
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	DisplayName *string `json:"display_name"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Password   string `json:"password"`
}

type UpdateProfileRequest struct {
	DisplayName *string `json:"display_name"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type AuthResponse struct {
	User    AccountUserDTO `json:"user"`
	Session AuthSessionDTO `json:"session"`
}

type AccountUserDTO struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
}

type AuthSessionDTO struct {
	Authenticated        bool   `json:"authenticated"`
	AccessToken          string `json:"access_token"`
	AccessTokenExpiresAt string `json:"access_token_expires_at"`
	RefreshToken         string `json:"refresh_token"`
	ExpiresAt            string `json:"expires_at"`
}

type accessTokenClaims struct {
	SessionID string `json:"session_id"`
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

func NewAccountService(repo AccountRepository, cfg config.Config, clock Clock) *AccountService {
	if clock == nil {
		clock = SystemClock{}
	}
	return &AccountService{
		repo:              repo,
		clock:             clock,
		accessTokenSecret: []byte(cfg.SyncToken.Secret),
		accessTokenTTL:    15 * time.Minute,
		sessionTTL:        30 * 24 * time.Hour,
		hasher: PasswordHasher{
			memory:      64 * 1024,
			iterations:  3,
			parallelism: 2,
			saltLength:  16,
			keyLength:   32,
		},
	}
}

func (s *AccountService) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	password := req.Password
	displayName := normalizeDisplayName(req.DisplayName, username)

	if err := validateUsername(username); err != nil {
		return AuthResponse{}, err
	}
	if err := validateEmail(email); err != nil {
		return AuthResponse{}, err
	}
	if err := validatePassword(password); err != nil {
		return AuthResponse{}, err
	}
	if displayName != nil && len(*displayName) > 120 {
		return AuthResponse{}, validationError("registration is invalid", domain.FieldDisplayName, "too_long")
	}

	now := s.clock.Now().UTC()
	userID, err := newUUID()
	if err != nil {
		return AuthResponse{}, serviceUnavailable("generate user id")
	}
	sessionID, err := newUUID()
	if err != nil {
		return AuthResponse{}, serviceUnavailable("generate session id")
	}
	refreshToken, err := randomToken(32)
	if err != nil {
		return AuthResponse{}, serviceUnavailable("generate refresh token")
	}
	passwordHash, params, err := s.hasher.Hash(password)
	if err != nil {
		return AuthResponse{}, serviceUnavailable("hash password")
	}
	expiresAt := now.Add(s.sessionTTL)

	user := repository.UserRecord{
		ID:          userID,
		Username:    username,
		Email:       email,
		DisplayName: displayName,
		Status:      "active",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	session := repository.SessionRecord{
		ID:               sessionID,
		UserID:           userID,
		RefreshTokenHash: hashToken(refreshToken),
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
	}
	onboardingItems, err := defaultOnboardingItems(userID, now)
	if err != nil {
		return AuthResponse{}, serviceUnavailable("generate onboarding items")
	}

	err = s.repo.InTx(ctx, func(ctx context.Context, tx accountTx) error {
		if err := tx.CreateAccount(ctx, user, repository.CredentialRecord{
			UserID:                userID,
			PasswordHash:          passwordHash,
			PasswordHashAlgorithm: "argon2id",
			PasswordHashParams:    params,
			PasswordChangedAt:     now,
			CreatedAt:             now,
			UpdatedAt:             now,
		}, defaultSettings(userID, now), onboardingItems); err != nil {
			return mapAccountRepositoryError(err)
		}
		if err := tx.CreateSession(ctx, session); err != nil {
			return mapAccountRepositoryError(err)
		}
		return nil
	})
	if err != nil {
		return AuthResponse{}, err
	}

	return s.authResponse(user, sessionID, refreshToken, expiresAt)
}

func (s *AccountService) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	identifier := strings.TrimSpace(req.Identifier)
	if identifier == "" {
		return AuthResponse{}, validationError("login is invalid", domain.FieldIdentifier, "required")
	}
	if req.Password == "" {
		return AuthResponse{}, validationError("login is invalid", domain.FieldPassword, "required")
	}

	now := s.clock.Now().UTC()
	var account repository.AccountWithCredentialRecord
	var sessionID string
	var refreshToken string
	var expiresAt time.Time

	err := s.repo.InTx(ctx, func(ctx context.Context, tx accountTx) error {
		found, err := tx.FindAccountByIdentifier(ctx, identifier)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return invalidCredentials()
			}
			return mapAccountRepositoryError(err)
		}
		if found.User.Status == "disabled" {
			return accountDisabled()
		}
		ok, err := s.hasher.Verify(req.Password, found.Credential.PasswordHash, found.Credential.PasswordHashParams)
		if err != nil || !ok {
			return invalidCredentials()
		}

		sessionID, err = newUUID()
		if err != nil {
			return serviceUnavailable("generate session id")
		}
		refreshToken, err = randomToken(32)
		if err != nil {
			return serviceUnavailable("generate refresh token")
		}
		expiresAt = now.Add(s.sessionTTL)
		if err := tx.CreateSession(ctx, repository.SessionRecord{
			ID:               sessionID,
			UserID:           found.User.ID,
			RefreshTokenHash: hashToken(refreshToken),
			CreatedAt:        now,
			ExpiresAt:        expiresAt,
		}); err != nil {
			return mapAccountRepositoryError(err)
		}
		account = found
		return nil
	})
	if err != nil {
		return AuthResponse{}, err
	}
	return s.authResponse(account.User, sessionID, refreshToken, expiresAt)
}

func (s *AccountService) Refresh(ctx context.Context, refreshToken string) (AuthResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return AuthResponse{}, sessionExpired()
	}

	now := s.clock.Now().UTC()
	var user repository.UserRecord
	var newSessionID string
	var newRefreshToken string
	var expiresAt time.Time

	err := s.repo.InTx(ctx, func(ctx context.Context, tx accountTx) error {
		current, err := tx.FindActiveSessionByRefreshHashForUpdate(ctx, hashToken(refreshToken), now)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return sessionExpired()
			}
			return mapAccountRepositoryError(err)
		}
		if current.User.Status == "disabled" {
			return accountDisabled()
		}

		newSessionID, err = newUUID()
		if err != nil {
			return serviceUnavailable("generate session id")
		}
		newRefreshToken, err = randomToken(32)
		if err != nil {
			return serviceUnavailable("generate refresh token")
		}
		expiresAt = now.Add(s.sessionTTL)
		if err := tx.CreateSession(ctx, repository.SessionRecord{
			ID:               newSessionID,
			UserID:           current.User.ID,
			RefreshTokenHash: hashToken(newRefreshToken),
			CreatedAt:        now,
			ExpiresAt:        expiresAt,
		}); err != nil {
			return mapAccountRepositoryError(err)
		}
		if err := tx.RevokeSession(ctx, current.Session.ID, now, &newSessionID); err != nil {
			return mapAccountRepositoryError(err)
		}
		user = current.User
		return nil
	})
	if err != nil {
		return AuthResponse{}, err
	}
	return s.authResponse(user, newSessionID, newRefreshToken, expiresAt)
}

func (s *AccountService) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.parseAccessToken(accessToken)
	if err != nil {
		return ErrUnauthenticated
	}
	now := s.clock.Now().UTC()
	if claims.ExpiresAt <= now.Unix() {
		return ErrUnauthenticated
	}

	return s.repo.InTx(ctx, func(ctx context.Context, tx accountTx) error {
		session, err := tx.GetActiveSessionWithUserForUpdate(ctx, claims.SessionID, now)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrUnauthenticated
			}
			return mapAccountRepositoryError(err)
		}
		if session.User.Status == "disabled" {
			return accountDisabled()
		}
		return tx.RevokeSession(ctx, session.Session.ID, now, nil)
	})
}

func (s *AccountService) UpdateProfile(ctx context.Context, userID string, req UpdateProfileRequest) (AccountUserDTO, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return AccountUserDTO{}, ErrUnauthenticated
	}
	displayName := normalizeOptionalDisplayName(req.DisplayName)
	if displayName != nil && len(*displayName) > 120 {
		return AccountUserDTO{}, validationError("profile is invalid", domain.FieldDisplayName, "too_long")
	}

	now := s.clock.Now().UTC()
	var user repository.UserRecord
	err := s.repo.InTx(ctx, func(ctx context.Context, tx accountTx) error {
		updated, err := tx.UpdateUserProfile(ctx, userID, displayName, now)
		if err != nil {
			return mapAccountRepositoryError(err)
		}
		if updated.Status == "disabled" {
			return accountDisabled()
		}
		user = updated
		return nil
	})
	if err != nil {
		return AccountUserDTO{}, err
	}

	return accountUserDTO(user), nil
}

func (s *AccountService) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return ErrUnauthenticated
	}
	if strings.TrimSpace(req.CurrentPassword) == "" {
		return validationError("password change is invalid", domain.FieldPassword, "current_required")
	}
	if err := validatePassword(req.NewPassword); err != nil {
		return err
	}

	now := s.clock.Now().UTC()
	return s.repo.InTx(ctx, func(ctx context.Context, tx accountTx) error {
		user, err := tx.GetUserForUpdate(ctx, userID)
		if err != nil {
			return mapAccountRepositoryError(err)
		}
		if user.Status == "disabled" {
			return accountDisabled()
		}

		credential, err := tx.GetCredentialForUpdate(ctx, userID)
		if err != nil {
			return mapAccountRepositoryError(err)
		}
		ok, err := s.hasher.Verify(req.CurrentPassword, credential.PasswordHash, credential.PasswordHashParams)
		if err != nil || !ok {
			return invalidCredentials()
		}

		passwordHash, params, err := s.hasher.Hash(req.NewPassword)
		if err != nil {
			return serviceUnavailable("hash password")
		}
		credential.PasswordHash = passwordHash
		credential.PasswordHashAlgorithm = "argon2id"
		credential.PasswordHashParams = params
		credential.PasswordChangedAt = now
		credential.UpdatedAt = now
		if err := tx.UpdateCredential(ctx, credential); err != nil {
			return mapAccountRepositoryError(err)
		}
		return nil
	})
}

func (s *AccountService) Authenticate(ctx context.Context, token string) (User, error) {
	claims, err := s.parseAccessToken(token)
	if err != nil {
		return User{}, ErrUnauthenticated
	}
	now := s.clock.Now().UTC()
	if claims.ExpiresAt <= now.Unix() {
		return User{}, ErrUnauthenticated
	}

	var out User
	err = s.repo.InTx(ctx, func(ctx context.Context, tx accountTx) error {
		session, err := tx.GetActiveSessionWithUserForUpdate(ctx, claims.SessionID, now)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ErrUnauthenticated
			}
			return mapAccountRepositoryError(err)
		}
		if session.User.ID != claims.UserID {
			return ErrUnauthenticated
		}
		if session.User.Status == "disabled" {
			return accountDisabled()
		}
		out = userFromRecord(session.User)
		return nil
	})
	if err != nil {
		return User{}, err
	}
	return out, nil
}

func (s *AccountService) authResponse(user repository.UserRecord, sessionID, refreshToken string, sessionExpiresAt time.Time) (AuthResponse, error) {
	accessTokenExpiresAt := s.clock.Now().UTC().Add(s.accessTokenTTL)
	accessToken, err := s.signAccessToken(accessTokenClaims{
		SessionID: sessionID,
		UserID:    user.ID,
		ExpiresAt: accessTokenExpiresAt.Unix(),
	})
	if err != nil {
		return AuthResponse{}, serviceUnavailable("sign access token")
	}
	return AuthResponse{
		User: accountUserDTO(user),
		Session: AuthSessionDTO{
			Authenticated:        true,
			AccessToken:          accessToken,
			AccessTokenExpiresAt: accessTokenExpiresAt.Format(time.RFC3339),
			RefreshToken:         refreshToken,
			ExpiresAt:            sessionExpiresAt.UTC().Format(time.RFC3339),
		},
	}, nil
}

func (s *AccountService) signAccessToken(claims accessTokenClaims) (string, error) {
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.RawURLEncoding.EncodeToString(payload)
	mac := hmac.New(sha256.New, s.accessTokenSecret)
	mac.Write([]byte(encodedPayload))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return encodedPayload + "." + signature, nil
}

func (s *AccountService) parseAccessToken(token string) (accessTokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return accessTokenClaims{}, ErrUnauthenticated
	}
	mac := hmac.New(sha256.New, s.accessTokenSecret)
	mac.Write([]byte(parts[0]))
	want := mac.Sum(nil)
	got, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil || !hmac.Equal(got, want) {
		return accessTokenClaims{}, ErrUnauthenticated
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return accessTokenClaims{}, ErrUnauthenticated
	}
	var claims accessTokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return accessTokenClaims{}, ErrUnauthenticated
	}
	if claims.SessionID == "" || claims.UserID == "" || claims.ExpiresAt == 0 {
		return accessTokenClaims{}, ErrUnauthenticated
	}
	return claims, nil
}

func (h PasswordHasher) Hash(password string) (string, json.RawMessage, error) {
	salt, err := randomBytes(h.saltLength)
	if err != nil {
		return "", nil, err
	}
	key := argon2.IDKey([]byte(password), salt, h.iterations, h.memory, h.parallelism, h.keyLength)
	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		h.memory,
		h.iterations,
		h.parallelism,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	)
	params, err := json.Marshal(map[string]any{
		"memory":      h.memory,
		"iterations":  h.iterations,
		"parallelism": h.parallelism,
		"salt_length": h.saltLength,
		"key_length":  h.keyLength,
	})
	if err != nil {
		return "", nil, err
	}
	return encoded, params, nil
}

func (h PasswordHasher) Verify(password, encoded string, _ json.RawMessage) (bool, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
		return false, nil
	}
	paramParts := strings.Split(parts[3], ",")
	if len(paramParts) != 3 {
		return false, nil
	}
	memory, err := parseUint32Param(paramParts[0], "m=")
	if err != nil {
		return false, nil
	}
	iterations, err := parseUint32Param(paramParts[1], "t=")
	if err != nil {
		return false, nil
	}
	parallelism64, err := parseUint32Param(paramParts[2], "p=")
	if err != nil || parallelism64 > 255 {
		return false, nil
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, nil
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, nil
	}
	got := argon2.IDKey([]byte(password), salt, iterations, memory, uint8(parallelism64), uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

func parseUint32Param(value, prefix string) (uint32, error) {
	if !strings.HasPrefix(value, prefix) {
		return 0, fmt.Errorf("missing prefix")
	}
	parsed, err := strconv.ParseUint(strings.TrimPrefix(value, prefix), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(parsed), nil
}

func validateUsername(username string) error {
	if username == "" {
		return validationError("registration is invalid", domain.FieldUsername, "required")
	}
	if len(username) < 3 {
		return validationError("registration is invalid", domain.FieldUsername, "too_short")
	}
	if len(username) > 40 {
		return validationError("registration is invalid", domain.FieldUsername, "too_long")
	}
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return validationError("registration is invalid", domain.FieldUsername, "invalid_format")
	}
	return nil
}

func validateEmail(email string) error {
	if email == "" {
		return validationError("registration is invalid", domain.FieldEmail, "required")
	}
	if len(email) > 254 {
		return validationError("registration is invalid", domain.FieldEmail, "too_long")
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email || !strings.Contains(email, ".") {
		return validationError("registration is invalid", domain.FieldEmail, "invalid_format")
	}
	return nil
}

func validatePassword(password string) error {
	if password == "" {
		return validationError("registration is invalid", domain.FieldPassword, "required")
	}
	if len(password) < 8 {
		return validationError("registration is invalid", domain.FieldPassword, "too_short")
	}
	if len(password) > 1024 {
		return validationError("registration is invalid", domain.FieldPassword, "too_long")
	}
	return nil
}

func normalizeDisplayName(value *string, username string) *string {
	if value == nil {
		return &username
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return &username
	}
	return &trimmed
}

func normalizeOptionalDisplayName(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func defaultSettings(userID string, now time.Time) repository.UserSettingsRecord {
	return repository.UserSettingsRecord{
		UserID:                    userID,
		Language:                  "ja",
		Locale:                    "ja-JP",
		WeekStartDay:              "monday",
		TimeZone:                  "Asia/Tokyo",
		DateDisplayFormat:         "yyyy-MM-dd",
		TimeDisplayFormat:         "24h",
		DefaultTimeZoneMode:       "floating",
		TodayPrimaryLookaheadDays: 3,
		DeadlineAwarenessDays:     14,
		WeatherCity:               "Tokyo",
		CreatedAt:                 now,
		UpdatedAt:                 now,
		ClientUpdatedAt:           now,
		ServerUpdatedAt:           now,
		CreatedByDeviceID:         "account-registration",
		UpdatedByDeviceID:         "account-registration",
		Revision:                  1,
	}
}

func defaultOnboardingItems(userID string, now time.Time) ([]repository.ItemRecord, error) {
	localDate := now.In(time.FixedZone("JST", 9*60*60))
	today := localDate.Format("2006-01-02")
	nextWeek := localDate.AddDate(0, 0, 7).Format("2006-01-02")
	reviewDate := localDate.AddDate(0, 0, 14).Format("2006-01-02")
	deviceID := "account-registration"
	floating := "floating"
	dateOnly := "date_only"
	effortSmall := 1
	effortMedium := 2
	importanceNormal := 3
	importanceHigh := 4
	items := []repository.ItemRecord{
		{
			UserID:             userID,
			ItemType:           "inbox",
			Title:              "Capture a quick thought here",
			Note:               stringPtr("Inbox items are for things you have not organized yet."),
			Status:             "active",
			QuickAddSourceText: stringPtr("Try Quick Add or save a loose thought"),
		},
		{
			UserID:                userID,
			ItemType:              "date_task",
			Title:                 "Plan today's focus",
			Note:                  stringPtr("A date task appears on a specific day."),
			Status:                "active",
			ScheduledDate:         &today,
			ScheduledTimeZoneMode: &floating,
			EstimatedEffort:       &effortSmall,
			Importance:            &importanceNormal,
		},
		{
			UserID:               userID,
			ItemType:             "deadline_task",
			Title:                "Review a sample deadline",
			Note:                 stringPtr("A deadline task keeps the due date visible while you plan work."),
			Status:               "active",
			PlannedWorkDate:      &today,
			DeadlineDate:         &nextWeek,
			DeadlineTimeZoneMode: &dateOnly,
			EstimatedEffort:      &effortMedium,
			Importance:           &importanceHigh,
		},
		{
			UserID:          userID,
			ItemType:        "idea",
			Title:           "Save an idea to revisit",
			Note:            stringPtr("Ideas stay visible without becoming overdue tasks."),
			Status:          "active",
			ReviewDate:      &reviewDate,
			EstimatedEffort: &effortSmall,
			Importance:      &importanceNormal,
		},
	}
	for i := range items {
		id, err := newUUID()
		if err != nil {
			return nil, err
		}
		items[i].ID = id
		items[i].PressureMetadata = []byte("{}")
		items[i].CreatedAt = now
		items[i].UpdatedAt = now
		items[i].ClientUpdatedAt = now
		items[i].ServerUpdatedAt = now
		items[i].CreatedByDeviceID = deviceID
		items[i].UpdatedByDeviceID = deviceID
		items[i].Revision = 1
	}
	return items, nil
}

func accountUserDTO(user repository.UserRecord) AccountUserDTO {
	return AccountUserDTO{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}
}

func stringPtr(value string) *string {
	return &value
}

func mapAccountRepositoryError(err error) error {
	if errors.Is(err, repository.ErrNotFound) {
		return ErrUnauthenticated
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.ConstraintName {
		case "users_username_ci_unique":
			return &domain.Error{
				Code:      domain.ErrorUsernameAlreadyTaken,
				Message:   "username is already taken",
				Fields:    map[domain.FieldKey]string{domain.FieldUsername: "already_taken"},
				Retryable: false,
			}
		case "users_email_ci_unique":
			return &domain.Error{
				Code:      domain.ErrorEmailAlreadyRegistered,
				Message:   "email is already registered",
				Fields:    map[domain.FieldKey]string{domain.FieldEmail: "already_registered"},
				Retryable: false,
			}
		}
	}
	return &domain.Error{
		Code:      domain.ErrorValidationFailed,
		Message:   "account write failed validation",
		Fields:    map[domain.FieldKey]string{},
		Retryable: false,
	}
}

func validationError(message string, field domain.FieldKey, reason string) error {
	return &domain.Error{
		Code:      domain.ErrorValidationFailed,
		Message:   message,
		Fields:    map[domain.FieldKey]string{field: reason},
		Retryable: false,
	}
}

func invalidCredentials() error {
	return &domain.Error{
		Code:      domain.ErrorInvalidCredentials,
		Message:   "invalid credentials",
		Fields:    map[domain.FieldKey]string{},
		Retryable: false,
	}
}

func sessionExpired() error {
	return &domain.Error{
		Code:      domain.ErrorSessionExpired,
		Message:   "session expired",
		Fields:    map[domain.FieldKey]string{},
		Retryable: false,
	}
}

func accountDisabled() error {
	return &domain.Error{
		Code:      domain.ErrorAccountDisabled,
		Message:   "account is disabled",
		Fields:    map[domain.FieldKey]string{},
		Retryable: false,
	}
}

func serviceUnavailable(message string) error {
	return &domain.Error{
		Code:      domain.ErrorServiceUnavailable,
		Message:   message,
		Fields:    map[domain.FieldKey]string{},
		Retryable: true,
	}
}

func userFromRecord(user repository.UserRecord) User {
	displayName := ""
	if user.DisplayName != nil {
		displayName = *user.DisplayName
	}
	return User{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: displayName,
	}
}

func newUUID() (string, error) {
	b, err := randomBytes(16)
	if err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

func randomToken(length uint32) (string, error) {
	b, err := randomBytes(length)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func randomBytes(length uint32) ([]byte, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
