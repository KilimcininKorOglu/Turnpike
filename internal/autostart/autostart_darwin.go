//go:build darwin

package autostart

import (
	"fmt"
	"os"
	"path/filepath"
)

// Enable creates a launchd plist for auto-start on macOS
func (m *Manager) Enable() error {
	plistDir := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents")
	if err := os.MkdirAll(plistDir, 0755); err != nil {
		return err
	}

	plistPath := filepath.Join(plistDir, m.plistName())

	plist := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>%s</string>
    <key>ProgramArguments</key>
    <array>
        <string>%s</string>
        <string>/minimized</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
</dict>
</plist>`, m.bundleID(), m.exePath)

	return os.WriteFile(plistPath, []byte(plist), 0644)
}

// Disable removes the launchd plist
func (m *Manager) Disable() error {
	plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", m.plistName())
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(plistPath)
}

// IsEnabled checks if the launchd plist exists
func (m *Manager) IsEnabled() bool {
	plistPath := filepath.Join(os.Getenv("HOME"), "Library", "LaunchAgents", m.plistName())
	_, err := os.Stat(plistPath)
	return err == nil
}

func (m *Manager) bundleID() string {
	return "com.sophos-xg-login.app"
}

func (m *Manager) plistName() string {
	return m.bundleID() + ".plist"
}
