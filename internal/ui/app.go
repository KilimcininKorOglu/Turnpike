package ui

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"github.com/KilimcininKorOglu/SophosXG-User-Client/internal/auth"
	"github.com/KilimcininKorOglu/SophosXG-User-Client/internal/autostart"
	"github.com/KilimcininKorOglu/SophosXG-User-Client/internal/cli"
	"github.com/KilimcininKorOglu/SophosXG-User-Client/internal/config"
	"github.com/KilimcininKorOglu/SophosXG-User-Client/internal/i18n"
	"github.com/KilimcininKorOglu/SophosXG-User-Client/internal/logging"
)

// Application is the main application controller
type Application struct {
	mu sync.Mutex

	fyneApp    fyne.App
	mainWindow fyne.Window

	// Services
	authService  *auth.Service
	credManager  *config.CredentialManager
	logger       *logging.Service
	localization *i18n.Manager
	autoStart    *autostart.Manager
	notification *NotificationManager

	// UI Widgets
	serverIPEntry     *widget.Entry
	portEntry         *widget.Entry
	usernameEntry     *widget.Entry
	passwordEntry     *widget.Entry
	useHTTPSCheck     *widget.Check
	skipSSLCheck      *widget.Check
	rememberCheck     *widget.Check
	autoStartCheck    *widget.Check
	autoLoginCheck    *widget.Check
	loginButton       *widget.Button
	logoutButton      *widget.Button
	quickLogoutButton *widget.Button
	logoutMenuItem    *fyne.MenuItem
	statusLabel       *widget.Label
	statusDot         *canvas.Circle
	connectionStatus  *widget.Label
	userInfo          *widget.Label
	sessionTimerLabel *widget.Label
	statusIndicator    *canvas.Circle
	connectionPanel    *fyne.Container
	connectionPanelBg  *canvas.Rectangle

	// State
	currentStatus       auth.LoginStatus
	isTryingToLogIn     bool
	isLive              bool
	isConnecting        bool
	userInitiatedLogout bool
	reconnectAttempts   int
	isReconnecting      bool
	loginTime           time.Time
	userID              string

	// Timers
	refreshTicker *time.Ticker
	sessionTicker *time.Ticker
	refreshDone   chan struct{}
	sessionDone   chan struct{}
}

// NewApplication creates and initializes the main application
func NewApplication(startMinimized bool) *Application {
	a := &Application{
		isTryingToLogIn: true,
		currentStatus:   auth.StatusLoggedOut,
		refreshDone:     make(chan struct{}),
		sessionDone:     make(chan struct{}),
	}

	// Determine settings directory
	exePath, _ := os.Executable()
	settingsDir := filepath.Dir(exePath)

	// Initialize Fyne first (needed for preferences)
	a.fyneApp = app.NewWithID("com.sophos-xg-login.app")
	a.fyneApp.Settings().SetTheme(&OLEDTheme{})

	// Initialize localization with system language detection, then load saved preference
	detectedLang := i18n.DetectSystemLanguage()
	a.localization = i18n.NewManager(detectedLang)
	a.localization.SetPreferences(a.fyneApp.Preferences())
	a.localization.LoadFromPreferences()

	// Initialize remaining services
	a.logger = logging.NewService(settingsDir, false)
	a.credManager = config.NewCredentialManager(settingsDir)
	a.notification = NewNotificationManager()

	creds := a.credManager.GetCredentials()
	a.authService = auth.NewService(
		creds.ServerIP,
		creds.CaptivePortalPort,
		creds.UseHTTPS,
		creds.SkipSSLValidation,
		config.HTTPRequestTimeout,
	)

	a.autoStart = autostart.NewManager("Sophos XG User Login", exePath)

	a.logger.LogApplicationStart(cli.AppVersion)

	a.buildUI()

	// Auto-login if enabled and credentials exist
	if a.fyneApp.Preferences().BoolWithFallback("AutoLogin", false) {
		creds2 := a.credManager.GetCredentials()
		if creds2.Username != "" && creds2.Password != "" {
			go a.performLogin()
		}
	}

	// Handle password migration
	if a.credManager.NeedsPasswordMigration() {
		a.logger.LogSecurityEvent("PasswordMigration", "Starting migration", "")
		a.credManager.MigratePasswordEncryption()
		a.notification.ShowInfo(a.localization.GetString("InfoCredentialsUpgraded"), config.NotificationDurationInfo)
		a.logger.LogSecurityEvent("PasswordMigration", "Completed", "")
	}

	// System tray integration
	a.setupSystemTray()

	// Handle minimized start
	if startMinimized {
		a.mainWindow.Hide()
		a.logger.LogUIEvent("AutoStart", "Application started minimized")
	}

	return a
}

