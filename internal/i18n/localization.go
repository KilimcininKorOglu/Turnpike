package i18n

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"
)

// LangEnglish is the language code for English.
const LangEnglish = "en"

// LangTurkish is the language code for Turkish.
const LangTurkish = "tr"

// PreferenceStore is an interface for persisting language preferences.
// Compatible with fyne.Preferences.
type PreferenceStore interface {
	String(key string) string
	SetString(key string, value string)
}

// Manager handles application localization.
// It is a direct port of LocalizationManager.cs from the C# WPF application.
// Language changes are broadcast to all registered listeners via callbacks.
type Manager struct {
	mu              sync.RWMutex
	currentLanguage string
	listeners       []func(string)
	translations    map[string]map[string]string // key -> language -> translation
	prefs           PreferenceStore
}

// NewManager creates a new localization manager with the given default language.
// If language is empty or unsupported, English is used as the fallback.
func NewManager(language string) *Manager {
	m := &Manager{}
	m.initTranslations()
	if language == LangTurkish {
		m.currentLanguage = LangTurkish
	} else {
		m.currentLanguage = LangEnglish
	}
	return m
}

// GetString returns the localized string for the given key in the current language.
// If the key does not exist in the current language, the English fallback is attempted.
// If no translation is found at all, the key itself is returned.
func (m *Manager) GetString(key string) string {
	m.mu.RLock()
	lang := m.currentLanguage
	m.mu.RUnlock()

	if langMap, ok := m.translations[key]; ok {
		if val, ok := langMap[lang]; ok {
			return val
		}
		// Fall back to English if the key is missing in the current language.
		if val, ok := langMap[LangEnglish]; ok {
			return val
		}
	}
	// Last resort: return the key itself so the caller always gets a non-empty string.
	return key
}

// GetStringf returns the localized string for the given key with fmt.Sprintf formatting.
// It is equivalent to fmt.Sprintf(m.GetString(key), args...).
func (m *Manager) GetStringf(key string, args ...interface{}) string {
	return fmt.Sprintf(m.GetString(key), args...)
}

// SetLanguage changes the current language and notifies all registered listeners.
// Only LangEnglish ("en") and LangTurkish ("tr") are supported; other values are ignored.
func (m *Manager) SetLanguage(language string) {
	if language != LangEnglish && language != LangTurkish {
		return
	}

	m.mu.Lock()
	if m.currentLanguage == language {
		m.mu.Unlock()
		return
	}
	m.currentLanguage = language
	prefs := m.prefs
	// Copy listeners slice to safely call outside the lock.
	listeners := make([]func(string), len(m.listeners))
	copy(listeners, m.listeners)
	m.mu.Unlock()

	// Persist language preference
	if prefs != nil {
		prefs.SetString("Language", language)
	}

	for _, cb := range listeners {
		cb(language)
	}
}

// CurrentLanguage returns the active language code ("en" or "tr").
func (m *Manager) CurrentLanguage() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentLanguage
}

// IsCurrentLanguage reports whether the given language code is the active language.
func (m *Manager) IsCurrentLanguage(language string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.currentLanguage == language
}

// SetPreferences sets the preference store for language persistence.
// After setting, call LoadFromPreferences to apply the saved language.
func (m *Manager) SetPreferences(prefs PreferenceStore) {
	m.mu.Lock()
	m.prefs = prefs
	m.mu.Unlock()
}

// LoadFromPreferences loads the saved language from the preference store.
// If no saved language exists, does nothing (keeps current language).
func (m *Manager) LoadFromPreferences() {
	m.mu.RLock()
	prefs := m.prefs
	m.mu.RUnlock()

	if prefs == nil {
		return
	}

	saved := prefs.String("Language")
	if saved == LangEnglish || saved == LangTurkish {
		m.SetLanguage(saved)
	}
}

// DetectSystemLanguage returns the system's UI language code.
// Returns "tr" if the system locale is Turkish, otherwise "en".
func DetectSystemLanguage() string {
	var locale string

	if runtime.GOOS == "windows" {
		// On Windows, check common environment variables
		locale = os.Getenv("LANG")
		if locale == "" {
			locale = os.Getenv("LANGUAGE")
		}
	} else {
		// On Unix-like systems, check LC_ALL, LC_MESSAGES, LANG
		locale = os.Getenv("LC_ALL")
		if locale == "" {
			locale = os.Getenv("LC_MESSAGES")
		}
		if locale == "" {
			locale = os.Getenv("LANG")
		}
	}

	locale = strings.ToLower(locale)
	if strings.HasPrefix(locale, "tr") {
		return LangTurkish
	}
	return LangEnglish
}

