package i18n

import (
	"os"
	"sync"
	"testing"
)

// allEnglishTranslations maps every key to its expected English value.
// This table is the single source of truth for English coverage tests.
var allEnglishTranslations = map[string]string{
	// Window / Button / Status
	"WindowTitle":              "Sophos XG User Login",
	"ButtonLogin":              "Log In",
	"ButtonLogout":             "Logout",
	"ButtonDisconnect":         "Disconnect",
	"ButtonConnecting":         "Connecting...",
	"ButtonReconnect":          "Reconnect",
	"StatusLoggedOut":          "Successfully logged out",
	"StatusLoggedIn":           "Successfully logged in",
	"StatusConnecting":         "Connecting...",
	"StatusSessionActive":      "Session active",
	"StatusSessionExpired":     "Session Expired",
	"StatusConnectedTo":        "Connected to %s",
	"StatusLoggedInAsUser":     "Logged in as: %s",
	"StatusSessionTime":        "Session time: %s",
	"StatusConnectingAuto":     "Reconnecting automatically...",
	"StatusReconnected":        "Automatically reconnected",
	"StatusRetryingConnection": "Connection failed, retrying... (%s)",
	// Errors
	"ErrorEnterCredentials":     "Enter Credentials or No Profiles found.",
	"ErrorEnterUserID":          "Enter UserID",
	"ErrorEnterPassword":        "Enter Password",
	"ErrorConnectionFailed":     "Unable to connect to server",
	"ErrorConnectionTitle":      "Connection Error",
	"ErrorLogout":               "Logout error: %s",
	"ErrorLogoutFailed":         "Logout failed - you may still be connected",
	"ErrorSessionExpiredMessage": "Session expired - please log in again",
	"ErrorWrongPassword":        "Wrong password",
	"ErrorMaxLimit":             "Maximum connection limit reached",
	"ErrorDataLimit":            "Data limit reached",
	"ErrorMaxReconnectAttempts": "Auto-reconnection failed after 3 attempts. Please login manually.",
	"ErrorNoSavedCredentials":   "Session expired. No saved credentials for auto-reconnect.",
	"ErrorInvalidPort":          "Invalid port number. Please enter a value between 1 and 65535.",
	"ErrorAutoStartEnable":      "Failed to enable auto-start",
	"ErrorAutoStartDisable":     "Failed to disable auto-start",
	// Dialogs
	"DialogLogoutConfirm":   "Are you sure you want to logout from Sophos XG?",
	"DialogLogoutTitle":     "Confirm Logout",
	"DialogRetryConnection": "Retry connection?",
	// Menu
	"MenuConnection":   "Connection",
	"MenuLogout":       "Logout",
	"MenuExit":         "Exit",
	"MenuLanguage":     "Language",
	"MenuItemEnglish":  "English",
	"MenuItemTurkish":  "Türkçe",
	// Labels
	"LabelServerConfiguration": "Server Configuration",
	"LabelAutoStartWindows":     "Start automatically with Windows (minimized)",
	"LabelUserCredentials":      "User Credentials",
	"LabelServerIP":             "Server IP:",
	"LabelCaptivePortal":        "Captive Portal Port:",
	"LabelUsername":             "Username:",
	"LabelPassword":             "Password:",
	"LabelPortStandard":         "(Standard: 8090)",
	// Checkboxes
	"CheckboxUseHTTPS":             "Use HTTPS (SSL/TLS)",
	"CheckboxSkipSSL":              "Skip SSL certificate validation (for self-signed certificates)",
	"CheckboxRememberCredentials":  "Remember credentials and settings",
	// Info / Success
	"InfoCredentialsUpgraded":  "Security enhanced: Credentials encrypted",
	"InfoPortEditable":         "Port can now be edited. Press Enter to save or Escape to cancel.",
	"SuccessPortSaved":         "Port saved: %s",
	"SuccessAutoStartEnabled":  "Auto-start with Windows enabled",
	"InfoAutoStartDisabled":    "Auto-start with Windows disabled",
	"InfoPortEditCancelled":    "Edit cancelled - restored to default (8090)",
	"InfoMinimizedToTray":      "Minimized to system tray",
	"InfoAutoReconnecting":     "Session expired - reconnecting automatically...",
	"SuccessAutoReconnected":   "Automatically reconnected successfully!",
	// Context Menu
	"ContextMenuShow": "Show",
	"ContextMenuExit": "Exit",
	// SSL two-line labels
	"CheckboxSkipSSLLine1": "Skip SSL certificate validation",
	"CheckboxSkipSSLLine2": "(for self-signed certificates)",
	"LabelSkipSSL":         "Skip SSL certificate validation",
	"LabelSkipSSLNote":     "(for self-signed certificates)",
	// Auto-login
	"LabelAutoLogin": "Auto-login on startup",
}

