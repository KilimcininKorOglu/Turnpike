package ui

import (
	"testing"
)

func TestNewNotificationManager(t *testing.T) {
	nm := NewNotificationManager()
	if nm == nil {
		t.Fatal("expected non-nil NotificationManager")
	}
}

func TestNotificationManager_HideWithoutOverlay(t *testing.T) {
	nm := NewNotificationManager()
	// Should not panic
	nm.Hide()
}

func TestNotificationManager_ShowWithoutOverlay(t *testing.T) {
	nm := NewNotificationManager()
	// Should not panic when overlay is nil
	nm.ShowSuccess("test", 1000)
	nm.ShowError("test", 1000)
	nm.ShowWarning("test", 1000)
	nm.ShowInfo("test", 1000)
}

func TestGetNotificationIcon_AllTypes(t *testing.T) {
	tests := []struct {
		notifType NotificationType
		expected  string
	}{
		{NotificationSuccess, "✓"},
		{NotificationError, "✕"},
		{NotificationWarning, "⚠"},
		{NotificationInfo, "ℹ"},
		{NotificationType(99), "•"},
	}

	for _, tt := range tests {
		icon := GetNotificationIcon(tt.notifType)
		if icon != tt.expected {
			t.Errorf("GetNotificationIcon(%d) = %q, want %q", tt.notifType, icon, tt.expected)
		}
	}
}

func TestGetNotificationBackgroundColor_NotNil(t *testing.T) {
	types := []NotificationType{
		NotificationSuccess,
		NotificationError,
		NotificationWarning,
		NotificationInfo,
		NotificationType(99),
	}

	for _, nt := range types {
		c := GetNotificationBackgroundColor(nt)
		if c == nil {
			t.Errorf("expected non-nil color for notification type %d", nt)
		}
	}
}

func TestNotificationType_Constants(t *testing.T) {
	if NotificationSuccess != 0 {
		t.Errorf("expected NotificationSuccess = 0, got %d", NotificationSuccess)
	}
	if NotificationError != 1 {
		t.Errorf("expected NotificationError = 1, got %d", NotificationError)
	}
	if NotificationWarning != 2 {
		t.Errorf("expected NotificationWarning = 2, got %d", NotificationWarning)
	}
	if NotificationInfo != 3 {
		t.Errorf("expected NotificationInfo = 3, got %d", NotificationInfo)
	}
}
