package crypto

import "testing"

func TestHashPasswordAndVerify(t *testing.T) {
	hash, err := HashPassword("s3cret-pass")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if hash == "s3cret-pass" {
		t.Fatal("expected password hash to differ from the plaintext password")
	}

	if err := VerifyPassword("s3cret-pass", hash); err != nil {
		t.Fatalf("verify password: %v", err)
	}
}

func TestVerifyPasswordRejectsInvalidPassword(t *testing.T) {
	hash, err := HashPassword("correct-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	if err := VerifyPassword("wrong-password", hash); err == nil {
		t.Fatal("expected invalid password to be rejected")
	}
}