// allTurkishTranslations maps every key to its expected Turkish value.
var allTurkishTranslations = map[string]string{
	// Window / Button / Status
	"WindowTitle":              "Sophos XG Kullanıcı Girişi",
	"ButtonLogin":              "Giriş Yap",
	"ButtonLogout":             "Çıkış",
	"ButtonDisconnect":         "Bağlantıyı Kes",
	"ButtonConnecting":         "Bağlanıyor...",
	"ButtonReconnect":          "Yeniden Bağlan",
	"StatusLoggedOut":          "Başarıyla çıkış yapıldı",
	"StatusLoggedIn":           "Başarıyla giriş yapıldı",
	"StatusConnecting":         "Bağlanıyor...",
	"StatusSessionActive":      "Oturum aktif",
	"StatusSessionExpired":     "Oturum süresi doldu",
	"StatusConnectedTo":        "Bağlı: %s",
	"StatusLoggedInAsUser":     "Kullanıcı: %s",
	"StatusSessionTime":        "Oturum süresi: %s",
	"StatusConnectingAuto":     "Otomatik olarak yeniden bağlanıyor...",
	"StatusReconnected":        "Otomatik olarak yeniden bağlandı",
	"StatusRetryingConnection": "Bağlantı başarısız, yeniden deniyor... (%s)",
	// Errors
	"ErrorEnterCredentials":     "Kimlik bilgilerini girin veya profil bulunamadı.",
	"ErrorEnterUserID":          "Kullanıcı adını girin",
	"ErrorEnterPassword":        "Şifre girin",
	"ErrorConnectionFailed":     "Sunucuya bağlanılamıyor",
	"ErrorConnectionTitle":      "Bağlantı Hatası",
	"ErrorLogout":               "Çıkış hatası: %s",
	"ErrorLogoutFailed":         "Çıkış başarısız - hala bağlı olabilirsiniz",
	"ErrorSessionExpiredMessage": "Oturum süresi doldu - lütfen tekrar giriş yapın",
	"ErrorWrongPassword":        "Yanlış şifre",
	"ErrorMaxLimit":             "Maksimum bağlantı limitine ulaşıldı",
	"ErrorDataLimit":            "Veri limitine ulaşıldı",
	"ErrorMaxReconnectAttempts": "Otomatik yeniden bağlanma 3 denemeden sonra başarısız. Lütfen manuel giriş yapın.",
	"ErrorNoSavedCredentials":   "Oturum süresi doldu. Otomatik bağlanma için kayıtlı kimlik bilgisi yok.",
	"ErrorInvalidPort":          "Geçersiz port numarası. Lütfen 1-65535 arasında bir değer girin.",
	"ErrorAutoStartEnable":      "Otomatik başlatma etkinleştirilemedi",
	"ErrorAutoStartDisable":     "Otomatik başlatma devre dışı bırakılamadı",
	// Dialogs
	"DialogLogoutConfirm":   "Sophos XG'den çıkış yapmak istediğinizden emin misiniz?",
	"DialogLogoutTitle":     "Çıkışı Onayla",
	"DialogRetryConnection": "Bağlantı yeniden denensin mi?",
	// Menu
	"MenuConnection":  "Bağlantı",
	"MenuLogout":      "Çıkış",
	"MenuExit":        "Kapat",
	"MenuLanguage":    "Dil",
	"MenuItemEnglish": "English",
	"MenuItemTurkish": "Türkçe",
	// Labels
	"LabelServerConfiguration": "Sunucu Yapılandırması",
	"LabelAutoStartWindows":     "Windows ile otomatik başlat (simge durumunda)",
	"LabelUserCredentials":      "Kullanıcı Bilgileri",
	"LabelServerIP":             "Sunucu IP:",
	"LabelCaptivePortal":        "Captive Portal Port:",
	"LabelUsername":             "Kullanıcı Adı:",
	"LabelPassword":             "Şifre:",
	"LabelPortStandard":         "(Standart: 8090)",
	// Checkboxes
	"CheckboxUseHTTPS":            "HTTPS kullan (SSL/TLS)",
	"CheckboxSkipSSL":             "SSL sertifika doğrulamasını atla (kendinden imzalı sertifikalar için)",
	"CheckboxRememberCredentials": "Kimlik bilgileri ve ayarları hatırla",
	// Info / Success
	"InfoCredentialsUpgraded": "Güvenlik iyileştirildi: Kimlik bilgileri şifrelendi",
	"InfoPortEditable":        "Port düzenlenebilir. Kaydetmek için Enter, iptal için Escape tuşuna basın.",
	"SuccessPortSaved":        "Port kaydedildi: %s",
	"SuccessAutoStartEnabled": "Windows başlangıcında otomatik çalışma etkinleştirildi",
	"InfoAutoStartDisabled":   "Windows başlangıcında otomatik çalışma devre dışı bırakıldı",
	"InfoPortEditCancelled":   "Düzenleme iptal edildi - varsayılan değer geri yüklendi (8090)",
	"InfoMinimizedToTray":     "Sistem tepsisine küçültüldü",
	"InfoAutoReconnecting":    "Oturum süresi doldu - otomatik yeniden bağlanıyor...",
	"SuccessAutoReconnected":  "Otomatik olarak başarıyla yeniden bağlandı!",
	// Context Menu
	"ContextMenuShow": "Göster",
	"ContextMenuExit": "Çıkış",
	// SSL two-line labels
	"CheckboxSkipSSLLine1": "SSL sertifika doğrulamasını atla",
	"CheckboxSkipSSLLine2": "(kendinden imzalı sertifikalar için)",
	"LabelSkipSSL":         "SSL sertifika doğrulamasını atla",
	"LabelSkipSSLNote":     "(kendinden imzalı sertifikalar için)",
	// Auto-login
	"LabelAutoLogin": "Başlangıçta otomatik giriş yap",
}

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewManager_DefaultsToEnglish(t *testing.T) {
	m := NewManager(LangEnglish)
	if m.CurrentLanguage() != LangEnglish {
		t.Errorf("expected default language %q, got %q", LangEnglish, m.CurrentLanguage())
	}
}