func (a *Application) buildUI() {
	a.mainWindow = a.fyneApp.NewWindow(a.localization.GetString("WindowTitle"))
	a.mainWindow.Resize(fyne.NewSize(420, 650))
	a.mainWindow.SetFixedSize(false)
	a.mainWindow.CenterOnScreen()

	// Override close to minimize to tray
	a.mainWindow.SetCloseIntercept(func() {
		a.mainWindow.Hide()
		a.logger.LogUIEvent("WindowClosing", "Window minimized to system tray instead of closing")
	})

	// Build form elements
	creds := a.credManager.GetCredentials()

	a.serverIPEntry = widget.NewEntry()
	a.serverIPEntry.SetText(creds.ServerIP)
	a.serverIPEntry.OnChanged = func(s string) {
		a.credManager.SetServerIP(s)
	}

	a.portEntry = widget.NewEntry()
	a.portEntry.SetText(strconv.Itoa(creds.CaptivePortalPort))
	a.portEntry.OnChanged = func(s string) {
		port, err := strconv.Atoi(s)
		if err == nil && config.ValidatePort(port) {
			a.credManager.SetCaptivePortalPort(port)
		}
	}

	a.usernameEntry = widget.NewEntry()
	a.usernameEntry.SetText(creds.Username)
	a.usernameEntry.OnChanged = func(s string) {
		a.userID = s
		a.credManager.SetUsername(s)
	}
	a.usernameEntry.OnSubmitted = func(_ string) {
		go a.handleLoginLogout()
	}
	a.userID = creds.Username

	a.passwordEntry = widget.NewPasswordEntry()
	a.passwordEntry.SetText(creds.Password)
	a.passwordEntry.OnChanged = func(s string) {
		a.credManager.SetPassword(s)
	}
	a.passwordEntry.OnSubmitted = func(_ string) {
		go a.handleLoginLogout()
	}

	a.useHTTPSCheck = widget.NewCheck(a.localization.GetString("CheckboxUseHTTPS"), func(checked bool) {
		a.credManager.SetUseHTTPS(checked)
	})
	a.useHTTPSCheck.SetChecked(creds.UseHTTPS)

	a.skipSSLCheck = widget.NewCheck(a.localization.GetString("CheckboxSkipSSL"), func(checked bool) {
		a.credManager.SetSkipSSLValidation(checked)
	})
	a.skipSSLCheck.SetChecked(creds.SkipSSLValidation)

	a.rememberCheck = widget.NewCheck(a.localization.GetString("CheckboxRememberCredentials"), func(checked bool) {
		a.credManager.SetRememberCredentials(checked)
	})
	a.rememberCheck.SetChecked(a.credManager.RememberCredentials())

	a.autoStartCheck = widget.NewCheck(a.localization.GetString("LabelAutoStartWindows"), func(checked bool) {
		if checked {
			if err := a.autoStart.Enable(); err != nil {
				a.logger.LogError(err, "Failed to enable auto-start", "AutoStart")
				a.notification.ShowError(a.localization.GetString("ErrorAutoStartEnable"), config.NotificationDurationError)
				a.autoStartCheck.SetChecked(false)
				return
			}
			a.logger.LogUIEvent("AutoStartEnabled", "")
			a.notification.ShowSuccess(a.localization.GetString("SuccessAutoStartEnabled"), config.NotificationDurationSuccess)
		} else {
			if err := a.autoStart.Disable(); err != nil {
				a.logger.LogError(err, "Failed to disable auto-start", "AutoStart")
				a.notification.ShowError(a.localization.GetString("ErrorAutoStartDisable"), config.NotificationDurationError)
				a.autoStartCheck.SetChecked(true)
				return
			}
			a.logger.LogUIEvent("AutoStartDisabled", "")
			a.notification.ShowInfo(a.localization.GetString("InfoAutoStartDisabled"), config.NotificationDurationInfo)
		}
	})
	a.autoStartCheck.SetChecked(a.autoStart.IsEnabled())

	a.autoLoginCheck = widget.NewCheck(a.localization.GetString("LabelAutoLogin"), func(checked bool) {
		a.fyneApp.Preferences().SetBool("AutoLogin", checked)
	})
	a.autoLoginCheck.SetChecked(a.fyneApp.Preferences().BoolWithFallback("AutoLogin", false))

	// Login/Logout buttons
	a.loginButton = widget.NewButton(a.localization.GetString("ButtonLogin"), func() {
		go a.handleLoginLogout()
	})
	a.loginButton.Importance = widget.HighImportance

	a.logoutButton = widget.NewButton(a.localization.GetString("ButtonLogout"), func() {
		a.confirmLogout()
	})
	a.logoutButton.Importance = widget.DangerImportance
	a.logoutButton.Hide()

	// Status bar
	a.statusDot = canvas.NewCircle(ColorDisabled)
	a.statusDot.Resize(fyne.NewSize(10, 10))
	a.statusLabel = widget.NewLabel("")

	// Connection status panel
	a.statusIndicator = canvas.NewCircle(ColorSuccess)
	a.statusIndicator.Resize(fyne.NewSize(12, 12))
	a.connectionStatus = widget.NewLabel("")
	a.connectionStatus.TextStyle = fyne.TextStyle{Bold: true}
	a.userInfo = widget.NewLabel("")
	a.sessionTimerLabel = widget.NewLabel("")

	a.quickLogoutButton = widget.NewButton(a.localization.GetString("ButtonLogout"), func() {
		go a.performLogout()
	})
	a.quickLogoutButton.Importance = widget.DangerImportance

	a.connectionPanelBg = canvas.NewRectangle(ColorNearBlack)
	a.connectionPanelBg.CornerRadius = 5

	panelContent := container.NewHBox(
		a.statusIndicator,
		container.NewVBox(a.connectionStatus, a.userInfo, a.sessionTimerLabel),
		layout.NewSpacer(),
		a.quickLogoutButton,
	)
	a.connectionPanel = container.NewStack(a.connectionPanelBg, container.NewPadded(panelContent))
	a.connectionPanel.Hide()

	// Section headers
	serverHeader := canvas.NewRectangle(ColorTurquoise)
	serverHeader.CornerRadius = 4
	serverHeaderLabel := widget.NewLabel(a.localization.GetString("LabelServerConfiguration"))
	serverHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	serverHeaderBox := container.NewStack(serverHeader, container.NewPadded(serverHeaderLabel))

	userHeader := canvas.NewRectangle(ColorOrange)
	userHeader.CornerRadius = 4
	userHeaderLabel := widget.NewLabel(a.localization.GetString("LabelUserCredentials"))
	userHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	userHeaderBox := container.NewStack(userHeader, container.NewPadded(userHeaderLabel))

	// Server configuration form
	serverForm := container.NewVBox(
		serverHeaderBox,
		container.NewGridWithColumns(2,
			widget.NewLabel(a.localization.GetString("LabelServerIP")),
			a.serverIPEntry,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel(a.localization.GetString("LabelCaptivePortal")),
			container.NewHBox(a.portEntry, widget.NewLabel(a.localization.GetString("LabelPortStandard"))),
		),
		a.useHTTPSCheck,
		a.skipSSLCheck,
	)

	// User credentials form
	userForm := container.NewVBox(
		userHeaderBox,
		container.NewGridWithColumns(2,
			widget.NewLabel(a.localization.GetString("LabelUsername")),
			a.usernameEntry,
		),
		container.NewGridWithColumns(2,
			widget.NewLabel(a.localization.GetString("LabelPassword")),
			a.passwordEntry,
		),
		a.rememberCheck,
		a.autoStartCheck,
		a.autoLoginCheck,
	)

	// Main content panel
	mainPanel := container.NewVBox(
		serverForm,
		widget.NewSeparator(),
		userForm,
	)

	// Title
	title := widget.NewLabel(a.localization.GetString("WindowTitle"))
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	// Button row
	buttonRow := container.NewHBox(
		layout.NewSpacer(),
		a.loginButton,
		a.logoutButton,
		layout.NewSpacer(),
	)

	// Status bar
	statusBar := container.NewHBox(
		container.NewWithoutLayout(a.statusDot),
		a.statusLabel,
	)

	// Notification overlay
	notificationOverlay := container.NewVBox()
	a.notification.SetOverlay(notificationOverlay)

	// Full layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		container.NewPadded(mainPanel),
		a.connectionPanel,
		buttonRow,
		layout.NewSpacer(),
		notificationOverlay,
		widget.NewSeparator(),
		statusBar,
	)

	a.mainWindow.SetContent(container.NewPadded(content))

	// Setup menus
	a.setupMenus()

	// Register language change callback
	a.localization.OnLanguageChanged(func(lang string) {
		a.applyLocalization()
	})
}

