package auth

import (
	"testing"
	"time"
)

const testSecret = "test-secret-do-not-use-in-prod"

func TestGenerateAndParseAccessToken_RoundTrip(t *testing.T) {
	token, err := GenerateAccessToken(testSecret, "user-123", "student", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	claims, err := ParseAccessToken(testSecret, token)
	if err != nil {
		t.Fatalf("ParseAccessToken returned error: %v", err)
	}

	if claims.UserID != "user-123" {
		t.Errorf("UserID = %q, want %q", claims.UserID, "user-123")
	}
	if claims.Role != "student" {
		t.Errorf("Role = %q, want %q", claims.Role, "student")
	}
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	token, err := GenerateAccessToken(testSecret, "user-123", "student", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	_, err = ParseAccessToken("a-different-secret", token)
	if err == nil {
		t.Fatal("expected error when parsing with the wrong secret, got nil")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	token, err := GenerateAccessToken(testSecret, "user-123", "student", -time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	_, err = ParseAccessToken(testSecret, token)
	if err == nil {
		t.Fatal("expected error when parsing an expired token, got nil")
	}
}

func TestParseAccessToken_Malformed(t *testing.T) {
	_, err := ParseAccessToken(testSecret, "this.is.not-a-jwt")
	if err == nil {
		t.Fatal("expected error when parsing a malformed token, got nil")
	}
}

func TestParseAccessToken_EmptyString(t *testing.T) {
	_, err := ParseAccessToken(testSecret, "")
	if err == nil {
		t.Fatal("expected error when parsing an empty token, got nil")
	}
}

func TestParseAccessToken_TamperedPayload(t *testing.T) {
	token, err := GenerateAccessToken(testSecret, "user-123", "student", time.Hour)
	if err != nil {
		t.Fatalf("GenerateAccessToken returned error: %v", err)
	}

	// Flip the last character of the signature segment to simulate a
	// tampered/corrupted token; the signature must no longer verify.
	tampered := token[:len(token)-1] + "x"

	_, err = ParseAccessToken(testSecret, tampered)
	if err == nil {
		t.Fatal("expected error when parsing a tampered token, got nil")
	}
}

func TestGenerateOpaqueToken_UniqueAndNonEmpty(t *testing.T) {
	a, err := generateOpaqueToken()
	if err != nil {
		t.Fatalf("generateOpaqueToken returned error: %v", err)
	}
	b, err := generateOpaqueToken()
	if err != nil {
		t.Fatalf("generateOpaqueToken returned error: %v", err)
	}

	if a == "" || b == "" {
		t.Fatal("expected non-empty tokens")
	}
	if a == b {
		t.Fatal("expected two generated tokens to differ")
	}
}

func TestHashToken_DeterministicAndDistinct(t *testing.T) {
	if hashToken("same-input") != hashToken("same-input") {
		t.Fatal("expected hashToken to be deterministic for the same input")
	}
	if hashToken("input-a") == hashToken("input-b") {
		t.Fatal("expected different inputs to hash differently")
	}
}