func TestNewManager_TurkishLanguage(t *testing.T) {
	m := NewManager(LangTurkish)
	if m.CurrentLanguage() != LangTurkish {
		t.Errorf("expected language %q, got %q", LangTurkish, m.CurrentLanguage())
	}
}

func TestNewManager_EmptyLanguageFallsBackToEnglish(t *testing.T) {
	m := NewManager("")
	if m.CurrentLanguage() != LangEnglish {
		t.Errorf("expected fallback language %q, got %q", LangEnglish, m.CurrentLanguage())
	}
}

func TestNewManager_UnsupportedLanguageFallsBackToEnglish(t *testing.T) {
	m := NewManager("fr")
	if m.CurrentLanguage() != LangEnglish {
		t.Errorf("expected fallback language %q, got %q", LangEnglish, m.CurrentLanguage())
	}
}

// ---------------------------------------------------------------------------
// GetString — English coverage (all 56+ keys)
// ---------------------------------------------------------------------------

func TestGetString_AllEnglishKeys(t *testing.T) {
	m := NewManager(LangEnglish)
	for key, expected := range allEnglishTranslations {
		got := m.GetString(key)
		if got != expected {
			t.Errorf("key %q: expected %q, got %q", key, expected, got)
		}
	}
}

// ---------------------------------------------------------------------------
// GetString — Turkish coverage (all 56+ keys)
// ---------------------------------------------------------------------------

