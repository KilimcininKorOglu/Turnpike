package config

import "time"

const (
	// Default server configuration
	DefaultServerIP          = "172.16.100.2"
	DefaultCaptivePortalPort = 8090
	DefaultUseHTTPS          = true
	DefaultSkipSSLValidation = false

	// Auto-reconnection
	MaxReconnectAttempts = 3

	// UI timing
	SessionTimerIntervalSeconds = 1

	// Notification durations (milliseconds)
	NotificationDurationSuccess  = 3000
	NotificationDurationInfo     = 4000
	NotificationDurationWarning  = 4000
	NotificationDurationError    = 5000
	NotificationDurationCritical = 8000

	// Log file
	MaxLogFileSizeBytes int64 = 10 * 1024 * 1024 // 10 MB

	// Port validation
	MinPortNumber = 1
	MaxPortNumber = 65535
)

// Network timeouts are declared as variables because time.Duration
// multiplications (e.g. 30 * time.Second) are not constant expressions in Go.
var (
	HTTPRequestTimeout   = 30 * time.Second
	RetryDelay           = 5 * time.Second
	ConnectionRetryDelay = 2 * time.Second
)

// ValidatePort reports whether port is within the valid TCP/UDP port range
// [MinPortNumber, MaxPortNumber] (1-65535).
func ValidatePort(port int) bool {
	return port >= MinPortNumber && port <= MaxPortNumber
}
