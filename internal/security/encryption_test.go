package security

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// EncryptPassword
// ---------------------------------------------------------------------------

func TestEncryptPassword_EmptyInput(t *testing.T) {
	result, err := EncryptPassword("")
	if err != nil {
		t.Fatalf("expected no error for empty input, got: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for empty input, got: %q", result)
	}
}

func TestEncryptPassword_ValidPassword_HasPrefix(t *testing.T) {
	result, err := EncryptPassword("mysecretpassword")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(result, EncryptedPrefix) {
		t.Errorf("expected result to have prefix %q, got: %q", EncryptedPrefix, result)
	}
}

func TestEncryptPassword_ValidPassword_NonEmpty(t *testing.T) {
	result, err := EncryptPassword("hunter2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == "" {
		t.Error("expected non-empty ciphertext for non-empty password")
	}
}

// ---------------------------------------------------------------------------
// DecryptPassword
// ---------------------------------------------------------------------------

func TestDecryptPassword_EmptyInput(t *testing.T) {
	result, err := DecryptPassword("")
	if err != nil {
		t.Fatalf("expected no error for empty input, got: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string for empty input, got: %q", result)
	}
}

func TestDecryptPassword_InvalidBase64_ReturnsError(t *testing.T) {
	_, err := DecryptPassword("this-is-not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64 input, got nil")
	}
}

func TestDecryptPassword_CorruptedData_ReturnsError(t *testing.T) {
	// Valid base64 but not a valid AES-GCM ciphertext encrypted with our key.
	corrupt := EncryptedPrefix + "dGhpcyBpcyBjb3JydXB0ZWQgZGF0YSB0aGF0IGlzIGxvbmcgZW5vdWdoIHRvIHBhc3MgbGVuZ3RoIGNoZWNrcw=="
	_, err := DecryptPassword(corrupt)
	if err == nil {
		t.Error("expected error for corrupted ciphertext, got nil")
	}
}

// ---------------------------------------------------------------------------
// Round-trip
// ---------------------------------------------------------------------------

func TestRoundTrip_CorrectlyDecryptsEncryptedPassword(t *testing.T) {
	original := "P@ssw0rd!#2024"

	encrypted, err := EncryptPassword(original)
	if err != nil {
		t.Fatalf("EncryptPassword error: %v", err)
	}

	decrypted, err := DecryptPassword(encrypted)
	if err != nil {
		t.Fatalf("DecryptPassword error: %v", err)
	}

	if decrypted != original {
		t.Errorf("round-trip mismatch: got %q, want %q", decrypted, original)
	}
}

func TestRoundTrip_UnicodePassword(t *testing.T) {
	original := "şifrê_ünïcödé_123"

	encrypted, err := EncryptPassword(original)
	if err != nil {
		t.Fatalf("EncryptPassword error: %v", err)
	}

	decrypted, err := DecryptPassword(encrypted)
	if err != nil {
		t.Fatalf("DecryptPassword error: %v", err)
	}

	if decrypted != original {
		t.Errorf("unicode round-trip mismatch: got %q, want %q", decrypted, original)
	}
}

func TestRoundTrip_WithoutPrefix_StillDecrypts(t *testing.T) {
	// DecryptPassword must handle input that lacks the prefix (legacy path).
	original := "legacyPassword"

	encrypted, err := EncryptPassword(original)
	if err != nil {
		t.Fatalf("EncryptPassword error: %v", err)
	}

	// Strip prefix to simulate legacy stored value.
	withoutPrefix := strings.TrimPrefix(encrypted, EncryptedPrefix)

	decrypted, err := DecryptPassword(withoutPrefix)
	if err != nil {
		t.Fatalf("DecryptPassword (no prefix) error: %v", err)
	}

	if decrypted != original {
		t.Errorf("round-trip without prefix mismatch: got %q, want %q", decrypted, original)
	}
}

// ---------------------------------------------------------------------------
// IsEncrypted
// ---------------------------------------------------------------------------

func TestIsEncrypted_EmptyString_ReturnsFalse(t *testing.T) {
	if IsEncrypted("") {
		t.Error("expected false for empty string")
	}
}

func TestIsEncrypted_PrefixedString_ReturnsTrue(t *testing.T) {
	encrypted, err := EncryptPassword("somepassword")
	if err != nil {
		t.Fatalf("EncryptPassword error: %v", err)
	}
	if !IsEncrypted(encrypted) {
		t.Errorf("expected true for prefixed encrypted string, got false; value: %q", encrypted)
	}
}

func TestIsEncrypted_PlaintextPassword_ReturnsFalse(t *testing.T) {
	if IsEncrypted("plainTextPassword") {
		t.Error("expected false for plaintext password")
	}
}

func TestIsEncrypted_ShortBase64_ReturnsFalse(t *testing.T) {
	// Base64 of fewer than 50 bytes must return false (heuristic check).
	// "short" encodes to 5 bytes, well below the 50-byte threshold.
	shortBase64 := "c2hvcnQ=" // base64("short") = 5 bytes
	if IsEncrypted(shortBase64) {
		t.Errorf("expected false for short base64 string %q", shortBase64)
	}
}

func TestIsEncrypted_LongBase64_WithoutPrefix_ReturnsTrue(t *testing.T) {
	// A valid base64 string that decodes to ≥ 50 bytes should be treated as
	// a legacy encrypted value.
	// AES-GCM overhead: 12 (nonce) + len(plaintext) + 16 (tag) must be ≥ 50,
	// so the plaintext must be at least 22 bytes. We use a password long enough
	// to guarantee the decoded blob clears the 50-byte threshold.
	longPassword := "this-password-is-long-enough-for-the-test"
	encrypted, err := EncryptPassword(longPassword)
	if err != nil {
		t.Fatalf("EncryptPassword error: %v", err)
	}
	withoutPrefix := strings.TrimPrefix(encrypted, EncryptedPrefix)
	if !IsEncrypted(withoutPrefix) {
		t.Error("expected true for long base64 string without prefix")
	}
}

func TestIsEncrypted_InvalidBase64_ReturnsFalse(t *testing.T) {
	if IsEncrypted("not-valid-base64!!!") {
		t.Error("expected false for invalid base64 string")
	}
}

// ---------------------------------------------------------------------------
// Nonce randomness: same plaintext must not produce the same ciphertext twice
// ---------------------------------------------------------------------------

func TestEncryptPassword_ProducesDifferentCiphertextsForSameInput(t *testing.T) {
	password := "samepassword"

	first, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("first EncryptPassword error: %v", err)
	}
	second, err := EncryptPassword(password)
	if err != nil {
		t.Fatalf("second EncryptPassword error: %v", err)
	}

	if first == second {
		t.Error("expected different ciphertexts for the same plaintext (nonce must be random)")
	}
}

// ---------------------------------------------------------------------------
// Key derivation stability
// ---------------------------------------------------------------------------

func TestDeriveKey_ConsistentAcrossCallsOnSameMachine(t *testing.T) {
	key1 := deriveKey()
	key2 := deriveKey()

	if len(key1) != 32 {
		t.Errorf("expected 32-byte key, got %d bytes", len(key1))
	}
	if string(key1) != string(key2) {
		t.Error("deriveKey must return the same key on consecutive calls on the same machine")
	}
}

func TestDeriveKey_Returns32Bytes(t *testing.T) {
	key := deriveKey()
	if len(key) != 32 {
		t.Errorf("AES-256 requires a 32-byte key, got %d bytes", len(key))
	}
}