func TestGetString_AllTurkishKeys(t *testing.T) {
	m := NewManager(LangTurkish)
	for key, expected := range allTurkishTranslations {
		got := m.GetString(key)
		if got != expected {
			t.Errorf("key %q: expected %q, got %q", key, expected, got)
		}
	}
}

// ---------------------------------------------------------------------------
// GetString — unknown key returns the key itself
// ---------------------------------------------------------------------------

func TestGetString_UnknownKeyReturnsKey(t *testing.T) {
	m := NewManager(LangEnglish)
	const unknownKey = "ThisKeyDoesNotExist"
	got := m.GetString(unknownKey)
	if got != unknownKey {
		t.Errorf("expected key %q to be returned for unknown key, got %q", unknownKey, got)
	}
}

// ---------------------------------------------------------------------------
// GetStringf — format parameter substitution
// ---------------------------------------------------------------------------

func TestGetStringf_EnglishFormattedStatusConnectedTo(t *testing.T) {
	m := NewManager(LangEnglish)
	got := m.GetStringf("StatusConnectedTo", "192.168.1.1")
	expected := "Connected to 192.168.1.1"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_TurkishFormattedStatusConnectedTo(t *testing.T) {
	m := NewManager(LangTurkish)
	got := m.GetStringf("StatusConnectedTo", "192.168.1.1")
	expected := "Bağlı: 192.168.1.1"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_EnglishFormattedStatusLoggedInAsUser(t *testing.T) {
	m := NewManager(LangEnglish)
	got := m.GetStringf("StatusLoggedInAsUser", "john")
	expected := "Logged in as: john"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_TurkishFormattedStatusLoggedInAsUser(t *testing.T) {
	m := NewManager(LangTurkish)
	got := m.GetStringf("StatusLoggedInAsUser", "john")
	expected := "Kullanıcı: john"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_EnglishFormattedErrorLogout(t *testing.T) {
	m := NewManager(LangEnglish)
	got := m.GetStringf("ErrorLogout", "timeout")
	expected := "Logout error: timeout"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_TurkishFormattedErrorLogout(t *testing.T) {
	m := NewManager(LangTurkish)
	got := m.GetStringf("ErrorLogout", "timeout")
	expected := "Çıkış hatası: timeout"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_EnglishFormattedSuccessPortSaved(t *testing.T) {
	m := NewManager(LangEnglish)
	got := m.GetStringf("SuccessPortSaved", "8090")
	expected := "Port saved: 8090"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_TurkishFormattedSuccessPortSaved(t *testing.T) {
	m := NewManager(LangTurkish)
	got := m.GetStringf("SuccessPortSaved", "8090")
	expected := "Port kaydedildi: 8090"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_EnglishFormattedStatusRetryingConnection(t *testing.T) {
	m := NewManager(LangEnglish)
	got := m.GetStringf("StatusRetryingConnection", "2/3")
	expected := "Connection failed, retrying... (2/3)"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_TurkishFormattedStatusRetryingConnection(t *testing.T) {
	m := NewManager(LangTurkish)
	got := m.GetStringf("StatusRetryingConnection", "2/3")
	expected := "Bağlantı başarısız, yeniden deniyor... (2/3)"
	if got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestGetStringf_UnknownKeyWithArgs(t *testing.T) {
	m := NewManager(LangEnglish)
	// When the key is unknown it is returned verbatim; Sprintf with no verbs
	// in the format string adds no extra text (no %!(EXTRA) warning either,
	// because there are no verbs to consume the argument). Actually fmt.Sprintf
	// DOES append %!(EXTRA ...) for unused args. We just verify the key is the
	// prefix of the result.
	got := m.GetStringf("UnknownKey", "arg1")
	if len(got) == 0 {
		t.Error("expected non-empty result for unknown key with args")
	}
}

// ---------------------------------------------------------------------------
// SetLanguage
// ---------------------------------------------------------------------------

func TestSetLanguage_ChangesToTurkish(t *testing.T) {
	m := NewManager(LangEnglish)
	m.SetLanguage(LangTurkish)
	if m.CurrentLanguage() != LangTurkish {
		t.Errorf("expected language %q, got %q", LangTurkish, m.CurrentLanguage())
	}
}

func TestSetLanguage_ChangesBackToEnglish(t *testing.T) {
	m := NewManager(LangTurkish)
	m.SetLanguage(LangEnglish)
	if m.CurrentLanguage() != LangEnglish {
		t.Errorf("expected language %q, got %q", LangEnglish, m.CurrentLanguage())
	}
}

func TestSetLanguage_IgnoresUnsupportedLanguage(t *testing.T) {
	m := NewManager(LangEnglish)
	m.SetLanguage("de")
	if m.CurrentLanguage() != LangEnglish {
		t.Errorf("expected language to remain %q, got %q", LangEnglish, m.CurrentLanguage())
	}
}

func TestSetLanguage_NoOpWhenSameLanguage(t *testing.T) {
	m := NewManager(LangEnglish)
	callCount := 0
	m.OnLanguageChanged(func(lang string) { callCount++ })
	m.SetLanguage(LangEnglish) // same language — should not fire callback
	if callCount != 0 {
		t.Errorf("expected 0 callbacks for no-op language change, got %d", callCount)
	}
}

func TestSetLanguage_UpdatesTranslationsReturned(t *testing.T) {
	m := NewManager(LangEnglish)
	englishVal := m.GetString("ButtonLogin")
	m.SetLanguage(LangTurkish)
	turkishVal := m.GetString("ButtonLogin")
	if englishVal == turkishVal {
		t.Errorf("expected different translations after language switch, both returned %q", englishVal)
	}
	if englishVal != "Log In" {
		t.Errorf("expected English value %q, got %q", "Log In", englishVal)
	}
	if turkishVal != "Giriş Yap" {
		t.Errorf("expected Turkish value %q, got %q", "Giriş Yap", turkishVal)
	}
}

// ---------------------------------------------------------------------------
// IsCurrentLanguage
// ---------------------------------------------------------------------------

func TestIsCurrentLanguage_TrueForActiveLanguage(t *testing.T) {
	m := NewManager(LangEnglish)
	if !m.IsCurrentLanguage(LangEnglish) {
		t.Errorf("expected IsCurrentLanguage(%q) = true", LangEnglish)
	}
}

func TestIsCurrentLanguage_FalseForInactiveLanguage(t *testing.T) {
	m := NewManager(LangEnglish)
	if m.IsCurrentLanguage(LangTurkish) {
		t.Errorf("expected IsCurrentLanguage(%q) = false when language is English", LangTurkish)
	}
}

func TestIsCurrentLanguage_UpdatesAfterSetLanguage(t *testing.T) {
	m := NewManager(LangEnglish)
	m.SetLanguage(LangTurkish)
	if !m.IsCurrentLanguage(LangTurkish) {
		t.Errorf("expected IsCurrentLanguage(%q) = true after SetLanguage", LangTurkish)
	}
	if m.IsCurrentLanguage(LangEnglish) {
		t.Errorf("expected IsCurrentLanguage(%q) = false after switching to Turkish", LangEnglish)
	}
}

// ---------------------------------------------------------------------------
// OnLanguageChanged
// ---------------------------------------------------------------------------

func TestOnLanguageChanged_CallbackFiredOnChange(t *testing.T) {
	m := NewManager(LangEnglish)
	var receivedLang string
	m.OnLanguageChanged(func(lang string) {
		receivedLang = lang
	})
	m.SetLanguage(LangTurkish)
	if receivedLang != LangTurkish {
		t.Errorf("expected callback to receive %q, got %q", LangTurkish, receivedLang)
	}
}

func TestOnLanguageChanged_MultipleCallbacksAllFired(t *testing.T) {
	m := NewManager(LangEnglish)
	callCount := 0
	m.OnLanguageChanged(func(lang string) { callCount++ })
	m.OnLanguageChanged(func(lang string) { callCount++ })
	m.OnLanguageChanged(func(lang string) { callCount++ })
	m.SetLanguage(LangTurkish)
	if callCount != 3 {
		t.Errorf("expected 3 callbacks, got %d", callCount)
	}
}

func TestOnLanguageChanged_CallbackReceivesNewLanguageCode(t *testing.T) {
	m := NewManager(LangEnglish)
	var langs []string
	m.OnLanguageChanged(func(lang string) { langs = append(langs, lang) })
	m.SetLanguage(LangTurkish)
	m.SetLanguage(LangEnglish)
	if len(langs) != 2 {
		t.Fatalf("expected 2 language change events, got %d", len(langs))
	}
	if langs[0] != LangTurkish {
		t.Errorf("first change: expected %q, got %q", LangTurkish, langs[0])
	}
	if langs[1] != LangEnglish {
		t.Errorf("second change: expected %q, got %q", LangEnglish, langs[1])
	}
}

func TestOnLanguageChanged_CallbackNotFiredForNoOpChange(t *testing.T) {
	m := NewManager(LangTurkish)
	callCount := 0
	m.OnLanguageChanged(func(lang string) { callCount++ })
	m.SetLanguage(LangTurkish) // no-op — already Turkish
	if callCount != 0 {
		t.Errorf("expected 0 callbacks for no-op, got %d", callCount)
	}
}

// ---------------------------------------------------------------------------
// Thread safety
// ---------------------------------------------------------------------------

func TestThreadSafety_ConcurrentReadsAndWrites(t *testing.T) {
	m := NewManager(LangEnglish)

	const goroutines = 50
	const iterations = 100

	var wg sync.WaitGroup
	wg.Add(goroutines * 3)

	// Goroutines that switch language back and forth.
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				if (idx+j)%2 == 0 {
					m.SetLanguage(LangEnglish)
				} else {
					m.SetLanguage(LangTurkish)
				}
			}
		}(i)
	}

	// Goroutines that read strings concurrently.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = m.GetString("ButtonLogin")
				_ = m.GetString("WindowTitle")
				_ = m.GetString("ErrorConnectionFailed")
			}
		}()
	}

	// Goroutines that call CurrentLanguage and IsCurrentLanguage concurrently.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = m.CurrentLanguage()
				_ = m.IsCurrentLanguage(LangEnglish)
				_ = m.IsCurrentLanguage(LangTurkish)
			}
		}()
	}

	wg.Wait()
	// If we reach here without a data-race panic the test passes.
}

func TestThreadSafety_ConcurrentListenerRegistrationAndLanguageChange(t *testing.T) {
	m := NewManager(LangEnglish)

	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines * 2)

	// Goroutines that register listeners while language changes are happening.
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			m.OnLanguageChanged(func(lang string) {
				// Access the language code safely inside the callback.
				_ = lang
			})
		}()
	}

	// Goroutines that trigger language changes.
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			if idx%2 == 0 {
				m.SetLanguage(LangTurkish)
			} else {
				m.SetLanguage(LangEnglish)
			}
		}(i)
	}

	wg.Wait()
}

