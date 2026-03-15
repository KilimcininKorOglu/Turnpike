package logging

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// newTestService returns a Service wired to a temporary directory.
// The caller is responsible for cleaning up the returned directory via t.Cleanup.
func newTestService(t *testing.T, debug bool) (*Service, string) {
	t.Helper()
	dir := t.TempDir()
	return NewService(dir, debug), dir
}

// readLog reads the entire contents of the active log file.
func readLog(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatalf("readLog: %v", err)
	}
	return string(data)
}

// logFileExists reports whether the active log file is present.
func logFileExists(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, logFileName))
	return err == nil
}

// ── NewService ────────────────────────────────────────────────────────────────

func TestNewService_CorrectPaths(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir, false)

	if svc.logDir != dir {
		t.Errorf("logDir: got %q, want %q", svc.logDir, dir)
	}
	wantPath := filepath.Join(dir, logFileName)
	if svc.logFilePath != wantPath {
		t.Errorf("logFilePath: got %q, want %q", svc.logFilePath, wantPath)
	}
}

func TestNewService_CreatesDirectory(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "nested", "logdir")
	NewService(dir, false)

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Errorf("expected log directory %q to be created", dir)
	}
}

// ── logMessage / file writing ─────────────────────────────────────────────────

func TestLogMessage_WritesToFile(t *testing.T) {
	svc, dir := newTestService(t, true)

	svc.logMessage(LevelInfo, "hello world", "Test")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "hello world") {
		t.Errorf("log file does not contain expected message; got:\n%s", contents)
	}
}

func TestLogMessage_Format(t *testing.T) {
	svc, dir := newTestService(t, true)

	svc.logMessage(LevelInfo, "test message", "Category")

	line := readLog(t, dir)
	// Timestamp prefix: "YYYY-MM-DD HH:MM:SS.mmm"
	if len(line) < 24 {
		t.Fatalf("log line too short: %q", line)
	}
	if !strings.Contains(line, "INFO ") {
		t.Errorf("expected padded level 'INFO '; got: %q", line)
	}
	if !strings.Contains(line, "[Category]") {
		t.Errorf("expected category '[Category]'; got: %q", line)
	}
	if !strings.Contains(line, "test message") {
		t.Errorf("expected 'test message'; got: %q", line)
	}
}

func TestLogMessage_AllLevels(t *testing.T) {
	cases := []struct {
		level Level
		want  string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN "},
		{LevelError, "ERROR"},
	}

	for _, tc := range cases {
		t.Run(string(tc.level), func(t *testing.T) {
			svc, dir := newTestService(t, true)
			svc.logMessage(tc.level, "msg", "Test")
			if !strings.Contains(readLog(t, dir), tc.want) {
				t.Errorf("level %q not found in log", tc.want)
			}
		})
	}
}

// ── Log rotation ──────────────────────────────────────────────────────────────

func TestRotateLogFile_CreatesOldBackup(t *testing.T) {
	svc, dir := newTestService(t, true)

	// Write a dummy log file that exceeds the max size.
	logPath := filepath.Join(dir, logFileName)
	bigData := make([]byte, int(maxLogFileSize)+1)
	if err := os.WriteFile(logPath, bigData, 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	svc.mu.Lock()
	svc.rotateLogFile()
	svc.mu.Unlock()

	backupPath := logPath + ".old"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("expected .old backup file to exist after rotation")
	}
	if logFileExists(dir) {
		t.Error("expected active log file to be gone after rotation")
	}
}

func TestRotateLogFile_ReplacesExistingBackup(t *testing.T) {
	svc, dir := newTestService(t, true)

	logPath := filepath.Join(dir, logFileName)
	oldPath := logPath + ".old"

	if err := os.WriteFile(logPath, []byte("new"), 0644); err != nil {
		t.Fatalf("setup log: %v", err)
	}
	if err := os.WriteFile(oldPath, []byte("old backup"), 0644); err != nil {
		t.Fatalf("setup old: %v", err)
	}

	svc.mu.Lock()
	svc.rotateLogFile()
	svc.mu.Unlock()

	data, err := os.ReadFile(oldPath)
	if err != nil {
		t.Fatalf("read backup: %v", err)
	}
	if string(data) != "new" {
		t.Errorf("backup should contain old log content 'new'; got %q", string(data))
	}
}

