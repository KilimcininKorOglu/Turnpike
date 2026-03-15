package auth

import (
	"encoding/xml"
	"strings"
)

// sophosResponse represents the XML response structure from Sophos XG
type sophosResponse struct {
	XMLName  xml.Name              `xml:"ajax-response"`
	Response sophosResponseContent `xml:"response"`
}

type sophosResponseContent struct {
	Status  string `xml:"status"`
	Message string `xml:"message"`
}

// ParseCaptivePortalResponse parses the captive portal HTTP response text
// and determines the login status
func ParseCaptivePortalResponse(responseText string) LoginResult {
	if responseText == "" {
		return CreateFailure(StatusNoConnection, "Empty response")
	}

	normalized := strings.ToLower(responseText)

	// Check for Sophos XG specific success patterns first (more reliable)
	if strings.Contains(normalized, "<status>live</status>") ||
		strings.Contains(normalized, "\"status\":\"live\"") ||
		strings.Contains(normalized, "live user") ||
		containsWord(normalized, "success") ||
		containsWord(normalized, "authenticated") {
		return CreateSuccess(StatusLoggedIn, "Login successful")
	}

	// Check for error patterns with word boundary matching
	if containsWord(normalized, "invalid") ||
		containsWord(normalized, "wrong") ||
		containsWord(normalized, "incorrect") ||
		strings.Contains(normalized, "authentication failed") ||
		strings.Contains(normalized, "login failed") {
		return CreateFailure(StatusWrongPassword, "Invalid credentials")
	}

	if strings.Contains(normalized, "data limit") ||
		strings.Contains(normalized, "quota") {
		return CreateFailure(StatusDataLimit, "Data limit reached")
	}

	if containsWord(normalized, "limit") ||
		containsWord(normalized, "maximum") ||
		strings.Contains(normalized, "max user") ||
		strings.Contains(normalized, "concurrent") {
		return CreateFailure(StatusMaxLimit, "Connection limit reached")
	}

	// If response contains "welcome" as standalone word, consider it success
	if containsWord(normalized, "welcome") &&
		!strings.Contains(normalized, "error") &&
		!strings.Contains(normalized, "fail") {
		return CreateSuccess(StatusLoggedIn, "Login successful")
	}

	return CreateFailure(StatusNoConnection, "Unable to parse response")
}

// ParseXMLResponse parses Sophos XG XML response format
func ParseXMLResponse(xmlResponse string) LoginResult {
	if xmlResponse == "" {
		return CreateFailure(StatusNoConnection, "Empty response from server")
	}

	var resp sophosResponse
	err := xml.Unmarshal([]byte(xmlResponse), &resp)
	if err != nil {
		// Try alternative parsing for different response formats
		return parseAlternativeResponse(xmlResponse)
	}

	if resp.Response.Status == "" {
		return parseAlternativeResponse(xmlResponse)
	}

	status := mapSophosStatusToLoginStatus(resp.Response.Status)
	isSuccess := isSuccessStatus(resp.Response.Status)

	return NewLoginResult(isSuccess, status, resp.Response.Message, xmlResponse)
}

// containsWord checks if a word exists as a standalone word (not part of another word)
func containsWord(text, word string) bool {
	if text == "" || word == "" {
		return false
	}

	index := strings.Index(text, word)
	for index >= 0 {
		startOk := index == 0 || !isLetterOrDigit(text[index-1])
		endIdx := index + len(word)
		endOk := endIdx >= len(text) || !isLetterOrDigit(text[endIdx])

		if startOk && endOk {
			return true
		}

		if endIdx >= len(text) {
			break
		}
		nextIdx := strings.Index(text[index+1:], word)
		if nextIdx < 0 {
			break
		}
		index = index + 1 + nextIdx
	}
	return false
}

func isLetterOrDigit(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
}

func parseAlternativeResponse(response string) LoginResult {
	lower := strings.ToLower(response)

	if strings.Contains(lower, "success") || strings.Contains(lower, "authenticated") || strings.Contains(lower, "200") {
		return CreateSuccess(StatusLoggedIn, "Authentication successful")
	}
	if strings.Contains(lower, "authentication failed") || strings.Contains(lower, "invalid") || strings.Contains(lower, "401") {
		return CreateFailure(StatusWrongPassword, "Authentication failed")
	}
	if strings.Contains(lower, "user limit") || strings.Contains(lower, "maximum") {
		return CreateFailure(StatusMaxLimit, "User limit exceeded")
	}
	if strings.Contains(lower, "data limit") {
		return CreateFailure(StatusDataLimit, "Data limit exceeded")
	}
	if strings.Contains(lower, "timeout") || strings.Contains(lower, "session") {
		return CreateFailure(StatusLogInAgain, "Session expired")
	}

	truncated := response
	if len(truncated) > 100 {
		truncated = truncated[:100]
	}
	return CreateFailure(StatusLoggedOut, "Unknown response: "+truncated)
}

func isSuccessStatus(status string) bool {
	if status == "" {
		return false
	}
	lower := strings.ToLower(status)
	return strings.Contains(lower, "success") ||
		strings.Contains(lower, "200") ||
		strings.Contains(lower, "ok") ||
		strings.Contains(lower, "authenticated")
}

func mapSophosStatusToLoginStatus(sophosStatus string) LoginStatus {
	if sophosStatus == "" {
		return StatusNoConnection
	}
	lower := strings.ToLower(sophosStatus)

	switch {
	case strings.Contains(lower, "success") || strings.Contains(lower, "200") || strings.Contains(lower, "authenticated"):
		return StatusLoggedIn
	case strings.Contains(lower, "authentication failed") || strings.Contains(lower, "invalid") || strings.Contains(lower, "401"):
		return StatusWrongPassword
	case strings.Contains(lower, "user limit") || strings.Contains(lower, "maximum"):
		return StatusMaxLimit
	case strings.Contains(lower, "data limit") || strings.Contains(lower, "quota"):
		return StatusDataLimit
	case strings.Contains(lower, "timeout") || strings.Contains(lower, "session") || strings.Contains(lower, "expired"):
		return StatusLogInAgain
	default:
		return StatusLoggedOut
	}
}
