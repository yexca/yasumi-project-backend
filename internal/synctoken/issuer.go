package synctoken

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/yasumi/yasumi-project-backend/internal/config"
)

type Token struct {
	Value     string
	ExpiresAt time.Time
	UserID    string
}

type Issuer interface {
	Issue(ctx context.Context, userID, deviceID, clientVersion string) (Token, error)
}

type HMACIssuer struct {
	secret []byte
	ttl    time.Duration
	now    func() time.Time
}

func NewHMACIssuer(cfg config.SyncTokenConfig) *HMACIssuer {
	return &HMACIssuer{
		secret: []byte(cfg.Secret),
		ttl:    cfg.TTL,
		now:    time.Now,
	}
}

func (i *HMACIssuer) Issue(_ context.Context, userID, deviceID, clientVersion string) (Token, error) {
	issuedAt := i.now().UTC()
	expiresAt := issuedAt.Add(i.ttl)
	payload := tokenPayload{
		UserID:        userID,
		DeviceID:      deviceID,
		ClientVersion: clientVersion,
		IssuedAt:      issuedAt.Format(time.RFC3339),
		ExpiresAt:     expiresAt.Format(time.RFC3339),
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Token{}, fmt.Errorf("marshal sync token payload: %w", err)
	}

	mac := hmac.New(sha256.New, i.secret)
	if _, err := mac.Write(payloadBytes); err != nil {
		return Token{}, fmt.Errorf("sign sync token payload: %w", err)
	}

	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	encodedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return Token{
		Value:     encodedPayload + "." + encodedSignature,
		ExpiresAt: expiresAt,
		UserID:    userID,
	}, nil
}

type tokenPayload struct {
	UserID        string `json:"user_id"`
	DeviceID      string `json:"device_id,omitempty"`
	ClientVersion string `json:"client_version,omitempty"`
	IssuedAt      string `json:"iat"`
	ExpiresAt     string `json:"exp"`
}