func (a *Application) setupMenus() {
	a.logoutMenuItem = fyne.NewMenuItem(a.localization.GetString("MenuLogout"), func() {
		a.confirmLogout()
	})
	a.logoutMenuItem.Disabled = true

	connectionMenu := fyne.NewMenu(a.localization.GetString("MenuConnection"),
		a.logoutMenuItem,
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem(a.localization.GetString("MenuExit"), func() {
			a.trueApplicationExit()
		}),
	)

	languageMenu := fyne.NewMenu(a.localization.GetString("MenuLanguage"),
		fyne.NewMenuItem("English", func() {
			a.logger.LogConfigurationChange("Language", a.localization.CurrentLanguage(), "en")
			a.localization.SetLanguage("en")
		}),
		fyne.NewMenuItem("Türkçe", func() {
			a.logger.LogConfigurationChange("Language", a.localization.CurrentLanguage(), "tr")
			a.localization.SetLanguage("tr")
		}),
	)

	a.mainWindow.SetMainMenu(fyne.NewMainMenu(connectionMenu, languageMenu))
}

func (a *Application) setupSystemTray() {
	if desk, ok := a.fyneApp.(desktop.App); ok {
		menu := fyne.NewMenu("",
			fyne.NewMenuItem(a.localization.GetString("ContextMenuShow"), func() {
				a.mainWindow.Show()
				a.mainWindow.RequestFocus()
				a.logger.LogUIEvent("WindowRestore", "Window restored from system tray")
			}),
			fyne.NewMenuItem(a.localization.GetString("ContextMenuExit"), func() {
				a.trueApplicationExit()
			}),
		)
		desk.SetSystemTrayMenu(menu)
	}
}