func TestLogMessage_TriggersRotationWhenOversize(t *testing.T) {
	svc, dir := newTestService(t, true)

	// Pre-fill the log file so the next write tips it over the limit.
	logPath := filepath.Join(dir, logFileName)
	bigData := make([]byte, int(maxLogFileSize))
	if err := os.WriteFile(logPath, bigData, 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	svc.logMessage(LevelInfo, "trigger rotation", "Test")

	backupPath := logPath + ".old"
	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		t.Error("expected .old backup to exist after size-triggered rotation")
	}
}

// ── sanitizeURL ───────────────────────────────────────────────────────────────

func TestSanitizeURL_StripsQueryParameters(t *testing.T) {
	result := sanitizeURL("http://192.168.1.1:8090/login?username=admin&password=secret")
	if strings.Contains(result, "username") || strings.Contains(result, "password") || strings.Contains(result, "secret") {
		t.Errorf("sanitizeURL should strip query params; got: %q", result)
	}
	if !strings.Contains(result, "/login") {
		t.Errorf("sanitizeURL should preserve path; got: %q", result)
	}
}

func TestSanitizeURL_PreservesSchemeHostPortPath(t *testing.T) {
	result := sanitizeURL("https://fw.example.com:8090/api/v1/auth?token=xyz")
	want := "https://fw.example.com:8090/api/v1/auth"
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestSanitizeURL_EmptyInput(t *testing.T) {
	if got := sanitizeURL(""); got != "Unknown" {
		t.Errorf("empty URL: got %q, want %q", got, "Unknown")
	}
	if got := sanitizeURL("   "); got != "Unknown" {
		t.Errorf("whitespace URL: got %q, want %q", got, "Unknown")
	}
}

func TestSanitizeURL_InvalidURL(t *testing.T) {
	result := sanitizeURL("://not a valid url [[[")
	if result != "[URL parsing failed]" {
		t.Errorf("invalid URL: got %q, want %q", result, "[URL parsing failed]")
	}
}

func TestSanitizeURL_NoQueryString(t *testing.T) {
	input := "http://192.168.0.1:8090/login"
	result := sanitizeURL(input)
	if result != input {
		t.Errorf("URL without query changed: got %q, want %q", result, input)
	}
}

// ── sanitizeMessage ───────────────────────────────────────────────────────────

func TestSanitizeMessage_TruncatesLongMessage(t *testing.T) {
	long := strings.Repeat("a", 300)
	result := sanitizeMessage(long)
	if len(result) > maxMessageLength {
		t.Errorf("truncated message length %d exceeds max %d", len(result), maxMessageLength)
	}
	if !strings.HasSuffix(result, "...") {
		t.Errorf("truncated message should end with '...'; got: %q", result)
	}
}

func TestSanitizeMessage_PassesShortMessage(t *testing.T) {
	short := "a short message"
	if got := sanitizeMessage(short); got != short {
		t.Errorf("short message changed: got %q, want %q", got, short)
	}
}

func TestSanitizeMessage_ExactlyMaxLength(t *testing.T) {
	exact := strings.Repeat("x", maxMessageLength)
	if got := sanitizeMessage(exact); got != exact {
		t.Errorf("message at exactly max length should not be truncated")
	}
}

func TestSanitizeMessage_EmptyInput(t *testing.T) {
	if got := sanitizeMessage(""); got != "Empty" {
		t.Errorf("empty message: got %q, want %q", got, "Empty")
	}
}

// ── sanitizeValue ─────────────────────────────────────────────────────────────

func TestSanitizeValue_MarksLongValueAsSensitive(t *testing.T) {
	long := strings.Repeat("s", maxValueLength+1)
	result := sanitizeValue(long)
	if result != "[Value too long - potentially sensitive]" {
		t.Errorf("long value: got %q", result)
	}
}

func TestSanitizeValue_PassesShortValue(t *testing.T) {
	short := "shortval"
	if got := sanitizeValue(short); got != short {
		t.Errorf("short value changed: got %q, want %q", got, short)
	}
}

func TestSanitizeValue_ExactlyMaxLength(t *testing.T) {
	exact := strings.Repeat("v", maxValueLength)
	if got := sanitizeValue(exact); got != exact {
		t.Errorf("value at exactly max length should pass through unchanged")
	}
}

func TestSanitizeValue_EmptyInput(t *testing.T) {
	if got := sanitizeValue(""); got != "Empty" {
		t.Errorf("empty value: got %q, want %q", got, "Empty")
	}
}

// ── LogDebug ──────────────────────────────────────────────────────────────────

func TestLogDebug_WritesWhenDebugEnabled(t *testing.T) {
	svc, dir := newTestService(t, true)
	svc.LogDebug("debug message", "Test")

	if !strings.Contains(readLog(t, dir), "debug message") {
		t.Error("expected debug message to be written when debug mode is enabled")
	}
}

func TestLogDebug_DoesNotWriteWhenDebugDisabled(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogDebug("should not appear", "Test")

	if logFileExists(dir) {
		contents := readLog(t, dir)
		if strings.Contains(contents, "should not appear") {
			t.Error("debug message should not be written when debug mode is disabled")
		}
	}
	// If no log file was created that is also acceptable (nothing was logged).
}

// ── Thread safety ─────────────────────────────────────────────────────────────

func TestConcurrentLogging_NoCorruption(t *testing.T) {
	svc, dir := newTestService(t, true)

	const goroutines = 50
	const messagesEach = 20

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < messagesEach; j++ {
				svc.LogInfo(fmt.Sprintf("goroutine %d message %d", id, j), "ConcurrencyTest")
			}
		}(i)
	}
	wg.Wait()

	contents := readLog(t, dir)
	lines := strings.Split(strings.TrimRight(contents, "\n"), "\n")
	// Every line should be a well-formed log entry (starts with a digit = year).
	for _, line := range lines {
		if line == "" {
			continue
		}
		if len(line) < 1 || (line[0] < '0' || line[0] > '9') {
			t.Errorf("malformed log line (possible corruption): %q", line)
		}
	}
	total := goroutines * messagesEach
	if len(lines) < total {
		t.Errorf("expected at least %d log lines, got %d", total, len(lines))
	}
}

