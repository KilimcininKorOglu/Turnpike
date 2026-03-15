package logging

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

const (
	logFileName      = "sophos-xg-login.log"
	maxLogFileSize   int64 = 10 * 1024 * 1024 // 10MB
	maxMessageLength = 200
	maxValueLength   = 50
)

// Level represents log severity.
type Level string

const (
	LevelDebug Level = "DEBUG"
	LevelInfo  Level = "INFO"
	LevelWarn  Level = "WARN"
	LevelError Level = "ERROR"
)

// Service provides centralized, thread-safe logging with automatic file rotation.
type Service struct {
	mu          sync.Mutex
	logDir      string
	logFilePath string
	debug       bool
}

// NewService creates a new logging service that writes to logDir/sophos-xg-login.log.
// When debug is false, LogDebug calls are silently discarded.
func NewService(logDir string, debug bool) *Service {
	s := &Service{
		logDir:      logDir,
		logFilePath: filepath.Join(logDir, logFileName),
		debug:       debug,
	}

	// Best-effort: ensure log directory exists.
	_ = os.MkdirAll(logDir, 0755)

	return s
}

// LogApplicationStart logs application startup information.
func (s *Service) LogApplicationStart(version string) {
	message := "=== Sophos XG User Login Application Started ===" +
		" | Version: " + version +
		" | Go: " + runtime.Version() +
		" | OS: " + runtime.GOOS + "/" + runtime.GOARCH
	s.logMessage(LevelInfo, message, "Application")
}

// LogApplicationStop logs application shutdown.
func (s *Service) LogApplicationStop() {
	s.logMessage(LevelInfo, "=== Sophos XG User Login Application Stopped ===", "Application")
}

// LogAuthenticationAttempt logs a login attempt and its result.
func (s *Service) LogAuthenticationAttempt(username, serverIP string, success bool, message string) {
	status := "Login successful"
	if !success {
		status = "Login failed"
	}

	logMessage := status + " | User: " + username + " | Server: " + serverIP
	if !success && message != "" {
		logMessage += " | Reason: " + message
	}

	level := LevelInfo
	if !success {
		level = LevelWarn
	}
	s.logMessage(level, logMessage, "Authentication")
}

// LogLogout logs a logout event and its result.
func (s *Service) LogLogout(username, serverIP string, success bool, message string) {
	status := "Logout successful"
	if !success {
		status = "Logout failed"
	}

	logMessage := status + " | User: " + username + " | Server: " + serverIP
	if !success && message != "" {
		logMessage += " | Reason: " + message
	}

	level := LevelInfo
	if !success {
		level = LevelWarn
	}
	s.logMessage(level, logMessage, "Authentication")
}

// LogSessionCheck logs a session monitoring event.
// duration is optional; pass nil when the session duration is unknown.
func (s *Service) LogSessionCheck(username string, isActive bool, duration *time.Duration) {
	status := "Session active"
	if !isActive {
		status = "Session expired"
	}

	logMessage := status + " | User: " + username
	if duration != nil {
		h := int(duration.Hours())
		m := int(duration.Minutes()) % 60
		sec := int(duration.Seconds()) % 60
		logMessage += fmt.Sprintf(" | Duration: %02d:%02d:%02d", h, m, sec)
	}

	level := LevelDebug
	if !isActive {
		level = LevelInfo
	}
	s.logMessage(level, logMessage, "Authentication")
}

// LogNetworkRequest logs an outgoing HTTP request.
// statusCode and duration are optional; pass nil when not available.
func (s *Service) LogNetworkRequest(method, rawURL string, success bool, statusCode *int, duration *time.Duration) {
	logMessage := "HTTP " + method + " to " + sanitizeURL(rawURL) + " " +
		func() string {
			if success {
				return "succeeded"
			}
			return "failed"
		}()

	if statusCode != nil {
		logMessage += fmt.Sprintf(" | Status: %d", *statusCode)
	}
	if duration != nil {
		logMessage += fmt.Sprintf(" | Duration: %.3fms", float64(duration.Nanoseconds())/1e6)
	}

	level := LevelDebug
	if !success {
		level = LevelError
	}
	s.logMessage(level, logMessage, "Network")
}

// LogSecurityEvent logs a security-related event.
// username is optional; pass an empty string when not applicable.
func (s *Service) LogSecurityEvent(eventType, message, username string) {
	logMessage := "Security Event [" + eventType + "]: " + message
	if username != "" {
		logMessage += " | User: " + username
	}
	s.logMessage(LevelInfo, logMessage, "Security")
}

// LogPasswordOperation logs a password encryption or decryption operation.
// username is optional; pass an empty string when not applicable.
func (s *Service) LogPasswordOperation(operation string, success bool, username string) {
	logMessage := "Password " + operation + " " + func() string {
		if success {
			return "successful"
		}
		return "failed"
	}()
	if username != "" {
		logMessage += " | User: " + username
	}

	level := LevelDebug
	if !success {
		level = LevelError
	}
	s.logMessage(level, logMessage, "Security")
}