func (a *Application) handleLoginLogout() {
	if a.isTryingToLogIn {
		a.performLogin()
	} else {
		a.performLogout()
	}
}

func (a *Application) performLogin() {
	username := a.usernameEntry.Text
	password := a.passwordEntry.Text

	// Validate
	if username == "" && password == "" {
		a.notification.ShowWarning(a.localization.GetString("ErrorEnterCredentials"), config.NotificationDurationWarning)
		return
	}
	if username == "" {
		a.notification.ShowWarning(a.localization.GetString("ErrorEnterUserID"), config.NotificationDurationWarning)
		return
	}
	if password == "" {
		a.notification.ShowWarning(a.localization.GetString("ErrorEnterPassword"), config.NotificationDurationWarning)
		return
	}

	a.setConnectingState(true)
	a.logger.LogInfo("Login attempt started", "Authentication")

	// Update service settings
	creds := a.credManager.GetCredentials()
	a.authService.UpdateSettings(creds.ServerIP, creds.CaptivePortalPort, creds.UseHTTPS, creds.SkipSSLValidation)

	result, err := a.authService.LoginAsync(username, password)
	if err != nil {
		a.logger.LogError(err, "Login process failed", "Authentication")
		a.setStatus(a.localization.GetString("ErrorConnectionFailed"))
		a.setStatusColor(ColorError)
		a.notification.ShowError(a.localization.GetString("ErrorConnectionFailed"), config.NotificationDurationError)
		a.setConnectingState(false)
		return
	}

	a.currentStatus = result.Status

	if result.Success {
		a.logger.LogAuthenticationAttempt(username, creds.ServerIP, true, "Login successful")
		a.isTryingToLogIn = false
		a.userID = username
		a.isLive = true
		a.userInitiatedLogout = false
		a.reconnectAttempts = 0

		a.showLoginSuccess(username)
		a.setStatus(a.localization.GetString("StatusLoggedIn"))
		a.notification.ShowSuccess(
			a.localization.GetString("StatusLoggedIn")+" - "+username,
			config.NotificationDurationSuccess,
		)

		// Save credentials
		if a.credManager.RememberCredentials() {
			a.credManager.SaveCredentials()
			a.logger.LogInfo("Credentials saved (encrypted)", "Security")
		}

		// Start session monitoring
		a.startRefreshTimer()
		a.startSessionTimer()
		// Clear password from memory after successful login
		a.clearPasswordFromMemory()
	} else {
		errorMsg := a.getLoginErrorMessage(result.Status)
		a.logger.LogAuthenticationAttempt(username, creds.ServerIP, false, errorMsg)
		a.setStatus(errorMsg)

		var statusColor color.Color
		switch result.Status {
		case auth.StatusWrongPassword:
			statusColor = ColorError
		case auth.StatusMaxLimit:
			statusColor = ColorInfo // Blue (matching .NET)
		case auth.StatusDataLimit:
			statusColor = ColorWarning // Yellow/Orange (matching .NET)
		default:
			statusColor = ColorError
		}
		a.setStatusColor(statusColor)
		a.notification.ShowError(errorMsg, config.NotificationDurationError)

		// Clear password from memory after failed login
		a.clearPasswordFromMemory()
	}

	a.setConnectingState(false)
}