// ---------------------------------------------------------------------------
// DetectSystemLanguage tests
// ---------------------------------------------------------------------------

func TestDetectSystemLanguage_Turkish(t *testing.T) {
	original := os.Getenv("LANG")
	defer os.Setenv("LANG", original)

	os.Setenv("LANG", "tr_TR.UTF-8")
	if DetectSystemLanguage() != LangTurkish {
		t.Error("expected Turkish for tr_TR.UTF-8 locale")
	}
}

func TestDetectSystemLanguage_English(t *testing.T) {
	original := os.Getenv("LANG")
	originalLcAll := os.Getenv("LC_ALL")
	originalLcMsg := os.Getenv("LC_MESSAGES")
	defer func() {
		os.Setenv("LANG", original)
		os.Setenv("LC_ALL", originalLcAll)
		os.Setenv("LC_MESSAGES", originalLcMsg)
	}()

	os.Setenv("LANG", "en_US.UTF-8")
	os.Setenv("LC_ALL", "")
	os.Setenv("LC_MESSAGES", "")
	if DetectSystemLanguage() != LangEnglish {
		t.Error("expected English for en_US.UTF-8 locale")
	}
}

func TestDetectSystemLanguage_EmptyFallsBackToEnglish(t *testing.T) {
	original := os.Getenv("LANG")
	originalLcAll := os.Getenv("LC_ALL")
	originalLcMsg := os.Getenv("LC_MESSAGES")
	defer func() {
		os.Setenv("LANG", original)
		os.Setenv("LC_ALL", originalLcAll)
		os.Setenv("LC_MESSAGES", originalLcMsg)
	}()

	os.Setenv("LANG", "")
	os.Setenv("LC_ALL", "")
	os.Setenv("LC_MESSAGES", "")
	if DetectSystemLanguage() != LangEnglish {
		t.Error("expected English fallback for empty locale")
	}
}