// ── LogAuthenticationAttempt ──────────────────────────────────────────────────

func TestLogAuthenticationAttempt_Success(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogAuthenticationAttempt("alice", "10.0.0.1", true, "")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "Login successful") {
		t.Error("expected 'Login successful'")
	}
	if !strings.Contains(contents, "User: alice") {
		t.Error("expected 'User: alice'")
	}
	if !strings.Contains(contents, "Server: 10.0.0.1") {
		t.Error("expected 'Server: 10.0.0.1'")
	}
	if !strings.Contains(contents, "INFO") {
		t.Error("expected INFO level for successful login")
	}
}

func TestLogAuthenticationAttempt_Failure(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogAuthenticationAttempt("bob", "10.0.0.2", false, "invalid credentials")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "Login failed") {
		t.Error("expected 'Login failed'")
	}
	if !strings.Contains(contents, "Reason: invalid credentials") {
		t.Error("expected reason in log")
	}
	if !strings.Contains(contents, "WARN") {
		t.Error("expected WARN level for failed login")
	}
}

func TestLogAuthenticationAttempt_FailureNoMessage(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogAuthenticationAttempt("bob", "10.0.0.2", false, "")

	contents := readLog(t, dir)
	if strings.Contains(contents, "Reason:") {
		t.Error("should not include Reason field when message is empty")
	}
}

