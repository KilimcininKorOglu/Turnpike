package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const (
	// EncryptedPrefix identifies encrypted passwords produced by this package.
	// NOTE: The .NET version uses "SXGV1:" with Windows DPAPI encryption.
	// SXGV1 and SXGV2 formats are NOT interoperable. Users migrating from the
	// .NET version to the Go version will need to re-enter their credentials.
	EncryptedPrefix = "SXGV2:"

	// appEntropy is application-specific entropy mixed into the key derivation.
	// Legacy entropy string kept for backward compatibility with existing encrypted credentials
	appEntropy = "SophosXGUserLogin2024"
)

// EncryptPassword encrypts a plaintext password using AES-256-GCM with a
// machine-derived key. Returns a base64-encoded ciphertext with EncryptedPrefix
// prepended, or an empty string if plainPassword is empty.
func EncryptPassword(plainPassword string) (string, error) {
	if plainPassword == "" {
		return "", nil
	}

	key := deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New("security: failed to create AES cipher: " + err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New("security: failed to create GCM: " + err.Error())
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", errors.New("security: failed to generate nonce: " + err.Error())
	}

	// Seal appends the ciphertext and GCM tag to nonce, so the stored blob is:
	// [ nonce (12 bytes) | ciphertext | tag (16 bytes) ]
	ciphertext := gcm.Seal(nonce, nonce, []byte(plainPassword), nil)

	return EncryptedPrefix + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPassword decrypts a password that was produced by EncryptPassword.
// Returns the plaintext password, or an error if decryption fails.
// An empty input string returns an empty string without error.
func DecryptPassword(encryptedPassword string) (string, error) {
	if encryptedPassword == "" {
		return "", nil
	}

	base64Data := encryptedPassword
	if strings.HasPrefix(encryptedPassword, EncryptedPrefix) {
		base64Data = encryptedPassword[len(EncryptedPrefix):]
	}

	data, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", errors.New("security: invalid base64 encoding: " + err.Error())
	}

	key := deriveKey()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", errors.New("security: failed to create AES cipher: " + err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New("security: failed to create GCM: " + err.Error())
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("security: ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("security: decryption failed (corrupted data or wrong key): " + err.Error())
	}

	return string(plaintext), nil
}

// IsEncrypted reports whether password appears to be an encrypted value
// produced by this package.
//
// Detection rules (matching C# PasswordEncryption.IsEncrypted):
//  1. Empty string → false
//  2. Has EncryptedPrefix → true
//  3. Valid base64 that decodes to ≥ 50 bytes → true (legacy heuristic)
//  4. Otherwise → false
func IsEncrypted(password string) bool {
	if password == "" {
		return false
	}
	if strings.HasPrefix(password, EncryptedPrefix) {
		return true
	}
	decoded, err := base64.StdEncoding.DecodeString(password)
	if err != nil {
		return false
	}
	return len(decoded) >= 50
}

// deriveKey builds a 32-byte AES-256 key from the machine ID and appEntropy.
// SHA-256 is used so the output is always exactly 32 bytes regardless of the
// length or content of the inputs.
func deriveKey() []byte {
	machineID := getMachineID()
	combined := machineID + appEntropy
	hash := sha256.Sum256([]byte(combined))
	return hash[:]
}

// getMachineID returns a best-effort machine-specific string that is stable
// across process restarts. The strategy is tried in priority order so that
// the most reliable source available on each OS is used first, with hostname
// + runtime constants as the final fallback.
func getMachineID() string {
	switch runtime.GOOS {
	case "linux":
		if id := readFile("/etc/machine-id"); id != "" {
			return strings.TrimSpace(id)
		}
		if id := readFile("/var/lib/dbus/machine-id"); id != "" {
			return strings.TrimSpace(id)
		}

	case "darwin":
		// ioreg exposes the IOPlatformUUID which is stable across reboots.
		if id := macOSHardwareUUID(); id != "" {
			return id
		}

	case "windows":
		// On Windows the most reliable stable identifier is the MachineGuid
		// from the registry. We read it via `reg query` so we stay dependency-free.
		if id := windowsMachineGUID(); id != "" {
			return id
		}
	}

	// Universal fallback: hostname + OS + architecture.
	return fallbackMachineID()
}

// readFile reads the first content of a file and returns it as a string.
// Returns an empty string if the file cannot be read.
func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// macOSHardwareUUID extracts the IOPlatformUUID from ioreg output.
// Example output line: "IOPlatformUUID" = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
func macOSHardwareUUID() string {
	out, err := exec.Command("ioreg", "-rd1", "-c", "IOPlatformExpertDevice").Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "IOPlatformUUID") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				uuid := strings.TrimSpace(parts[1])
				uuid = strings.Trim(uuid, "\"")
				if uuid != "" {
					return uuid
				}
			}
		}
	}
	return ""
}

// windowsMachineGUID reads the MachineGuid value from the Windows registry
// using the reg.exe command-line tool, which is available on all Windows versions.
func windowsMachineGUID() string {
	out, err := exec.Command(
		"reg", "query",
		`HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Cryptography`,
		"/v", "MachineGuid",
	).Output()
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "MachineGuid") {
			fields := strings.Fields(line)
			// Output format: "    MachineGuid    REG_SZ    <value>"
			if len(fields) >= 3 {
				return strings.TrimSpace(fields[len(fields)-1])
			}
		}
	}
	return ""
}

// fallbackMachineID returns a stable string derived from hostname, OS, and
// CPU architecture. This is used when OS-specific methods are unavailable.
func fallbackMachineID() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
	return hostname + "-" + runtime.GOOS + "-" + runtime.GOARCH
}