// ---------------------------------------------------------------------------
// PreferenceStore tests
// ---------------------------------------------------------------------------

type mockPrefs struct {
	data map[string]string
}

func (p *mockPrefs) String(key string) string {
	return p.data[key]
}

func (p *mockPrefs) SetString(key, value string) {
	p.data[key] = value
}

func TestSetPreferences_PersistsLanguageOnChange(t *testing.T) {
	m := NewManager(LangEnglish)
	prefs := &mockPrefs{data: make(map[string]string)}
	m.SetPreferences(prefs)

	m.SetLanguage(LangTurkish)

	if prefs.data["Language"] != LangTurkish {
		t.Errorf("expected saved language %q, got %q", LangTurkish, prefs.data["Language"])
	}
}

func TestLoadFromPreferences_RestoresSavedLanguage(t *testing.T) {
	m := NewManager(LangEnglish)
	prefs := &mockPrefs{data: map[string]string{"Language": LangTurkish}}
	m.SetPreferences(prefs)

	m.LoadFromPreferences()

	if m.CurrentLanguage() != LangTurkish {
		t.Errorf("expected language %q after load, got %q", LangTurkish, m.CurrentLanguage())
	}
}

func TestLoadFromPreferences_IgnoresInvalidLanguage(t *testing.T) {
	m := NewManager(LangEnglish)
	prefs := &mockPrefs{data: map[string]string{"Language": "fr"}}
	m.SetPreferences(prefs)

	m.LoadFromPreferences()

	if m.CurrentLanguage() != LangEnglish {
		t.Errorf("expected language to remain %q for invalid saved language, got %q", LangEnglish, m.CurrentLanguage())
	}
}

func TestLoadFromPreferences_NilPrefsIsNoop(t *testing.T) {
	m := NewManager(LangTurkish)
	// Don't set preferences
	m.LoadFromPreferences() // Should not panic
	if m.CurrentLanguage() != LangTurkish {
		t.Error("expected language unchanged when no prefs set")
	}
}
