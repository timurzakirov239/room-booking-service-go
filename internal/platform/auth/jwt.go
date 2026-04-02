package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrInvalidToken     = errors.New("auth: invalid token")
	ErrMissingToken     = errors.New("auth: missing token")
	ErrInvalidSignature = errors.New("auth: invalid signature")
)

type Claims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp,omitempty"`
	Iat    int64  `json:"iat,omitempty"`
	Iss    string `json:"iss,omitempty"`
}

type Signer struct {
	Secret   string
	Issuer   string
	Lifetime time.Duration
	Now      func() time.Time
}

func (s Signer) Sign(claims Claims) (string, error) {
	if s.Secret == "" {
		return "", ErrInvalidToken
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}

	claims.Iat = now.Unix()
	if s.Lifetime > 0 {
		claims.Exp = now.Add(s.Lifetime).Unix()
	}
	if s.Issuer != "" {
		claims.Iss = s.Issuer
	}

	headerBytes, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := encodeSegment(headerBytes)
	payload := encodeSegment(payloadBytes)
	signingInput := header + "." + payload
	signature := signHS256(signingInput, s.Secret)

	return signingInput + "." + signature, nil
}

func (s Signer) Verify(token string) (Claims, error) {
	if s.Secret == "" {
		return Claims{}, ErrInvalidToken
	}
	if strings.TrimSpace(token) == "" {
		return Claims{}, ErrMissingToken
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	expectedSignature := signHS256(signingInput, s.Secret)
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[2])) {
		return Claims{}, ErrInvalidSignature
	}

	payloadBytes, err := decodeSegment(parts[1])
	if err != nil {
		return Claims{}, fmt.Errorf("decode payload: %w", err)
	}

	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return Claims{}, fmt.Errorf("unmarshal claims: %w", err)
	}

	now := time.Now().UTC()
	if s.Now != nil {
		now = s.Now().UTC()
	}
	if claims.Exp > 0 && now.Unix() > claims.Exp {
		return Claims{}, ErrInvalidToken
	}

	return claims, nil
}

func encodeSegment(value []byte) string {
	return base64.RawURLEncoding.EncodeToString(value)
}

func decodeSegment(value string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(value)
}

func signHS256(input string, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(input))
	return encodeSegment(mac.Sum(nil))
}
