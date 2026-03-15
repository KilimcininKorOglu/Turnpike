package auth

import "testing"

func TestParseCaptivePortalResponse_Empty(t *testing.T) {
	result := ParseCaptivePortalResponse("")
	if result.Success {
		t.Error("expected failure for empty response")
	}
	if result.Status != StatusNoConnection {
		t.Errorf("expected StatusNoConnection, got %v", result.Status)
	}
}

func TestParseCaptivePortalResponse_LiveStatus(t *testing.T) {
	cases := []string{
		"<status>Live</status>",
		"<status>live</status>",
		`{"status":"live"}`,
		"Live user detected",
	}
	for _, input := range cases {
		result := ParseCaptivePortalResponse(input)
		if !result.Success {
			t.Errorf("expected success for %q", input)
		}
		if result.Status != StatusLoggedIn {
			t.Errorf("expected StatusLoggedIn for %q, got %v", input, result.Status)
		}
	}
}

func TestParseCaptivePortalResponse_Success(t *testing.T) {
	cases := []string{
		"Authentication success",
		"User has been authenticated",
	}
	for _, input := range cases {
		result := ParseCaptivePortalResponse(input)
		if !result.Success {
			t.Errorf("expected success for %q", input)
		}
	}
}

func TestParseCaptivePortalResponse_WrongPassword(t *testing.T) {
	cases := []string{
		"Invalid credentials provided",
		"Wrong username or password",
		"Incorrect login details",
		"Authentication failed for user",
		"Login failed: bad credentials",
	}
	for _, input := range cases {
		result := ParseCaptivePortalResponse(input)
		if result.Success {
			t.Errorf("expected failure for %q", input)
		}
		if result.Status != StatusWrongPassword {
			t.Errorf("expected StatusWrongPassword for %q, got %v", input, result.Status)
		}
	}
}

func TestParseCaptivePortalResponse_MaxLimit(t *testing.T) {
	cases := []string{
		"Maximum connections reached",
		"User limit exceeded",
		"Max user count reached",
		"Concurrent session limit",
	}
	for _, input := range cases {
		result := ParseCaptivePortalResponse(input)
		if result.Success {
			t.Errorf("expected failure for %q", input)
		}
		if result.Status != StatusMaxLimit {
			t.Errorf("expected StatusMaxLimit for %q, got %v", input, result.Status)
		}
	}
}

func TestParseCaptivePortalResponse_DataLimit(t *testing.T) {
	cases := []string{
		"Data limit exceeded",
		"Your quota has been reached",
	}
	for _, input := range cases {
		result := ParseCaptivePortalResponse(input)
		if result.Success {
			t.Errorf("expected failure for %q", input)
		}
		if result.Status != StatusDataLimit {
			t.Errorf("expected StatusDataLimit for %q, got %v", input, result.Status)
		}
	}
}

func TestParseCaptivePortalResponse_Welcome(t *testing.T) {
	result := ParseCaptivePortalResponse("Welcome to the network")
	if !result.Success {
		t.Error("expected success for welcome message")
	}
}

func TestParseCaptivePortalResponse_WelcomeWithError(t *testing.T) {
	result := ParseCaptivePortalResponse("Welcome error page - authentication failed")
	if result.Success {
		t.Error("expected failure for welcome with error")
	}
}

func TestParseCaptivePortalResponse_UnknownResponse(t *testing.T) {
	result := ParseCaptivePortalResponse("some random html page content")
	if result.Success {
		t.Error("expected failure for unknown response")
	}
}

func TestParseXMLResponse_Empty(t *testing.T) {
	result := ParseXMLResponse("")
	if result.Success {
		t.Error("expected failure for empty XML")
	}
	if result.Status != StatusNoConnection {
		t.Errorf("expected StatusNoConnection, got %v", result.Status)
	}
}

func TestParseXMLResponse_InvalidXML(t *testing.T) {
	result := ParseXMLResponse("not xml at all - success")
	if !result.Success {
		t.Error("expected success from alternative parser for text containing 'success'")
	}
}