// OnLanguageChanged registers a callback that is invoked whenever the language changes.
// The callback receives the new language code as its argument.
// Callbacks are called synchronously in the order they were registered.
func (m *Manager) OnLanguageChanged(callback func(string)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listeners = append(m.listeners, callback)
}

// initTranslations populates the full translation table.
// All 56+ strings are ported directly from LocalizationManager.GetFallbackString in C#.
func (m *Manager) initTranslations() {
	t := make(map[string]map[string]string)

	add := func(key, en, tr string) {
		t[key] = map[string]string{
			LangEnglish: en,
			LangTurkish: tr,
		}
	}

	// -------------------------------------------------------------------------
	// Window / Button / Status
	// -------------------------------------------------------------------------
	add("WindowTitle",
		"Turnpike",
		"Turnpike")

	add("ButtonLogin",
		"Log In",
		"Giriş Yap")

	add("ButtonLogout",
		"Logout",
		"Çıkış")

	add("ButtonDisconnect",
		"Disconnect",
		"Bağlantıyı Kes")

	add("ButtonConnecting",
		"Connecting...",
		"Bağlanıyor...")

	add("ButtonReconnect",
		"Reconnect",
		"Yeniden Bağlan")

	add("StatusLoggedOut",
		"Successfully logged out",
		"Başarıyla çıkış yapıldı")

	add("StatusLoggedIn",
		"Successfully logged in",
		"Başarıyla giriş yapıldı")

	add("StatusConnecting",
		"Connecting...",
		"Bağlanıyor...")

	add("StatusSessionActive",
		"Session active",
		"Oturum aktif")

	add("StatusSessionExpired",
		"Session Expired",
		"Oturum süresi doldu")

	add("StatusConnectedTo",
		"Connected to %s",
		"Bağlı: %s")

	add("StatusLoggedInAsUser",
		"Logged in as: %s",
		"Kullanıcı: %s")

	add("StatusSessionTime",
		"Session time: %s",
		"Oturum süresi: %s")

	add("StatusConnectingAuto",
		"Reconnecting automatically...",
		"Otomatik olarak yeniden bağlanıyor...")

	add("StatusReconnected",
		"Automatically reconnected",
		"Otomatik olarak yeniden bağlandı")

	add("StatusRetryingConnection",
		"Connection failed, retrying... (%s)",
		"Bağlantı başarısız, yeniden deniyor... (%s)")

	// -------------------------------------------------------------------------
	// Errors
	// -------------------------------------------------------------------------
	add("ErrorEnterCredentials",
		"Enter Credentials or No Profiles found.",
		"Kimlik bilgilerini girin veya profil bulunamadı.")

	add("ErrorEnterUserID",
		"Enter UserID",
		"Kullanıcı adını girin")

	add("ErrorEnterPassword",
		"Enter Password",
		"Şifre girin")

	add("ErrorConnectionFailed",
		"Unable to connect to server",
		"Sunucuya bağlanılamıyor")

	add("ErrorConnectionTitle",
		"Connection Error",
		"Bağlantı Hatası")

	add("ErrorLogout",
		"Logout error: %s",
		"Çıkış hatası: %s")

	add("ErrorLogoutFailed",
		"Logout failed - you may still be connected",
		"Çıkış başarısız - hala bağlı olabilirsiniz")

	add("ErrorSessionExpiredMessage",
		"Session expired - please log in again",
		"Oturum süresi doldu - lütfen tekrar giriş yapın")

	add("ErrorWrongPassword",
		"Wrong password",
		"Yanlış şifre")

	add("ErrorMaxLimit",
		"Maximum connection limit reached",
		"Maksimum bağlantı limitine ulaşıldı")

	add("ErrorDataLimit",
		"Data limit reached",
		"Veri limitine ulaşıldı")

	add("ErrorMaxReconnectAttempts",
		"Auto-reconnection failed after 3 attempts. Please login manually.",
		"Otomatik yeniden bağlanma 3 denemeden sonra başarısız. Lütfen manuel giriş yapın.")

	add("ErrorNoSavedCredentials",
		"Session expired. No saved credentials for auto-reconnect.",
		"Oturum süresi doldu. Otomatik bağlanma için kayıtlı kimlik bilgisi yok.")

	add("ErrorInvalidPort",
		"Invalid port number. Please enter a value between 1 and 65535.",
		"Geçersiz port numarası. Lütfen 1-65535 arasında bir değer girin.")

	add("ErrorAutoStartEnable",
		"Failed to enable auto-start",
		"Otomatik başlatma etkinleştirilemedi")

	add("ErrorAutoStartDisable",
		"Failed to disable auto-start",
		"Otomatik başlatma devre dışı bırakılamadı")

	// -------------------------------------------------------------------------
	// Dialogs
	// -------------------------------------------------------------------------
	add("DialogLogoutConfirm",
		"Are you sure you want to logout from Turnpike?",
		"Turnpike'den çıkış yapmak istediğinizden emin misiniz?")

	add("DialogLogoutTitle",
		"Confirm Logout",
		"Çıkışı Onayla")

	add("DialogRetryConnection",
		"Retry connection?",
		"Bağlantı yeniden denensin mi?")

	// -------------------------------------------------------------------------
	// Menu
	// -------------------------------------------------------------------------
	add("MenuConnection",
		"Connection",
		"Bağlantı")

	add("MenuLogout",
		"Logout",
		"Çıkış")

	add("MenuExit",
		"Exit",
		"Kapat")

	add("MenuLanguage",
		"Language",
		"Dil")

	add("MenuItemEnglish",
		"English",
		"English")

	add("MenuItemTurkish",
		"Türkçe",
		"Türkçe")

	// -------------------------------------------------------------------------
	// Labels
	// -------------------------------------------------------------------------
	add("LabelServerConfiguration",
		"Server Configuration",
		"Sunucu Yapılandırması")

	add("LabelAutoStartWindows",
		"Start automatically with Windows (minimized)",
		"Windows ile otomatik başlat (simge durumunda)")

	add("LabelUserCredentials",
		"User Credentials",
		"Kullanıcı Bilgileri")

	add("LabelServerIP",
		"Server IP:",
		"Sunucu IP:")

	add("LabelCaptivePortal",
		"Captive Portal Port:",
		"Captive Portal Port:")

	add("LabelUsername",
		"Username:",
		"Kullanıcı Adı:")

	add("LabelPassword",
		"Password:",
		"Şifre:")

	add("LabelPortStandard",
		"(Standard: 8090)",
		"(Standart: 8090)")

	// -------------------------------------------------------------------------
	// Checkboxes
	// -------------------------------------------------------------------------
	add("CheckboxUseHTTPS",
		"Use HTTPS (SSL/TLS)",
		"HTTPS kullan (SSL/TLS)")

	add("CheckboxSkipSSL",
		"Skip SSL certificate validation (for self-signed certificates)",
		"SSL sertifika doğrulamasını atla (kendinden imzalı sertifikalar için)")

	add("CheckboxRememberCredentials",
		"Remember credentials and settings",
		"Kimlik bilgileri ve ayarları hatırla")

	// -------------------------------------------------------------------------
	// Info / Success
	// -------------------------------------------------------------------------
	add("InfoCredentialsUpgraded",
		"Security enhanced: Credentials encrypted",
		"Güvenlik iyileştirildi: Kimlik bilgileri şifrelendi")

	add("InfoPortEditable",
		"Port can now be edited. Press Enter to save or Escape to cancel.",
		"Port düzenlenebilir. Kaydetmek için Enter, iptal için Escape tuşuna basın.")

	add("SuccessPortSaved",
		"Port saved: %s",
		"Port kaydedildi: %s")

	add("SuccessAutoStartEnabled",
		"Auto-start with Windows enabled",
		"Windows başlangıcında otomatik çalışma etkinleştirildi")

	add("InfoAutoStartDisabled",
		"Auto-start with Windows disabled",
		"Windows başlangıcında otomatik çalışma devre dışı bırakıldı")

	add("InfoPortEditCancelled",
		"Edit cancelled - restored to default (8090)",
		"Düzenleme iptal edildi - varsayılan değer geri yüklendi (8090)")

	add("InfoMinimizedToTray",
		"Minimized to system tray",
		"Sistem tepsisine küçültüldü")

	add("InfoAutoReconnecting",
		"Session expired - reconnecting automatically...",
		"Oturum süresi doldu - otomatik yeniden bağlanıyor...")

	add("SuccessAutoReconnected",
		"Automatically reconnected successfully!",
		"Otomatik olarak başarıyla yeniden bağlandı!")

	// -------------------------------------------------------------------------
	// Context Menu
	// -------------------------------------------------------------------------
	add("ContextMenuShow",
		"Show",
		"Göster")

	add("ContextMenuExit",
		"Exit",
		"Çıkış")

	// SSL two-line labels (for potential future use matching .NET layout)
	add("CheckboxSkipSSLLine1",
		"Skip SSL certificate validation",
		"SSL sertifika doğrulamasını atla")
	add("CheckboxSkipSSLLine2",
		"(for self-signed certificates)",
		"(kendinden imzalı sertifikalar için)")
	add("LabelSkipSSL",
		"Skip SSL certificate validation",
		"SSL sertifika doğrulamasını atla")
	add("LabelSkipSSLNote",
		"(for self-signed certificates)",
		"(kendinden imzalı sertifikalar için)")

	// Auto-login
	add("LabelAutoLogin",
		"Auto-login on startup",
		"Başlangıçta otomatik giriş yap")

	m.translations = t
}
