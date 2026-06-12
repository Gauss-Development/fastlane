package jwt

import (
	"testing"
	"time"

	"auth-service/internal/domain/entities"
)

var fakeClaimsCopy = entities.TokenClaims{UserID: "user-1", Email: "user@example.com", Type: "access"}

func TestMagicLinkRoundTrip(t *testing.T) {
	m := NewMagicLinkManager("0123456789abcdef0123456789abcdef", "")

	token, expiresAt, err := m.Generate("RFQ-20260609-0001-SZX", "supplier-1", time.Hour)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if time.Until(expiresAt) <= 0 {
		t.Errorf("expiresAt should be in the future, got %v", expiresAt)
	}

	rfqID, supplierID, err := m.Validate(token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if rfqID != "RFQ-20260609-0001-SZX" || supplierID != "supplier-1" {
		t.Errorf("scope = %q/%q", rfqID, supplierID)
	}
}

func TestMagicLinkRejectsExpired(t *testing.T) {
	m := NewMagicLinkManager("0123456789abcdef0123456789abcdef", "")
	token, _, err := m.Generate("RFQ-X", "supplier-1", -time.Minute)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, _, err := m.Validate(token); err == nil {
		t.Error("expired token validated successfully")
	}
}

func TestMagicLinkKeySeparation(t *testing.T) {
	// A token signed with the access-token secret must not validate as a
	// magic link, even when the manager derives its key from that secret.
	sharedSecret := "0123456789abcdef0123456789abcdef"
	m := NewMagicLinkManager("", sharedSecret)

	accessManager := NewManager(sharedSecret, "auth-service")
	accessToken, err := accessManager.GenerateToken(&fakeClaimsCopy, time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken: %v", err)
	}
	if _, _, err := m.Validate(accessToken); err == nil {
		t.Error("access token validated as magic link")
	}

	// And the derived-key manager still round-trips its own tokens.
	token, _, err := m.Generate("RFQ-X", "supplier-1", time.Hour)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, _, err := m.Validate(token); err != nil {
		t.Errorf("derived-key round trip failed: %v", err)
	}
}

func TestMagicLinkRejectsTamperedToken(t *testing.T) {
	m := NewMagicLinkManager("0123456789abcdef0123456789abcdef", "")
	other := NewMagicLinkManager("ffffffffffffffffffffffffffffffff", "")

	token, _, err := other.Generate("RFQ-X", "supplier-1", time.Hour)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if _, _, err := m.Validate(token); err == nil {
		t.Error("token signed with a different key validated successfully")
	}
}
