package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/yasumi/yasumi-project-backend/internal/config"
)

var ErrUnauthenticated = errors.New("auth: unauthenticated")

type User struct {
	ID          string
	DisplayName string
}

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (User, error)
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
