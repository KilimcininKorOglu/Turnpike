package config

import (
	"testing"
	"time"
)

// TestValidatePort covers boundary and invalid cases.
func TestValidatePort(t *testing.T) {
	cases := []struct {
		port  int
		valid bool
	}{
		{port: -1, valid: false},
		{port: 0, valid: false},
		{port: 1, valid: true},
		{port: 8090, valid: true},
		{port: 65535, valid: true},
		{port: 65536, valid: false},
	}

	for _, tc := range cases {
		got := ValidatePort(tc.port)
		if got != tc.valid {
			t.Errorf("ValidatePort(%d) = %v, want %v", tc.port, got, tc.valid)
		}
	}
}

// TestConstantValues ensures the ported constants match expected constant values.
func TestConstantValues(t *testing.T) {
	if DefaultServerIP != "172.16.100.2" {
		t.Errorf("DefaultServerIP = %q, want %q", DefaultServerIP, "172.16.100.2")
	}
	if DefaultCaptivePortalPort != 8090 {
		t.Errorf("DefaultCaptivePortalPort = %d, want 8090", DefaultCaptivePortalPort)
	}
	if DefaultUseHTTPS != true {
		t.Error("DefaultUseHTTPS should be true")
	}
	if DefaultSkipSSLValidation != false {
		t.Error("DefaultSkipSSLValidation should be false")
	}

	if MaxReconnectAttempts != 3 {
		t.Errorf("MaxReconnectAttempts = %d, want 3", MaxReconnectAttempts)
	}
	if SessionTimerIntervalSeconds != 1 {
		t.Errorf("SessionTimerIntervalSeconds = %d, want 1", SessionTimerIntervalSeconds)
	}

	if NotificationDurationSuccess != 3000 {
		t.Errorf("NotificationDurationSuccess = %d, want 3000", NotificationDurationSuccess)
	}
	if NotificationDurationInfo != 4000 {
		t.Errorf("NotificationDurationInfo = %d, want 4000", NotificationDurationInfo)
	}
	if NotificationDurationWarning != 4000 {
		t.Errorf("NotificationDurationWarning = %d, want 4000", NotificationDurationWarning)
	}
	if NotificationDurationError != 5000 {
		t.Errorf("NotificationDurationError = %d, want 5000", NotificationDurationError)
	}
	if NotificationDurationCritical != 8000 {
		t.Errorf("NotificationDurationCritical = %d, want 8000", NotificationDurationCritical)
	}

	const expectedMaxLog int64 = 10 * 1024 * 1024
	if MaxLogFileSizeBytes != expectedMaxLog {
		t.Errorf("MaxLogFileSizeBytes = %d, want %d", MaxLogFileSizeBytes, expectedMaxLog)
	}

	if MinPortNumber != 1 {
		t.Errorf("MinPortNumber = %d, want 1", MinPortNumber)
	}
	if MaxPortNumber != 65535 {
		t.Errorf("MaxPortNumber = %d, want 65535", MaxPortNumber)
	}
}

// TestTimeoutValues verifies the duration variables match expected timeout values.
func TestTimeoutValues(t *testing.T) {
	if HTTPRequestTimeout != 30*time.Second {
		t.Errorf("HTTPRequestTimeout = %v, want 30s", HTTPRequestTimeout)
	}
	if RetryDelay != 5*time.Second {
		t.Errorf("RetryDelay = %v, want 5s", RetryDelay)
	}
	if ConnectionRetryDelay != 2*time.Second {
		t.Errorf("ConnectionRetryDelay = %v, want 2s", ConnectionRetryDelay)
	}
}
