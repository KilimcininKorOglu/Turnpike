package autostart

// Manager handles auto-start configuration for the application
type Manager struct {
	appName string
	exePath string
}

// NewManager creates a new auto-start manager
func NewManager(appName, exePath string) *Manager {
	return &Manager{
		appName: appName,
		exePath: exePath,
	}
}

// AppName returns the application name
func (m *Manager) AppName() string {
	return m.appName
}

// ExePath returns the executable path
func (m *Manager) ExePath() string {
	return m.exePath
}
