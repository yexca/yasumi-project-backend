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
	header := tokenHeader{
		Algorithm: "HS256",
		Type:      "JWT",
		KeyID:     "local-dev-sync-key",
	}
	payload := tokenPayload{
		Subject:       userID,
		Audience:      []string{"powersync-dev", "powersync"},
		UserID:        userID,
		DeviceID:      deviceID,
		ClientVersion: clientVersion,
		IssuedAt:      issuedAt.Unix(),
		ExpiresAt:     expiresAt.Unix(),
	}

	headerBytes, err := json.Marshal(header)
	if err != nil {
		return Token{}, fmt.Errorf("marshal sync token header: %w", err)
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Token{}, fmt.Errorf("marshal sync token payload: %w", err)
	}

	encodedHeader := base64.RawURLEncoding.EncodeToString(headerBytes)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := encodedHeader + "." + encodedPayload
	mac := hmac.New(sha256.New, i.secret)
	if _, err := mac.Write([]byte(signingInput)); err != nil {
		return Token{}, fmt.Errorf("sign sync token payload: %w", err)
	}
	encodedSignature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return Token{
		Value:     signingInput + "." + encodedSignature,
		ExpiresAt: expiresAt,
		UserID:    userID,
	}, nil
}

type tokenHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
	KeyID     string `json:"kid"`
}

type tokenPayload struct {
	Subject       string   `json:"sub"`
	Audience      []string `json:"aud"`
	UserID        string   `json:"user_id"`
	DeviceID      string   `json:"device_id,omitempty"`
	ClientVersion string   `json:"client_version,omitempty"`
	IssuedAt      int64    `json:"iat"`
	ExpiresAt     int64    `json:"exp"`
}
