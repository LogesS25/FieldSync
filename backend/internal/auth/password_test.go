package auth

import "testing"

func TestHashPassword_ProducesVerifiableHash(t *testing.T) {
	hash, err := HashPassword("correct-horse-battery-staple")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "correct-horse-battery-staple" {
		t.Fatal("hash must not equal the plaintext password")
	}
}

func TestCheckPassword_CorrectPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if !CheckPassword("s3cret-password", hash) {
		t.Fatal("expected correct password to verify")
	}
}

func TestCheckPassword_WrongPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if CheckPassword("wrong-password", hash) {
		t.Fatal("expected wrong password to fail verification")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	hash, err := HashPassword("s3cret-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if CheckPassword("", hash) {
		t.Fatal("expected empty password to fail verification")
	}
}

func TestCheckPassword_MalformedHash(t *testing.T) {
	if CheckPassword("anything", "not-a-real-bcrypt-hash") {
		t.Fatal("expected malformed hash to fail verification, not panic or pass")
	}
}

func TestHashPassword_SameInputProducesDifferentHashes(t *testing.T) {
	// bcrypt salts each hash, so two hashes of the same password must not
	// be equal — otherwise a leaked hash table would reveal duplicate
	// passwords across users.
	hash1, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	hash2, err := HashPassword("same-password")
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash1 == hash2 {
		t.Fatal("expected different salts to produce different hashes")
	}
}