// LogUIEvent logs a UI interaction event.
// details is optional; pass an empty string when not applicable.
func (s *Service) LogUIEvent(eventType, details string) {
	logMessage := "UI Event [" + eventType + "]"
	if details != "" {
		logMessage += ": " + details
	}
	s.logMessage(LevelDebug, logMessage, "UI")
}

// LogNotification logs that a toast notification was displayed.
func (s *Service) LogNotification(notifType, message string, duration int) {
	logMessage := "Notification shown | Type: " + notifType +
		fmt.Sprintf(" | Duration: %dms", duration) +
		" | Message: " + sanitizeMessage(message)
	s.logMessage(LevelDebug, logMessage, "UI")
}

// LogConfigurationChange logs a settings change with before/after values.
func (s *Service) LogConfigurationChange(setting, oldValue, newValue string) {
	logMessage := "Configuration changed | Setting: " + setting +
		" | From: " + sanitizeValue(oldValue) +
		" | To: " + sanitizeValue(newValue)
	s.logMessage(LevelInfo, logMessage, "Configuration")
}

// LogError logs an error with an optional message and context label.
func (s *Service) LogError(err error, message, context string) {
	ctx := context
	if ctx == "" {
		ctx = "Application"
	}

	logMessage := "Error in " + ctx
	if message != "" {
		logMessage += ": " + message
	}
	logMessage += " | Exception: " + err.Error()

	s.logMessage(LevelError, logMessage, ctx)
}

// LogWarning logs a warning with an optional context label.
func (s *Service) LogWarning(message, context string) {
	ctx := context
	if ctx == "" {
		ctx = "Application"
	}
	s.logMessage(LevelWarn, message, ctx)
}

// LogInfo logs an informational message with an optional context label.
func (s *Service) LogInfo(message, context string) {
	ctx := context
	if ctx == "" {
		ctx = "Application"
	}
	s.logMessage(LevelInfo, message, ctx)
}

// LogDebug logs a debug message. The message is discarded when debug mode is disabled.
func (s *Service) LogDebug(message, context string) {
	if !s.debug {
		return
	}
	ctx := context
	if ctx == "" {
		ctx = "Application"
	}
	s.logMessage(LevelDebug, message, ctx)
}

// logMessage is the core logging method.
// Format: "2024-01-15 10:30:45.123 INFO  [Category] message\n"
func (s *Service) logMessage(level Level, message, category string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := string(level)
	// Pad level to 5 characters to match the C# PadRight(5) behaviour.
	for len(levelStr) < 5 {
		levelStr += " "
	}
	formatted := timestamp + " " + levelStr + " [" + category + "] " + message + "\n"

	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		// File logging unavailable; continue silently.
		return
	}
	_, _ = f.WriteString(formatted)
	_ = f.Close()

	// Rotate after write so the current entry is always preserved.
	if info, statErr := os.Stat(s.logFilePath); statErr == nil {
		if info.Size() > maxLogFileSize {
			s.rotateLogFile()
		}
	}
}

// rotateLogFile renames the current log file to .old and lets the next write create a fresh file.
// Must be called with s.mu already held.
func (s *Service) rotateLogFile() {
	backupPath := s.logFilePath + ".old"

	// Remove any previous backup.
	_ = os.Remove(backupPath)

	if err := os.Rename(s.logFilePath, backupPath); err != nil {
		// Rename failed; last-resort: delete the active log so it can be re-created.
		_ = os.Remove(s.logFilePath)
	}
}

// sanitizeURL strips query parameters and fragment from a URL to avoid logging credentials.
// Returns "Unknown" for empty input and "[URL parsing failed]" for invalid URLs.
func sanitizeURL(rawURL string) string {
	if strings.TrimSpace(rawURL) == "" {
		return "Unknown"
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "[URL parsing failed]"
	}

	// Reconstruct without query string or fragment, mirroring the C# implementation.
	port := parsed.Port()
	host := parsed.Hostname()
	scheme := parsed.Scheme
	path := parsed.Path

	if port != "" {
		return scheme + "://" + host + ":" + port + path
	}
	return scheme + "://" + host + path
}

// sanitizeMessage truncates messages that exceed maxMessageLength characters.
// Returns "Empty" for empty input.
func sanitizeMessage(message string) string {
	if message == "" {
		return "Empty"
	}
	if len(message) > maxMessageLength {
		return message[:197] + "..."
	}
	return message
}

// sanitizeValue rejects values longer than maxValueLength to avoid logging sensitive data.
// Returns "Empty" for empty input.
func sanitizeValue(value string) string {
	if value == "" {
		return "Empty"
	}
	if len(value) > maxValueLength {
		return "[Value too long - potentially sensitive]"
	}
	return value
}
