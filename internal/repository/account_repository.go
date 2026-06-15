package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

func (tx *Tx) CreateAccount(ctx context.Context, user UserRecord, credential CredentialRecord, settings UserSettingsRecord, onboardingItems []ItemRecord) error {
	const userQuery = `
		insert into users (
			id, username, email, email_verified_at, display_name, status, created_at, updated_at
		) values ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	if _, err := tx.tx.Exec(ctx, userQuery,
		user.ID,
		user.Username,
		user.Email,
		user.EmailVerifiedAt,
		user.DisplayName,
		user.Status,
		user.CreatedAt,
		user.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	const credentialQuery = `
		insert into user_credentials (
			user_id, password_hash, password_hash_algorithm, password_hash_params,
			password_changed_at, created_at, updated_at
		) values ($1, $2, $3, coalesce($4, '{}'::jsonb), $5, $6, $7)
	`
	if len(credential.PasswordHashParams) == 0 {
		credential.PasswordHashParams = []byte("{}")
	}
	if _, err := tx.tx.Exec(ctx, credentialQuery,
		credential.UserID,
		credential.PasswordHash,
		credential.PasswordHashAlgorithm,
		credential.PasswordHashParams,
		credential.PasswordChangedAt,
		credential.CreatedAt,
		credential.UpdatedAt,
	); err != nil {
		return fmt.Errorf("create user credential: %w", err)
	}

	if err := tx.UpsertUserSettings(ctx, settings); err != nil {
		return err
	}
	for _, item := range onboardingItems {
		if err := tx.UpsertItem(ctx, item); err != nil {
			return err
		}
	}
	return nil
}

func (tx *Tx) FindAccountByIdentifier(ctx context.Context, identifier string) (AccountWithCredentialRecord, error) {
	const query = `
		select u.id::text, u.username, u.email, u.email_verified_at, u.display_name,
			u.status, u.created_at, u.updated_at,
			c.user_id::text, c.password_hash, c.password_hash_algorithm,
			c.password_hash_params, c.password_changed_at, c.created_at, c.updated_at
		from users u
		join user_credentials c on c.user_id = u.id
		where lower(u.username) = lower($1) or lower(u.email) = lower($1)
	`
	var row AccountWithCredentialRecord
	if err := tx.tx.QueryRow(ctx, query, identifier).Scan(
		&row.User.ID,
		&row.User.Username,
		&row.User.Email,
		&row.User.EmailVerifiedAt,
		&row.User.DisplayName,
		&row.User.Status,
		&row.User.CreatedAt,
		&row.User.UpdatedAt,
		&row.Credential.UserID,
		&row.Credential.PasswordHash,
		&row.Credential.PasswordHashAlgorithm,
		&row.Credential.PasswordHashParams,
		&row.Credential.PasswordChangedAt,
		&row.Credential.CreatedAt,
		&row.Credential.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AccountWithCredentialRecord{}, ErrNotFound
		}
		return AccountWithCredentialRecord{}, fmt.Errorf("find account by identifier: %w", err)
	}
	return row, nil
}

func (tx *Tx) GetUserForUpdate(ctx context.Context, userID string) (UserRecord, error) {
	const query = `
		select id::text, username, email, email_verified_at, display_name,
			status, created_at, updated_at
		from users
		where id = $1
		for update
	`
	var row UserRecord
	if err := tx.tx.QueryRow(ctx, query, userID).Scan(
		&row.ID,
		&row.Username,
		&row.Email,
		&row.EmailVerifiedAt,
		&row.DisplayName,
		&row.Status,
		&row.CreatedAt,
		&row.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserRecord{}, ErrNotFound
		}
		return UserRecord{}, fmt.Errorf("get user for update: %w", err)
	}
	return row, nil
}

func (tx *Tx) UpdateUserProfile(ctx context.Context, userID string, displayName *string, updatedAt time.Time) (UserRecord, error) {
	const query = `
		update users
		set display_name = $2,
			updated_at = $3
		where id = $1
		returning id::text, username, email, email_verified_at, display_name,
			status, created_at, updated_at
	`
	var row UserRecord
	if err := tx.tx.QueryRow(ctx, query, userID, displayName, updatedAt).Scan(
		&row.ID,
		&row.Username,
		&row.Email,
		&row.EmailVerifiedAt,
		&row.DisplayName,
		&row.Status,
		&row.CreatedAt,
		&row.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return UserRecord{}, ErrNotFound
		}
		return UserRecord{}, fmt.Errorf("update user profile: %w", err)
	}
	return row, nil
}

func (tx *Tx) GetCredentialForUpdate(ctx context.Context, userID string) (CredentialRecord, error) {
	const query = `
		select user_id::text, password_hash, password_hash_algorithm,
			password_hash_params, password_changed_at, created_at, updated_at
		from user_credentials
		where user_id = $1
		for update
	`
	var row CredentialRecord
	if err := tx.tx.QueryRow(ctx, query, userID).Scan(
		&row.UserID,
		&row.PasswordHash,
		&row.PasswordHashAlgorithm,
		&row.PasswordHashParams,
		&row.PasswordChangedAt,
		&row.CreatedAt,
		&row.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CredentialRecord{}, ErrNotFound
		}
		return CredentialRecord{}, fmt.Errorf("get credential for update: %w", err)
	}
	return row, nil
}

func (tx *Tx) UpdateCredential(ctx context.Context, credential CredentialRecord) error {
	const query = `
		update user_credentials
		set password_hash = $2,
			password_hash_algorithm = $3,
			password_hash_params = coalesce($4, '{}'::jsonb),
			password_changed_at = $5,
			updated_at = $6
		where user_id = $1
	`
	if len(credential.PasswordHashParams) == 0 {
		credential.PasswordHashParams = []byte("{}")
	}
	tag, err := tx.tx.Exec(ctx, query,
		credential.UserID,
		credential.PasswordHash,
		credential.PasswordHashAlgorithm,
		credential.PasswordHashParams,
		credential.PasswordChangedAt,
		credential.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update credential: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (tx *Tx) CreateSession(ctx context.Context, session SessionRecord) error {
	const query = `
		insert into user_sessions (
			id, user_id, refresh_token_hash, created_at, last_used_at, expires_at,
			revoked_at, replaced_by_session_id, user_agent, ip_hash
		) values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	if _, err := tx.tx.Exec(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshTokenHash,
		session.CreatedAt,
		session.LastUsedAt,
		session.ExpiresAt,
		session.RevokedAt,
		session.ReplacedBySessionID,
		session.UserAgent,
		session.IPHash,
	); err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (tx *Tx) GetActiveSessionWithUserForUpdate(ctx context.Context, sessionID string, now time.Time) (SessionWithUserRecord, error) {
	const query = `
		select s.id::text, s.user_id::text, s.refresh_token_hash, s.created_at,
			s.last_used_at, s.expires_at, s.revoked_at, s.replaced_by_session_id::text,
			s.user_agent, s.ip_hash,
			u.id::text, u.username, u.email, u.email_verified_at, u.display_name,
			u.status, u.created_at, u.updated_at
		from user_sessions s
		join users u on u.id = s.user_id
		where s.id = $1 and s.revoked_at is null and s.expires_at > $2
		for update of s
	`
	var row SessionWithUserRecord
	if err := tx.tx.QueryRow(ctx, query, sessionID, now).Scan(
		&row.Session.ID,
		&row.Session.UserID,
		&row.Session.RefreshTokenHash,
		&row.Session.CreatedAt,
		&row.Session.LastUsedAt,
		&row.Session.ExpiresAt,
		&row.Session.RevokedAt,
		&row.Session.ReplacedBySessionID,
		&row.Session.UserAgent,
		&row.Session.IPHash,
		&row.User.ID,
		&row.User.Username,
		&row.User.Email,
		&row.User.EmailVerifiedAt,
		&row.User.DisplayName,
		&row.User.Status,
		&row.User.CreatedAt,
		&row.User.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionWithUserRecord{}, ErrNotFound
		}
		return SessionWithUserRecord{}, fmt.Errorf("get active session with user: %w", err)
	}
	return row, nil
}