func (a *Application) performLogout() {
	a.mu.Lock()
	a.userInitiatedLogout = true
	a.mu.Unlock()

	a.logger.LogInfo("User initiated logout", "Logout")
	a.setConnectingState(true)

	creds := a.credManager.GetCredentials()
	success, err := a.authService.LogoutAsync(a.userID)

	if err != nil {
		a.logger.LogError(err, "Logout failed", "Authentication")
		a.notification.ShowError(a.localization.GetString("ErrorLogoutFailed"), config.NotificationDurationError)
		a.setConnectingState(false)
		return
	}

	if success {
		a.logger.LogLogout(a.userID, creds.ServerIP, true, "")
		a.showLoginFailure()
		a.currentStatus = auth.StatusLoggedOut
		a.setStatus(a.localization.GetString("StatusLoggedOut"))
		a.notification.ShowSuccess(a.localization.GetString("StatusLoggedOut"), 2500)

		a.isTryingToLogIn = true
		a.isLive = false
		a.stopRefreshTimer()
		a.stopSessionTimer()
	} else {
		a.logger.LogLogout(a.userID, creds.ServerIP, false, "Logout failed")
		a.setStatus(a.localization.GetString("ErrorLogoutFailed"))
		a.notification.ShowError(a.localization.GetString("ErrorLogoutFailed"), config.NotificationDurationError)
	}

	a.setConnectingState(false)
}