func TestParseXMLResponse_ValidXML_Success(t *testing.T) {
	xml := `<ajax-response><response><status>success</status><message>Login successful</message></response></ajax-response>`
	result := ParseXMLResponse(xml)
	if !result.Success {
		t.Error("expected success for valid XML success response")
	}
	if result.Status != StatusLoggedIn {
		t.Errorf("expected StatusLoggedIn, got %v", result.Status)
	}
	if result.ResponseData != xml {
		t.Error("expected ResponseData to contain original XML")
	}
}

func TestParseXMLResponse_ValidXML_Failure(t *testing.T) {
	xml := `<ajax-response><response><status>authentication failed</status><message>Bad password</message></response></ajax-response>`
	result := ParseXMLResponse(xml)
	if result.Success {
		t.Error("expected failure for auth failed XML response")
	}
	if result.Status != StatusWrongPassword {
		t.Errorf("expected StatusWrongPassword, got %v", result.Status)
	}
}

func TestContainsWord(t *testing.T) {
	tests := []struct {
		text     string
		word     string
		expected bool
	}{
		{"authentication success", "success", true},
		{"successful login", "success", false}, // "success" is part of "successful"
		{"success", "success", true},
		{"the success is here", "success", true},
		{"no match", "success", false},
		{"", "success", false},
		{"success", "", false},
		{"invalid credentials", "invalid", true},
		{"invalidated", "invalid", false},
		{"pre-invalid check", "invalid", true},
		{"limit reached", "limit", true},
		{"unlimited", "limit", false},
	}

	for _, tt := range tests {
		result := containsWord(tt.text, tt.word)
		if result != tt.expected {
			t.Errorf("containsWord(%q, %q) = %v, want %v", tt.text, tt.word, result, tt.expected)
		}
	}
}

func TestIsLetterOrDigit(t *testing.T) {
	tests := []struct {
		b        byte
		expected bool
	}{
		{'a', true}, {'z', true}, {'A', true}, {'Z', true},
		{'0', true}, {'9', true},
		{' ', false}, {'-', false}, {'_', false}, {'.', false},
	}
	for _, tt := range tests {
		result := isLetterOrDigit(tt.b)
		if result != tt.expected {
			t.Errorf("isLetterOrDigit(%c) = %v, want %v", tt.b, result, tt.expected)
		}
	}
}

func TestMapSophosStatusToLoginStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected LoginStatus
	}{
		{"success", StatusLoggedIn},
		{"200 OK", StatusLoggedIn},
		{"authenticated", StatusLoggedIn},
		{"authentication failed", StatusWrongPassword},
		{"invalid credentials", StatusWrongPassword},
		{"401 Unauthorized", StatusWrongPassword},
		{"user limit exceeded", StatusMaxLimit},
		{"maximum reached", StatusMaxLimit},
		{"data limit exceeded", StatusDataLimit},
		{"quota reached", StatusDataLimit},
		{"session timeout", StatusLogInAgain},
		{"expired", StatusLogInAgain},
		{"random status", StatusLoggedOut},
		{"", StatusNoConnection},
	}
	for _, tt := range tests {
		result := mapSophosStatusToLoginStatus(tt.input)
		if result != tt.expected {
			t.Errorf("mapSophosStatusToLoginStatus(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsSuccessStatus(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"success", true},
		{"200", true},
		{"ok", true},
		{"authenticated", true},
		{"failed", false},
		{"error", false},
		{"", false},
	}
	for _, tt := range tests {
		result := isSuccessStatus(tt.input)
		if result != tt.expected {
			t.Errorf("isSuccessStatus(%q) = %v, want %v", tt.input, result, tt.expected)
		}
	}
}

func TestParseAlternativeResponse_Truncation(t *testing.T) {
	longResponse := ""
	for i := 0; i < 200; i++ {
		longResponse += "x"
	}
	result := parseAlternativeResponse(longResponse)
	if result.Success {
		t.Error("expected failure for unknown long response")
	}
	if len(result.Message) > 120 {
		t.Errorf("expected truncated message, got length %d", len(result.Message))
	}
}
