//go:build windows

package autostart

import (
	"golang.org/x/sys/windows/registry"
)

const registryPath = `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`

// Enable adds the application to Windows auto-start via registry
func (m *Manager) Enable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.SetStringValue(m.appName, "\""+m.exePath+"\" /minimized")
}

// Disable removes the application from Windows auto-start
func (m *Manager) Disable() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()

	return key.DeleteValue(m.appName)
}

// IsEnabled checks if auto-start is currently enabled
func (m *Manager) IsEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()

	_, _, err = key.GetStringValue(m.appName)
	return err == nil
}