func (a *Application) confirmLogout() {
	dialog.ShowConfirm(
		a.localization.GetString("DialogLogoutTitle"),
		a.localization.GetString("DialogLogoutConfirm"),
		func(confirmed bool) {
			if confirmed {
				go a.performLogout()
			}
		},
		a.mainWindow,
	)
}

func (a *Application) showLoginSuccess(username string) {
	a.loginTime = time.Now()
	a.connectionPanel.Show()
	a.logoutButton.Show()
	a.loginButton.SetText(a.localization.GetString("ButtonDisconnect"))

	creds := a.credManager.GetCredentials()
	a.connectionStatus.SetText(a.localization.GetStringf("StatusConnectedTo", creds.ServerIP))
	a.userInfo.SetText(a.localization.GetStringf("StatusLoggedInAsUser", username))
	a.sessionTimerLabel.SetText(a.localization.GetStringf("StatusSessionTime", "00:00:00"))

	a.statusIndicator.FillColor = ColorSuccess
	a.statusIndicator.Refresh()
	a.connectionPanelBg.FillColor = ColorNearBlack
	a.connectionPanelBg.Refresh()
	a.quickLogoutButton.SetText(a.localization.GetString("ButtonLogout"))
	a.setStatusColor(ColorSuccess)

	if a.logoutMenuItem != nil {
		a.logoutMenuItem.Disabled = false
	}
}

func (a *Application) showLoginFailure() {
	a.connectionPanel.Hide()
	a.logoutButton.Hide()
	a.loginButton.SetText(a.localization.GetString("ButtonLogin"))
	a.setStatusColor(ColorWhite)

	if a.logoutMenuItem != nil {
		a.logoutMenuItem.Disabled = true
	}
}

func (a *Application) showSessionExpired() {
	a.connectionStatus.SetText(a.localization.GetString("StatusSessionExpired"))
	a.statusIndicator.FillColor = ColorAmberBorder
	a.statusIndicator.Refresh()
	a.connectionPanelBg.FillColor = ColorDarkAmber
	a.connectionPanelBg.Refresh()
	a.quickLogoutButton.SetText(a.localization.GetString("ButtonReconnect"))
	a.setStatusColor(ColorWarning)
	a.stopSessionTimer()
}

func (a *Application) setConnectingState(connecting bool) {
	a.isConnecting = connecting
	if connecting {
		a.loginButton.Disable()
		a.serverIPEntry.Disable()
		a.portEntry.Disable()
		a.usernameEntry.Disable()
		a.passwordEntry.Disable()
		a.useHTTPSCheck.Disable()
		a.skipSSLCheck.Disable()
		a.setStatus(a.localization.GetString("StatusConnecting"))
		a.setStatusColor(ColorInfo)
	} else {
		a.loginButton.Enable()
		a.serverIPEntry.Enable()
		a.portEntry.Enable()
		a.usernameEntry.Enable()
		a.passwordEntry.Enable()
		a.useHTTPSCheck.Enable()
		a.skipSSLCheck.Enable()
		if a.isTryingToLogIn {
			a.loginButton.SetText(a.localization.GetString("ButtonLogin"))
		} else {
			a.loginButton.SetText(a.localization.GetString("ButtonDisconnect"))
		}
	}
}

func (a *Application) clearPasswordFromMemory() {
	a.passwordEntry.SetText("")
}

func (a *Application) setStatus(message string) {
	a.statusLabel.SetText(message)
}

func (a *Application) setStatusColor(c color.Color) {
	a.statusDot.FillColor = c
	a.statusDot.Refresh()
}

