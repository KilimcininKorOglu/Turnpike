package ui

import (
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

// NotificationType represents the type of notification
type NotificationType int

const (
	NotificationSuccess NotificationType = iota
	NotificationError
	NotificationWarning
	NotificationInfo
)

// NotificationManager handles in-app toast notifications
type NotificationManager struct {
	container *fyne.Container
	overlay   *fyne.Container
	timer     *time.Timer
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager() *NotificationManager {
	return &NotificationManager{}
}

// SetOverlay sets the parent overlay container for notifications
func (nm *NotificationManager) SetOverlay(overlay *fyne.Container) {
	nm.overlay = overlay
}

// ShowSuccess shows a green success notification
func (nm *NotificationManager) ShowSuccess(message string, durationMs int) {
	nm.showNotification(message, NotificationSuccess, durationMs)
}

// ShowError shows a red error notification
func (nm *NotificationManager) ShowError(message string, durationMs int) {
	nm.showNotification(message, NotificationError, durationMs)
}

// ShowWarning shows an orange warning notification
func (nm *NotificationManager) ShowWarning(message string, durationMs int) {
	nm.showNotification(message, NotificationWarning, durationMs)
}

// ShowInfo shows a blue info notification
func (nm *NotificationManager) ShowInfo(message string, durationMs int) {
	nm.showNotification(message, NotificationInfo, durationMs)
}

func (nm *NotificationManager) showNotification(message string, notifType NotificationType, durationMs int) {
	if nm.overlay == nil {
		return
	}

	// Cancel any existing timer
	if nm.timer != nil {
		nm.timer.Stop()
	}

	icon := getNotificationIcon(notifType)

	bg := getNotificationColor(notifType)
	bg.CornerRadius = 8

	iconLabel := widget.NewLabel(icon)
	iconLabel.TextStyle = fyne.TextStyle{Bold: true}

	msgLabel := widget.NewLabel(message)
	msgLabel.Wrapping = fyne.TextWrapWord

	content := container.NewHBox(iconLabel, msgLabel)
	nm.container = container.NewStack(bg, container.NewPadded(content))

	// Show the notification
	nm.overlay.Objects = []fyne.CanvasObject{
		layout.NewSpacer(),
		container.NewCenter(nm.container),
	}
	nm.overlay.Refresh()

	// Auto-hide after duration
	nm.timer = time.AfterFunc(time.Duration(durationMs)*time.Millisecond, func() {
		nm.Hide()
	})
}

// Hide hides the current notification
func (nm *NotificationManager) Hide() {
	if nm.overlay != nil {
		nm.overlay.Objects = nil
		nm.overlay.Refresh()
	}
}

func getNotificationColor(notifType NotificationType) *canvas.Rectangle {
	var c *canvas.Rectangle
	switch notifType {
	case NotificationSuccess:
		c = canvas.NewRectangle(ColorSuccess)
	case NotificationError:
		c = canvas.NewRectangle(ColorError)
	case NotificationWarning:
		c = canvas.NewRectangle(ColorWarning)
	case NotificationInfo:
		c = canvas.NewRectangle(ColorInfo)
	default:
		c = canvas.NewRectangle(ColorBorder)
	}
	return c
}

// GetNotificationBackgroundColor returns the background color for a notification type (exported for testing)
func GetNotificationBackgroundColor(notifType NotificationType) fyne.CanvasObject {
	return getNotificationColor(notifType)
}

func getNotificationIcon(notifType NotificationType) string {
	switch notifType {
	case NotificationSuccess:
		return "✓"
	case NotificationError:
		return "✕"
	case NotificationWarning:
		return "⚠"
	case NotificationInfo:
		return "ℹ"
	default:
		return "•"
	}
}

// GetNotificationIcon returns the icon text for a notification type (exported for testing)
func GetNotificationIcon(notifType NotificationType) string {
	return getNotificationIcon(notifType)
}