// ── LogNetworkRequest ─────────────────────────────────────────────────────────

func TestLogNetworkRequest_WithDurationAndStatus(t *testing.T) {
	svc, dir := newTestService(t, true)

	code := 200
	d := 123 * time.Millisecond
	svc.LogNetworkRequest("POST", "http://fw.local:8090/login", true, &code, &d)

	contents := readLog(t, dir)
	if !strings.Contains(contents, "Status: 200") {
		t.Error("expected status code in log")
	}
	if !strings.Contains(contents, "ms") {
		t.Error("expected duration in log")
	}
	if !strings.Contains(contents, "HTTP POST") {
		t.Error("expected method in log")
	}
}

func TestLogNetworkRequest_WithoutOptionalFields(t *testing.T) {
	svc, dir := newTestService(t, true)
	svc.LogNetworkRequest("GET", "http://fw.local:8090/check", true, nil, nil)

	contents := readLog(t, dir)
	if strings.Contains(contents, "Status:") {
		t.Error("should not log Status when statusCode is nil")
	}
	if strings.Contains(contents, "Duration:") {
		t.Error("should not log Duration when duration is nil")
	}
}

func TestLogNetworkRequest_FailureUsesErrorLevel(t *testing.T) {
	svc, dir := newTestService(t, true)
	svc.LogNetworkRequest("POST", "http://fw.local:8090/login", false, nil, nil)

	if !strings.Contains(readLog(t, dir), "ERROR") {
		t.Error("expected ERROR level for failed network request")
	}
}

func TestLogNetworkRequest_SuccessUsesDebugLevel(t *testing.T) {
	svc, dir := newTestService(t, true)
	svc.LogNetworkRequest("GET", "http://fw.local:8090/check", true, nil, nil)

	if !strings.Contains(readLog(t, dir), "DEBUG") {
		t.Error("expected DEBUG level for successful network request")
	}
}

// ── All log methods – level and category smoke tests ─────────────────────────

func TestLogInfo_LevelAndCategory(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogInfo("info msg", "MyCtx")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "INFO") {
		t.Error("expected INFO level")
	}
	if !strings.Contains(contents, "[MyCtx]") {
		t.Error("expected [MyCtx] category")
	}
}

func TestLogWarning_LevelAndCategory(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogWarning("warn msg", "WarnCtx")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "WARN") {
		t.Error("expected WARN level")
	}
	if !strings.Contains(contents, "[WarnCtx]") {
		t.Error("expected [WarnCtx] category")
	}
}

func TestLogError_LevelAndContext(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogError(errors.New("something broke"), "context msg", "ErrCtx")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "ERROR") {
		t.Error("expected ERROR level")
	}
	if !strings.Contains(contents, "something broke") {
		t.Error("expected error text in log")
	}
	if !strings.Contains(contents, "[ErrCtx]") {
		t.Error("expected [ErrCtx] category")
	}
}

func TestLogSecurityEvent_LevelAndCategory(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogSecurityEvent("ENCRYPT", "key rotated", "alice")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "INFO") {
		t.Error("expected INFO level")
	}
	if !strings.Contains(contents, "[Security]") {
		t.Error("expected [Security] category")
	}
	if !strings.Contains(contents, "ENCRYPT") {
		t.Error("expected event type in log")
	}
}

func TestLogPasswordOperation_SuccessIsDebug(t *testing.T) {
	svc, dir := newTestService(t, true)
	svc.LogPasswordOperation("encryption", true, "alice")

	if !strings.Contains(readLog(t, dir), "DEBUG") {
		t.Error("expected DEBUG level for successful password operation")
	}
}