func (a *Application) startRefreshTimer() {
	a.stopRefreshTimer()
	intervalSec := a.fyneApp.Preferences().IntWithFallback("SessionCheckInterval", 30)
	a.refreshTicker = time.NewTicker(time.Duration(intervalSec) * time.Second)
	go func() {
		for {
			select {
			case <-a.refreshTicker.C:
				a.refreshTimerTick()
			case <-a.refreshDone:
				return
			}
		}
	}()
}

func (a *Application) stopRefreshTimer() {
	if a.refreshTicker != nil {
		a.refreshTicker.Stop()
		select {
		case a.refreshDone <- struct{}{}:
		default:
		}
	}
}

func (a *Application) startSessionTimer() {
	a.stopSessionTimer()
	a.sessionTicker = time.NewTicker(1 * time.Second)
	go func() {
		for {
			select {
			case <-a.sessionTicker.C:
				if !a.loginTime.IsZero() {
					duration := time.Since(a.loginTime)
					hours := int(duration.Hours())
					minutes := int(duration.Minutes()) % 60
					seconds := int(duration.Seconds()) % 60
					timeStr := fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
					a.sessionTimerLabel.SetText(a.localization.GetStringf("StatusSessionTime", timeStr))
				}
			case <-a.sessionDone:
				return
			}
		}
	}()
}

func (a *Application) stopSessionTimer() {
	if a.sessionTicker != nil {
		a.sessionTicker.Stop()
		select {
		case a.sessionDone <- struct{}{}:
		default:
		}
	}
}

func (a *Application) refreshTimerTick() {
	if a.userID == "" {
		return
	}

	isOnline, err := a.authService.CheckSessionAsync(a.userID)
	if err != nil || !isOnline {
		a.currentStatus = auth.StatusLogInAgain
		a.attemptAutomaticReconnect()
	} else {
		a.currentStatus = auth.StatusLoggedIn
		a.setStatus(a.localization.GetString("StatusSessionActive"))
	}
}

func (a *Application) attemptAutomaticReconnect() {
	a.mu.Lock()
	if a.userInitiatedLogout || a.isReconnecting {
		a.mu.Unlock()
		return
	}
	a.isReconnecting = true
	a.mu.Unlock()

	defer func() {
		a.mu.Lock()
		a.isReconnecting = false
		a.mu.Unlock()
	}()

	a.logger.LogInfo(fmt.Sprintf("Auto-reconnect attempt %d/%d", a.reconnectAttempts+1, config.MaxReconnectAttempts), "AutoReconnect")
	a.showSessionExpired()
	a.notification.ShowInfo(a.localization.GetString("InfoAutoReconnecting"), config.NotificationDurationInfo)

	time.Sleep(config.RetryDelay)

	creds := a.credManager.GetCredentials()
	if creds.Username == "" || creds.Password == "" {
		a.handleNoCredentialsForReconnect()
		return
	}

	a.authService.UpdateSettings(creds.ServerIP, creds.CaptivePortalPort, creds.UseHTTPS, creds.SkipSSLValidation)
	result, _ := a.authService.LoginAsync(creds.Username, creds.Password)

	if result.Success {
		a.reconnectAttempts = 0
		a.currentStatus = auth.StatusLoggedIn
		a.logger.LogInfo("Auto-reconnection successful", "AutoReconnect")
		a.showLoginSuccess(creds.Username)
		a.notification.ShowSuccess(a.localization.GetString("SuccessAutoReconnected"), config.NotificationDurationSuccess)
		a.setStatus(a.localization.GetString("StatusReconnected"))
	} else {
		a.reconnectAttempts++
		a.logger.LogWarning(fmt.Sprintf("Auto-reconnect failed. Attempt %d/%d", a.reconnectAttempts, config.MaxReconnectAttempts), "AutoReconnect")

		if a.reconnectAttempts >= config.MaxReconnectAttempts {
			a.handleMaxReconnectAttemptsReached()
		} else {
			retryInfo := fmt.Sprintf("%d/%d", a.reconnectAttempts, config.MaxReconnectAttempts)
			a.setStatus(a.localization.GetStringf("StatusRetryingConnection", retryInfo))
		}
	}
}

