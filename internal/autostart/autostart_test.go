package autostart

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestNewManager(t *testing.T) {
	m := NewManager("TestApp", "/usr/bin/testapp")
	if m.AppName() != "TestApp" {
		t.Errorf("expected AppName TestApp, got %s", m.AppName())
	}
	if m.ExePath() != "/usr/bin/testapp" {
		t.Errorf("expected ExePath /usr/bin/testapp, got %s", m.ExePath())
	}
}

func TestAutostart_EnableDisableCycle(t *testing.T) {
	// Skip on Windows since it requires registry access
	if runtime.GOOS == "windows" {
		t.Skip("skipping autostart file test on Windows (uses registry)")
	}

	// Create a temporary HOME directory
	tempHome, err := os.MkdirTemp("", "autostart_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHome)

	// Override HOME for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	m := NewManager("Sophos XG User Login", "/tmp/sophos-xg-login")

	// Initially disabled
	if m.IsEnabled() {
		t.Error("expected auto-start to be disabled initially")
	}

	// Enable
	if err := m.Enable(); err != nil {
		t.Fatalf("failed to enable auto-start: %v", err)
	}

	if !m.IsEnabled() {
		t.Error("expected auto-start to be enabled after Enable()")
	}

	// Verify file was created
	var expectedPath string
	switch runtime.GOOS {
	case "darwin":
		expectedPath = filepath.Join(tempHome, "Library", "LaunchAgents", "com.sophos-xg-login.app.plist")
	case "linux":
		expectedPath = filepath.Join(tempHome, ".config", "autostart", "sophos-xg-login.desktop")
	}
	if expectedPath != "" {
		if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
			t.Errorf("expected autostart file at %s", expectedPath)
		}
	}

	// Disable
	if err := m.Disable(); err != nil {
		t.Fatalf("failed to disable auto-start: %v", err)
	}

	if m.IsEnabled() {
		t.Error("expected auto-start to be disabled after Disable()")
	}
}

func TestAutostart_DisableNonExistent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}

	tempHome, err := os.MkdirTemp("", "autostart_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHome)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	m := NewManager("TestApp", "/tmp/testapp")

	// Should not error when disabling non-existent autostart
	if err := m.Disable(); err != nil {
		t.Errorf("unexpected error disabling non-existent autostart: %v", err)
	}
}

func TestAutostart_EnableCreatesDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on Windows")
	}

	tempHome, err := os.MkdirTemp("", "autostart_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempHome)

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempHome)
	defer os.Setenv("HOME", originalHome)

	m := NewManager("TestApp", "/tmp/testapp")

	// Enable should create the autostart directory if it doesn't exist
	if err := m.Enable(); err != nil {
		t.Fatalf("failed to enable: %v", err)
	}

	if !m.IsEnabled() {
		t.Error("expected enabled after Enable()")
	}
}
