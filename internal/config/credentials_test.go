package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func createTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "credentials_test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestNewCredentialManager_Defaults(t *testing.T) {
	dir := createTempDir(t)
	cm := NewCredentialManager(dir)

	creds := cm.GetCredentials()
	if creds.ServerIP != DefaultServerIP {
		t.Errorf("expected default ServerIP %s, got %s", DefaultServerIP, creds.ServerIP)
	}
	if creds.CaptivePortalPort != DefaultCaptivePortalPort {
		t.Errorf("expected default port %d, got %d", DefaultCaptivePortalPort, creds.CaptivePortalPort)
	}
	if creds.UseHTTPS != DefaultUseHTTPS {
		t.Errorf("expected default UseHTTPS %v, got %v", DefaultUseHTTPS, creds.UseHTTPS)
	}
	if creds.SkipSSLValidation != DefaultSkipSSLValidation {
		t.Errorf("expected default SkipSSLValidation %v, got %v", DefaultSkipSSLValidation, creds.SkipSSLValidation)
	}
	if cm.RememberCredentials() {
		t.Error("expected RememberCredentials to be false by default")
	}
}

func TestCredentialManager_SetAndGetCredentials(t *testing.T) {
	dir := createTempDir(t)
	cm := NewCredentialManager(dir)

	creds := Credentials{
		Username:          "testuser",
		Password:          "testpass",
		ServerIP:          "10.0.0.1",
		CaptivePortalPort: 9090,
		UseHTTPS:          false,
		SkipSSLValidation: true,
	}
	cm.SetCredentials(creds)

	got := cm.GetCredentials()
	if got.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", got.Username)
	}
	if got.Password != "testpass" {
		t.Errorf("expected password testpass, got %s", got.Password)
	}
	if got.ServerIP != "10.0.0.1" {
		t.Errorf("expected ServerIP 10.0.0.1, got %s", got.ServerIP)
	}
	if got.CaptivePortalPort != 9090 {
		t.Errorf("expected port 9090, got %d", got.CaptivePortalPort)
	}
}

func TestCredentialManager_IndividualSetters(t *testing.T) {
	dir := createTempDir(t)
	cm := NewCredentialManager(dir)

	cm.SetUsername("user1")
	cm.SetPassword("pass1")
	cm.SetServerIP("192.168.1.1")
	cm.SetCaptivePortalPort(8443)
	cm.SetUseHTTPS(true)
	cm.SetSkipSSLValidation(true)

	creds := cm.GetCredentials()
	if creds.Username != "user1" {
		t.Errorf("expected username user1, got %s", creds.Username)
	}
	if creds.Password != "pass1" {
		t.Errorf("expected password pass1, got %s", creds.Password)
	}
	if creds.ServerIP != "192.168.1.1" {
		t.Errorf("expected ServerIP 192.168.1.1, got %s", creds.ServerIP)
	}
	if creds.CaptivePortalPort != 8443 {
		t.Errorf("expected port 8443, got %d", creds.CaptivePortalPort)
	}
	if !creds.UseHTTPS {
		t.Error("expected UseHTTPS true")
	}
	if !creds.SkipSSLValidation {
		t.Error("expected SkipSSLValidation true")
	}
}

func TestCredentialManager_SaveAndLoad(t *testing.T) {
	dir := createTempDir(t)

	// Create and save
	cm1 := NewCredentialManager(dir)
	cm1.SetRememberCredentials(true)
	cm1.SetUsername("saveduser")
	cm1.SetPassword("savedpass")
	cm1.SetServerIP("10.0.0.5")
	cm1.SetCaptivePortalPort(8443)
	cm1.SetUseHTTPS(false)
	cm1.SetSkipSSLValidation(true)

	// Load in a new manager
	cm2 := NewCredentialManager(dir)

	creds := cm2.GetCredentials()
	if creds.Username != "saveduser" {
		t.Errorf("expected loaded username saveduser, got %s", creds.Username)
	}
	if creds.Password != "savedpass" {
		t.Errorf("expected loaded password savedpass, got %s", creds.Password)
	}
	if creds.ServerIP != "10.0.0.5" {
		t.Errorf("expected loaded ServerIP 10.0.0.5, got %s", creds.ServerIP)
	}
	if creds.CaptivePortalPort != 8443 {
		t.Errorf("expected loaded port 8443, got %d", creds.CaptivePortalPort)
	}
	if creds.UseHTTPS {
		t.Error("expected loaded UseHTTPS false")
	}
	if !creds.SkipSSLValidation {
		t.Error("expected loaded SkipSSLValidation true")
	}
	if !cm2.RememberCredentials() {
		t.Error("expected RememberCredentials to be true after loading")
	}
}

