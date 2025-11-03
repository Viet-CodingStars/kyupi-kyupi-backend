package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash to not be empty")
	}
	if hash == password {
		t.Fatal("expected hash to be different from password")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// Test correct password
	err = CheckPassword(hash, password)
	if err != nil {
		t.Fatalf("expected no error for correct password, got %v", err)
	}

	// Test incorrect password
	err = CheckPassword(hash, "wrongpassword")
	if err == nil {
		t.Fatal("expected error for incorrect password")
	}
}

func TestGenerateToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"

	token, err := GenerateToken(userID, email, secret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token to not be empty")
	}
}

func TestValidateToken(t *testing.T) {
	userID := uuid.New()
	email := "test@example.com"
	secret := "test-secret"

	token, err := GenerateToken(userID, email, secret)
	if err != nil {
		t.Fatalf("expected no error generating token, got %v", err)
	}

	// Test valid token
	claims, err := ValidateToken(token, secret)
	if err != nil {
		t.Fatalf("expected no error validating token, got %v", err)
	}
	if claims.UserID != userID {
		t.Fatalf("expected userID %v, got %v", userID, claims.UserID)
	}
	if claims.Email != email {
		t.Fatalf("expected email %s, got %s", email, claims.Email)
	}

	// Test invalid token
	_, err = ValidateToken("invalid-token", secret)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}

	// Test wrong secret
	_, err = ValidateToken(token, "wrong-secret")
	if err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
