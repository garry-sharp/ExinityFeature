package services

import (
	"crypto/rand"
	"testing"
)

func TestEncryptAndDecrypt(t *testing.T) {
	plaintext := "This is a test message."
	ciphertext, err := Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Decrypt the ciphertext
	decrypted, err := Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("decryption failed: %v", err)
	}

	// Verify the decrypted plaintext matches the original plaintext
	if decrypted != plaintext {
		t.Errorf("decrypted text does not match original: got %q, want %q", decrypted, plaintext)
	}
}

func TestDecryptWithWrongKey(t *testing.T) {
	// Generate two different 32-byte keys
	key1 := make([]byte, 32)
	key2 := make([]byte, 32)
	if _, err := rand.Read(key1); err != nil {
		t.Fatalf("failed to generate key1: %v", err)
	}
	if _, err := rand.Read(key2); err != nil {
		t.Fatalf("failed to generate key2: %v", err)
	}

	// Define plaintext
	plaintext := "This is a test message."

	// Encrypt with the first key
	key = key1
	ciphertext, err := Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatalf("encryption failed: %v", err)
	}

	// Try decrypting with the second (incorrect) key
	key = key2
	_, err = Decrypt(ciphertext)
	if err == nil {
		t.Errorf("expected decryption to fail with wrong key, but it succeeded")
	}
}