func TestCredentialManager_PasswordEncryption(t *testing.T) {
	dir := createTempDir(t)

	cm := NewCredentialManager(dir)
	cm.SetRememberCredentials(true)
	cm.SetPassword("mysecretpassword")

	// Read raw file to verify password is encrypted
	filePath := filepath.Join(dir, credentialsFileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("failed to read credentials file: %v", err)
	}

	var stored Credentials
	if err := json.Unmarshal(data, &stored); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if stored.Password == "mysecretpassword" {
		t.Error("password stored in plaintext, expected encrypted")
	}
	if stored.Password == "" {
		t.Error("encrypted password is empty")
	}
}

func TestCredentialManager_ClearSavedCredentials(t *testing.T) {
	dir := createTempDir(t)

	cm := NewCredentialManager(dir)
	cm.SetRememberCredentials(true)
	cm.SetUsername("todelete")

	// Verify file exists
	filePath := filepath.Join(dir, credentialsFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Error("expected credentials file to exist after save")
	}

	// Clear
	cm.SetRememberCredentials(false)

	// Verify file is deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected credentials file to be deleted after clear")
	}
}

func TestCredentialManager_ClearNonExistentFile(t *testing.T) {
	dir := createTempDir(t)
	cm := NewCredentialManager(dir)

	// Should not error when clearing non-existent file
	err := cm.ClearSavedCredentials()
	if err != nil {
		t.Errorf("unexpected error clearing non-existent file: %v", err)
	}
}

func TestCredentialManager_LoadCorruptedFile(t *testing.T) {
	dir := createTempDir(t)

	// Write corrupted JSON
	filePath := filepath.Join(dir, credentialsFileName)
	os.WriteFile(filePath, []byte("not valid json{{{"), 0600)

	// Should not panic, should use defaults
	cm := NewCredentialManager(dir)
	creds := cm.GetCredentials()
	if creds.ServerIP != DefaultServerIP {
		t.Errorf("expected default ServerIP after corrupted load, got %s", creds.ServerIP)
	}
}

func TestCredentialManager_LoadPlaintextPassword(t *testing.T) {
	dir := createTempDir(t)

	// Write file with plaintext password (legacy format)
	stored := Credentials{
		Username:          "legacyuser",
		Password:          "plaintextpass",
		ServerIP:          "10.0.0.1",
		CaptivePortalPort: 8090,
		UseHTTPS:          true,
		SkipSSLValidation: false,
	}
	data, _ := json.MarshalIndent(stored, "", "  ")
	filePath := filepath.Join(dir, credentialsFileName)
	os.WriteFile(filePath, data, 0600)

	cm := NewCredentialManager(dir)
	creds := cm.GetCredentials()

	if creds.Username != "legacyuser" {
		t.Errorf("expected username legacyuser, got %s", creds.Username)
	}
	if creds.Password != "plaintextpass" {
		t.Errorf("expected password plaintextpass, got %s", creds.Password)
	}
	if !cm.NeedsPasswordMigration() {
		t.Error("expected NeedsPasswordMigration to be true for plaintext password")
	}
}

func TestCredentialManager_MigratePasswordEncryption(t *testing.T) {
	dir := createTempDir(t)

	// Write legacy plaintext
	stored := Credentials{
		Username:          "migrateuser",
		Password:          "migratepass",
		ServerIP:          DefaultServerIP,
		CaptivePortalPort: DefaultCaptivePortalPort,
		UseHTTPS:          DefaultUseHTTPS,
		SkipSSLValidation: DefaultSkipSSLValidation,
	}
	data, _ := json.MarshalIndent(stored, "", "  ")
	filePath := filepath.Join(dir, credentialsFileName)
	os.WriteFile(filePath, data, 0600)

	cm := NewCredentialManager(dir)
	if !cm.NeedsPasswordMigration() {
		t.Error("expected migration needed")
	}

	cm.MigratePasswordEncryption()

	if cm.NeedsPasswordMigration() {
		t.Error("expected migration no longer needed after migrate")
	}

	// Verify file now has encrypted password
	data2, _ := os.ReadFile(filePath)
	var stored2 Credentials
	json.Unmarshal(data2, &stored2)
	if stored2.Password == "migratepass" {
		t.Error("expected encrypted password after migration")
	}
}

func TestCredentialManager_SaveWithoutRemember(t *testing.T) {
	dir := createTempDir(t)
	cm := NewCredentialManager(dir)
	// Don't enable remember

	err := cm.SaveCredentials()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// File should not exist
	filePath := filepath.Join(dir, credentialsFileName)
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("expected no file when remember is disabled")
	}
}

func TestCredentialManager_FilePermissions(t *testing.T) {
	dir := createTempDir(t)
	cm := NewCredentialManager(dir)
	cm.SetRememberCredentials(true)
	cm.SetPassword("secret")

	filePath := filepath.Join(dir, credentialsFileName)
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("failed to stat credentials file: %v", err)
	}

	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("expected file permissions 0600, got %o", perm)
	}
}
