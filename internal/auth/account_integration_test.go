package auth_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/yasumi/yasumi-project-backend/internal/auth"
	"github.com/yasumi/yasumi-project-backend/internal/config"
	"github.com/yasumi/yasumi-project-backend/internal/domain"
	"github.com/yasumi/yasumi-project-backend/internal/migrations"
	"github.com/yasumi/yasumi-project-backend/internal/repository"
)

func TestAccountRegisterLoginRefreshAndLogout(t *testing.T) {
	pool := newTestPool(t)
	service := newAccountService(pool)
	ctx := context.Background()

	registered, err := service.Register(ctx, auth.RegisterRequest{
		Username: "yasumi_user",
		Email:    "USER@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if registered.User.ID == "" || registered.User.Email != "user@example.com" {
		t.Fatalf("registered user = %+v", registered.User)
	}
	if registered.Session.AccessToken == "" || registered.Session.RefreshToken == "" {
		t.Fatalf("missing session tokens: %+v", registered.Session)
	}

	assertAccountRows(t, pool, registered.User.ID, "password123", registered.Session.RefreshToken)

	if _, err := service.Authenticate(ctx, registered.Session.AccessToken); err != nil {
		t.Fatalf("authenticate registered session: %v", err)
	}

	loginByUsername, err := service.Login(ctx, auth.LoginRequest{
		Identifier: "yasumi_user",
		Password:   "password123",
	})
	if err != nil {
		t.Fatalf("login by username: %v", err)
	}
	if loginByUsername.User.ID != registered.User.ID {
		t.Fatalf("login user id = %s, want %s", loginByUsername.User.ID, registered.User.ID)
	}

	if _, err := service.Login(ctx, auth.LoginRequest{
		Identifier: "user@example.com",
		Password:   "password123",
	}); err != nil {
		t.Fatalf("login by email: %v", err)
	}

	for _, req := range []auth.LoginRequest{
		{Identifier: "unknown", Password: "password123"},
		{Identifier: "yasumi_user", Password: "wrong-password"},
	} {
		_, err := service.Login(ctx, req)
		assertDomainCode(t, err, domain.ErrorInvalidCredentials)
	}

	refreshed, err := service.Refresh(ctx, registered.Session.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if refreshed.Session.RefreshToken == registered.Session.RefreshToken {
		t.Fatal("refresh token was not rotated")
	}
	_, err = service.Refresh(ctx, registered.Session.RefreshToken)
	assertDomainCode(t, err, domain.ErrorSessionExpired)

	if err := service.Logout(ctx, refreshed.Session.AccessToken); err != nil {
		t.Fatalf("logout: %v", err)
	}
	if _, err := service.Authenticate(ctx, refreshed.Session.AccessToken); !errors.Is(err, auth.ErrUnauthenticated) {
		t.Fatalf("authenticate revoked session error = %v, want unauthenticated", err)
	}
}

func TestAccountRegisterDuplicateUsernameAndEmail(t *testing.T) {
	pool := newTestPool(t)
	service := newAccountService(pool)
	ctx := context.Background()

	if _, err := service.Register(ctx, auth.RegisterRequest{
		Username: "yasumi_user",
		Email:    "user@example.com",
		Password: "password123",
	}); err != nil {
		t.Fatalf("register seed: %v", err)
	}

	_, err := service.Register(ctx, auth.RegisterRequest{
		Username: "YASUMI_USER",
		Email:    "other@example.com",
		Password: "password123",
	})
	assertDomainCode(t, err, domain.ErrorUsernameAlreadyTaken)

	_, err = service.Register(ctx, auth.RegisterRequest{
		Username: "other_user",
		Email:    "USER@example.com",
		Password: "password123",
	})
	assertDomainCode(t, err, domain.ErrorEmailAlreadyRegistered)
}

func TestDisabledAccountCannotAuthenticate(t *testing.T) {
	pool := newTestPool(t)
	service := newAccountService(pool)
	ctx := context.Background()

	registered, err := service.Register(ctx, auth.RegisterRequest{
		Username: "disabled_user",
		Email:    "disabled@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := pool.Exec(ctx, "update users set status = 'disabled' where id = $1", registered.User.ID); err != nil {
		t.Fatalf("disable user: %v", err)
	}

	_, err = service.Login(ctx, auth.LoginRequest{Identifier: "disabled_user", Password: "password123"})
	assertDomainCode(t, err, domain.ErrorAccountDisabled)

	_, err = service.Authenticate(ctx, registered.Session.AccessToken)
	assertDomainCode(t, err, domain.ErrorAccountDisabled)

	_, err = service.Refresh(ctx, registered.Session.RefreshToken)
	assertDomainCode(t, err, domain.ErrorAccountDisabled)
}

func assertAccountRows(t *testing.T, pool *pgxpool.Pool, userID, rawPassword, rawRefreshToken string) {
	t.Helper()
	ctx := context.Background()
	var userCount int
	if err := pool.QueryRow(ctx, "select count(*) from users where id = $1", userID).Scan(&userCount); err != nil {
		t.Fatalf("count user: %v", err)
	}
	if userCount != 1 {
		t.Fatalf("user count = %d, want 1", userCount)
	}

	var passwordHash string
	if err := pool.QueryRow(ctx, "select password_hash from user_credentials where user_id = $1", userID).Scan(&passwordHash); err != nil {
		t.Fatalf("read password hash: %v", err)
	}
	if passwordHash == rawPassword || !strings.HasPrefix(passwordHash, "$argon2id$") {
		t.Fatalf("password hash was not stored safely: %q", passwordHash)
	}

	var refreshHash string
	if err := pool.QueryRow(ctx, "select refresh_token_hash from user_sessions where user_id = $1", userID).Scan(&refreshHash); err != nil {
		t.Fatalf("read refresh token hash: %v", err)
	}
	if refreshHash == rawRefreshToken || refreshHash == "" {
		t.Fatalf("refresh token hash was not stored safely")
	}

	var settingsCount int
	if err := pool.QueryRow(ctx, "select count(*) from user_settings where user_id = $1", userID).Scan(&settingsCount); err != nil {
		t.Fatalf("count settings: %v", err)
	}
	if settingsCount != 1 {
		t.Fatalf("settings count = %d, want 1", settingsCount)
	}
}

func newAccountService(pool *pgxpool.Pool) *auth.AccountService {
	cfg := config.MustLoad()
	cfg.SyncToken.Secret = "test-account-secret"
	return auth.NewAccountService(auth.NewRepositoryAdapter(repository.New(pool)), cfg, auth.SystemClock{})
}

func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	dsn := os.Getenv("YASUMI_TEST_DATABASE_URL")
	if dsn == "" {
		cfg := config.MustLoad()
		dsn = cfg.Postgres.DSN()
	}

	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("open postgres pool: %v", err)
	}
	if err := adminPool.Ping(ctx); err != nil {
		adminPool.Close()
		t.Skipf("PostgreSQL integration tests skipped: %v", err)
	}

	schemaName := "test_" + strings.ToLower(strings.ReplaceAll(t.Name(), "/", "_"))
	schemaName = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return '_'
	}, schemaName)
	quotedSchema := pgxIdentifier(schemaName)
	if _, err := adminPool.Exec(ctx, "drop schema if exists "+quotedSchema+" cascade"); err != nil {
		t.Fatalf("drop test schema: %v", err)
	}
	if _, err := adminPool.Exec(ctx, "create schema "+quotedSchema); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		_, _ = adminPool.Exec(context.Background(), "drop schema if exists "+quotedSchema+" cascade")
		adminPool.Close()
	})

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("parse test postgres config: %v", err)
	}
	poolConfig.ConnConfig.RuntimeParams["search_path"] = schemaName

	testPool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		t.Fatalf("open test postgres pool: %v", err)
	}
	t.Cleanup(testPool.Close)
	if err := migrations.Apply(ctx, testPool); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}
	return testPool
}

func pgxIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func assertDomainCode(t *testing.T, err error, code domain.ErrorCode) {
	t.Helper()
	var domainErr *domain.Error
	if !errors.As(err, &domainErr) || domainErr.Code != code {
		t.Fatalf("error = %v, want domain code %s", err, code)
	}
}