func TestLogPasswordOperation_FailureIsError(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogPasswordOperation("decryption", false, "bob")

	if !strings.Contains(readLog(t, dir), "ERROR") {
		t.Error("expected ERROR level for failed password operation")
	}
}

func TestLogUIEvent_LevelAndCategory(t *testing.T) {
	svc, dir := newTestService(t, true)
	svc.LogUIEvent("ButtonClick", "login button")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "DEBUG") {
		t.Error("expected DEBUG level for UI event")
	}
	if !strings.Contains(contents, "[UI]") {
		t.Error("expected [UI] category")
	}
}

func TestLogNotification_LevelAndCategory(t *testing.T) {
	svc, dir := newTestService(t, true)
	svc.LogNotification("Success", "Login ok", 3000)

	contents := readLog(t, dir)
	if !strings.Contains(contents, "DEBUG") {
		t.Error("expected DEBUG level for notification")
	}
	if !strings.Contains(contents, "[UI]") {
		t.Error("expected [UI] category")
	}
}

func TestLogConfigurationChange_LevelAndCategory(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogConfigurationChange("AutoReconnect", "false", "true")

	contents := readLog(t, dir)
	if !strings.Contains(contents, "INFO") {
		t.Error("expected INFO level")
	}
	if !strings.Contains(contents, "[Configuration]") {
		t.Error("expected [Configuration] category")
	}
}

func TestLogLogout_SuccessInfo_FailureWarn(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogLogout("alice", "10.0.0.1", true, "")

	if !strings.Contains(readLog(t, dir), "INFO") {
		t.Error("expected INFO for successful logout")
	}

	svc2, dir2 := newTestService(t, false)
	svc2.LogLogout("bob", "10.0.0.2", false, "timeout")

	contents2 := readLog(t, dir2)
	if !strings.Contains(contents2, "WARN") {
		t.Error("expected WARN for failed logout")
	}
	if !strings.Contains(contents2, "Reason: timeout") {
		t.Error("expected reason in failed logout")
	}
}

func TestLogSessionCheck_ActiveIsDebug_ExpiredIsInfo(t *testing.T) {
	svc, dir := newTestService(t, true)
	svc.LogSessionCheck("alice", true, nil)

	if !strings.Contains(readLog(t, dir), "DEBUG") {
		t.Error("expected DEBUG for active session check")
	}

	svc2, dir2 := newTestService(t, true)
	svc2.LogSessionCheck("alice", false, nil)

	if !strings.Contains(readLog(t, dir2), "INFO") {
		t.Error("expected INFO for expired session check")
	}
}

func TestLogSessionCheck_IncludesDuration(t *testing.T) {
	svc, dir := newTestService(t, true)
	d := 90*time.Minute + 5*time.Second
	svc.LogSessionCheck("alice", false, &d)

	if !strings.Contains(readLog(t, dir), "Duration: 01:30:05") {
		t.Errorf("expected formatted duration in session check log; got:\n%s", readLog(t, dir))
	}
}

func TestLogApplicationStart_ContainsVersion(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogApplicationStart("test-version")

	if !strings.Contains(readLog(t, dir), "test-version") {
		t.Error("expected version in application start log")
	}
}

func TestLogApplicationStop_WritesEntry(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogApplicationStop()

	if !strings.Contains(readLog(t, dir), "Application Stopped") {
		t.Error("expected stop message in log")
	}
}

// ── Default context fallback ──────────────────────────────────────────────────

func TestLogInfo_EmptyContextDefaultsToApplication(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogInfo("some message", "")

	if !strings.Contains(readLog(t, dir), "[Application]") {
		t.Error("expected [Application] as default category when context is empty")
	}
}

func TestLogError_EmptyContextDefaultsToApplication(t *testing.T) {
	svc, dir := newTestService(t, false)
	svc.LogError(errors.New("err"), "", "")

	if !strings.Contains(readLog(t, dir), "[Application]") {
		t.Error("expected [Application] as default category when context is empty")
	}
}
