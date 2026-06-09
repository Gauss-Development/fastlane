package jwt

import (
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const magicLinkTokenType = "magic_link"

// MagicLinkManager signs supplier RFQ-response tokens. It deliberately uses
// a key separate from access/refresh tokens so a leaked magic link can never
// be replayed against authenticated APIs, and vice versa.
type MagicLinkManager struct {
	secret []byte
}

// MagicLinkClaims scope a token to exactly one (rfq_id, supplier_id) pair.
type MagicLinkClaims struct {
	RFQID      string `json:"rfq_id"`
	SupplierID string `json:"supplier_id"`
	Type       string `json:"type"`
	jwt.RegisteredClaims
}

// NewMagicLinkManager derives a dedicated signing key when no explicit
// secret is configured, so dev environments work with just JWT_SECRET while
// production can rotate MAGIC_LINK_JWT_SECRET independently.
func NewMagicLinkManager(secret, fallbackSecret string) *MagicLinkManager {
	key := []byte(secret)
	if len(key) == 0 {
		derived := sha256.Sum256([]byte("magic-link:" + fallbackSecret))
		key = derived[:]
	}
	return &MagicLinkManager{secret: key}
}

func (m *MagicLinkManager) Generate(rfqID, supplierID string, ttl time.Duration) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(ttl)
	claims := &MagicLinkClaims{
		RFQID:      rfqID,
		SupplierID: supplierID,
		Type:       magicLinkTokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "auth-service",
			Subject:   supplierID,
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

func (m *MagicLinkManager) Validate(tokenString string) (rfqID, supplierID string, err error) {
	token, err := jwt.ParseWithClaims(tokenString, &MagicLinkClaims{}, func(token *jwt.Token) (interface{}, error) {
		if method, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		} else if method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing algorithm: %v", method.Alg())
		}
		return m.secret, nil
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to parse magic link token: %w", err)
	}
	if !token.Valid {
		return "", "", fmt.Errorf("invalid magic link token")
	}

	claims, ok := token.Claims.(*MagicLinkClaims)
	if !ok || claims.Type != magicLinkTokenType || claims.RFQID == "" || claims.SupplierID == "" {
		return "", "", fmt.Errorf("invalid magic link claims")
	}
	return claims.RFQID, claims.SupplierID, nil
}
