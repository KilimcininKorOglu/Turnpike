//go:build linux

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

// Enable creates a .desktop file for auto-start on Linux (XDG standard)
func (m *Manager) Enable() error {
	autostartDir := filepath.Join(os.Getenv("HOME"), ".config", "autostart")
	if err := os.MkdirAll(autostartDir, 0755); err != nil {
		return err
	}

	desktopPath := filepath.Join(autostartDir, m.desktopFileName())

	content := fmt.Sprintf(`[Desktop Entry]
Type=Application
Name=%s
Exec=%s /minimized
Hidden=false
NoDisplay=false
X-GNOME-Autostart-enabled=true
Comment=Turnpike Firewall Authentication
`, m.appName, m.exePath)

	return os.WriteFile(desktopPath, []byte(content), 0644)
}

// Disable removes the .desktop file
func (m *Manager) Disable() error {
	desktopPath := filepath.Join(os.Getenv("HOME"), ".config", "autostart", m.desktopFileName())
	if _, err := os.Stat(desktopPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(desktopPath)
}

// IsEnabled checks if the .desktop file exists
func (m *Manager) IsEnabled() bool {
	desktopPath := filepath.Join(os.Getenv("HOME"), ".config", "autostart", m.desktopFileName())
	_, err := os.Stat(desktopPath)
	return err == nil
}

func (m *Manager) desktopFileName() string {
	return "turnpike.desktop"
}
