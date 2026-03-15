package config

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/KilimcininKorOglu/Turnpike/internal/security"
)

const credentialsFileName = "user_credentials.json"

// Credentials holds the user's saved configuration and credentials
type Credentials struct {
	Username          string `json:"Username"`
	Password          string `json:"Password"`
	ServerIP          string `json:"ServerIP"`
	CaptivePortalPort int    `json:"CaptivePortalPort"`
	UseHTTPS          bool   `json:"UseHTTPS"`
	SkipSSLValidation bool   `json:"SkipSSLValidation"`
}

// CredentialManager handles loading and saving credentials
type CredentialManager struct {
	settingsDir         string
	credentials         Credentials
	rememberCredentials bool
}

// NewCredentialManager creates a new credential manager
func NewCredentialManager(settingsDir string) *CredentialManager {
	cm := &CredentialManager{
		settingsDir: settingsDir,
		credentials: Credentials{
			ServerIP:          DefaultServerIP,
			CaptivePortalPort: DefaultCaptivePortalPort,
			UseHTTPS:          DefaultUseHTTPS,
			SkipSSLValidation: DefaultSkipSSLValidation,
		},
	}
	cm.loadCredentials()
	return cm
}

// GetCredentials returns the current credentials
func (cm *CredentialManager) GetCredentials() Credentials {
	return cm.credentials
}

// SetCredentials updates credentials and saves if remember is enabled
func (cm *CredentialManager) SetCredentials(creds Credentials) {
	cm.credentials = creds
	if cm.rememberCredentials {
		cm.SaveCredentials()
	}
}

// SetUsername updates the username
func (cm *CredentialManager) SetUsername(username string) {
	cm.credentials.Username = username
	if cm.rememberCredentials {
		cm.SaveCredentials()
	}
}

// SetPassword updates the password
func (cm *CredentialManager) SetPassword(password string) {
	cm.credentials.Password = password
	if cm.rememberCredentials {
		cm.SaveCredentials()
	}
}

// SetServerIP updates the server IP
func (cm *CredentialManager) SetServerIP(ip string) {
	cm.credentials.ServerIP = ip
	if cm.rememberCredentials {
		cm.SaveCredentials()
	}
}

// SetCaptivePortalPort updates the port
func (cm *CredentialManager) SetCaptivePortalPort(port int) {
	cm.credentials.CaptivePortalPort = port
	if cm.rememberCredentials {
		cm.SaveCredentials()
	}
}

// SetUseHTTPS updates HTTPS setting
func (cm *CredentialManager) SetUseHTTPS(useHTTPS bool) {
	cm.credentials.UseHTTPS = useHTTPS
	if cm.rememberCredentials {
		cm.SaveCredentials()
	}
}

// SetSkipSSLValidation updates SSL validation setting
func (cm *CredentialManager) SetSkipSSLValidation(skip bool) {
	cm.credentials.SkipSSLValidation = skip
	if cm.rememberCredentials {
		cm.SaveCredentials()
	}
}

// RememberCredentials returns whether credentials should be saved
func (cm *CredentialManager) RememberCredentials() bool {
	return cm.rememberCredentials
}

// SetRememberCredentials enables or disables credential saving
func (cm *CredentialManager) SetRememberCredentials(remember bool) {
	cm.rememberCredentials = remember
	if remember {
		cm.SaveCredentials()
	} else {
		cm.ClearSavedCredentials()
	}
}

// SaveCredentials persists credentials to disk with encryption
func (cm *CredentialManager) SaveCredentials() error {
	if !cm.rememberCredentials {
		return nil
	}

	// Encrypt password before saving
	encryptedPassword, err := security.EncryptPassword(cm.credentials.Password)
	if err != nil {
		return err
	}

	// Create storage struct with encrypted password
	stored := Credentials{
		Username:          cm.credentials.Username,
		Password:          encryptedPassword,
		ServerIP:          cm.credentials.ServerIP,
		CaptivePortalPort: cm.credentials.CaptivePortalPort,
		UseHTTPS:          cm.credentials.UseHTTPS,
		SkipSSLValidation: cm.credentials.SkipSSLValidation,
	}

	data, err := json.MarshalIndent(stored, "", "  ")
	if err != nil {
		return err
	}

	filePath := filepath.Join(cm.settingsDir, credentialsFileName)
	return os.WriteFile(filePath, data, 0600)
}

// ClearSavedCredentials removes the credentials file
func (cm *CredentialManager) ClearSavedCredentials() error {
	filePath := filepath.Join(cm.settingsDir, credentialsFileName)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(filePath)
}

func (cm *CredentialManager) loadCredentials() {
	filePath := filepath.Join(cm.settingsDir, credentialsFileName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return // No saved credentials
	}

	var stored Credentials
	if err := json.Unmarshal(data, &stored); err != nil {
		return
	}

	// Only decrypt passwords that carry the encryption prefix.
	// Unrecognised values (plaintext, corrupted, etc.) are discarded;
	// the user must re-enter them.
	var decryptedPassword string
	if security.IsEncrypted(stored.Password) {
		decrypted, err := security.DecryptPassword(stored.Password)
		if err == nil {
			decryptedPassword = decrypted
		}
	}

	cm.credentials = Credentials{
		Username:          stored.Username,
		Password:          decryptedPassword,
		ServerIP:          stored.ServerIP,
		CaptivePortalPort: stored.CaptivePortalPort,
		UseHTTPS:          stored.UseHTTPS,
		SkipSSLValidation: stored.SkipSSLValidation,
	}
	cm.rememberCredentials = true
}