func (tx *Tx) FindActiveSessionByRefreshHashForUpdate(ctx context.Context, refreshTokenHash string, now time.Time) (SessionWithUserRecord, error) {
	const query = `
		select s.id::text, s.user_id::text, s.refresh_token_hash, s.created_at,
			s.last_used_at, s.expires_at, s.revoked_at, s.replaced_by_session_id::text,
			s.user_agent, s.ip_hash,
			u.id::text, u.username, u.email, u.email_verified_at, u.display_name,
			u.status, u.created_at, u.updated_at
		from user_sessions s
		join users u on u.id = s.user_id
		where s.refresh_token_hash = $1 and s.revoked_at is null and s.expires_at > $2
		for update of s
	`
	var row SessionWithUserRecord
	if err := tx.tx.QueryRow(ctx, query, refreshTokenHash, now).Scan(
		&row.Session.ID,
		&row.Session.UserID,
		&row.Session.RefreshTokenHash,
		&row.Session.CreatedAt,
		&row.Session.LastUsedAt,
		&row.Session.ExpiresAt,
		&row.Session.RevokedAt,
		&row.Session.ReplacedBySessionID,
		&row.Session.UserAgent,
		&row.Session.IPHash,
		&row.User.ID,
		&row.User.Username,
		&row.User.Email,
		&row.User.EmailVerifiedAt,
		&row.User.DisplayName,
		&row.User.Status,
		&row.User.CreatedAt,
		&row.User.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionWithUserRecord{}, ErrNotFound
		}
		return SessionWithUserRecord{}, fmt.Errorf("find active session by refresh hash: %w", err)
	}
	return row, nil
}

func (tx *Tx) RevokeSession(ctx context.Context, sessionID string, revokedAt time.Time, replacedBySessionID *string) error {
	const query = `
		update user_sessions
		set revoked_at = $2,
			replaced_by_session_id = $3,
			last_used_at = $2
		where id = $1 and revoked_at is null
	`
	tag, err := tx.tx.Exec(ctx, query, sessionID, revokedAt, replacedBySessionID)
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