func (a *Application) handleMaxReconnectAttemptsReached() {
	a.logger.LogWarning(fmt.Sprintf("Max reconnect attempts (%d) reached", config.MaxReconnectAttempts), "AutoReconnect")

	a.isLive = false
	a.isTryingToLogIn = true
	a.reconnectAttempts = 0
	a.stopRefreshTimer()

	a.setStatus(a.localization.GetString("ErrorMaxReconnectAttempts"))
	a.showSessionExpired()
	a.notification.ShowError(a.localization.GetString("ErrorMaxReconnectAttempts"), config.NotificationDurationCritical)

	// Restore window
	a.mainWindow.Show()
	a.mainWindow.RequestFocus()
}

func (a *Application) handleNoCredentialsForReconnect() {
	a.logger.LogWarning("No saved credentials for auto-reconnect", "AutoReconnect")
	a.isLive = false
	a.isTryingToLogIn = true
	a.stopRefreshTimer()

	a.setStatus(a.localization.GetString("ErrorNoSavedCredentials"))
	a.showSessionExpired()
	a.notification.ShowWarning(a.localization.GetString("ErrorNoSavedCredentials"), config.NotificationDurationError)

	a.mainWindow.Show()
	a.mainWindow.RequestFocus()
}

func (a *Application) getLoginErrorMessage(status auth.LoginStatus) string {
	switch status {
	case auth.StatusWrongPassword:
		return a.localization.GetString("ErrorWrongPassword")
	case auth.StatusMaxLimit:
		return a.localization.GetString("ErrorMaxLimit")
	case auth.StatusDataLimit:
		return a.localization.GetString("ErrorDataLimit")
	case auth.StatusNoConnection:
		return a.localization.GetString("ErrorConnectionFailed")
	case auth.StatusLogInAgain:
		return a.localization.GetString("ErrorSessionExpiredMessage")
	default:
		return a.localization.GetString("ErrorConnectionFailed")
	}
}

func (a *Application) applyLocalization() {
	a.mainWindow.SetTitle(a.localization.GetString("WindowTitle"))

	if a.isTryingToLogIn {
		a.loginButton.SetText(a.localization.GetString("ButtonLogin"))
	} else {
		a.loginButton.SetText(a.localization.GetString("ButtonDisconnect"))
	}
	a.logoutButton.SetText(a.localization.GetString("ButtonLogout"))

	a.useHTTPSCheck.Text = a.localization.GetString("CheckboxUseHTTPS")
	a.useHTTPSCheck.Refresh()
	a.skipSSLCheck.Text = a.localization.GetString("CheckboxSkipSSL")
	a.skipSSLCheck.Refresh()
	a.rememberCheck.Text = a.localization.GetString("CheckboxRememberCredentials")
	a.rememberCheck.Refresh()
	a.autoStartCheck.Text = a.localization.GetString("LabelAutoStartWindows")
	a.autoStartCheck.Refresh()

	// Update connection panel if connected
	if !a.isTryingToLogIn && a.userID != "" {
		creds := a.credManager.GetCredentials()
		a.connectionStatus.SetText(a.localization.GetStringf("StatusConnectedTo", creds.ServerIP))
		a.userInfo.SetText(a.localization.GetStringf("StatusLoggedInAsUser", a.userID))
	}
}

func (a *Application) trueApplicationExit() {
	a.logger.LogUIEvent("ApplicationExit", "Shutdown requested")

	a.stopRefreshTimer()
	a.stopSessionTimer()
	a.authService.Close()
	a.logger.LogApplicationStop()

	a.fyneApp.Quit()
}

// Run starts the application main loop
func (a *Application) Run() {
	a.mainWindow.ShowAndRun()
}
