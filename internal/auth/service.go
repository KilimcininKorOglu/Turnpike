package auth

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	connectTestURL      = "http://www.msftconnecttest.com/connecttest.txt"
	connectTestExpected = "Microsoft Connect Test"
	loginMode           = "191"
	logoutMode          = "193"
)

// Service handles Sophos XG captive portal authentication
type Service struct {
	serverIP          string
	captivePortalPort int
	useHTTPS          bool
	skipSSLValidation bool
	httpClient        *http.Client
	timeout           time.Duration
}

// NewService creates a new authentication service with default settings
func NewService(serverIP string, port int, useHTTPS, skipSSL bool, timeout time.Duration) *Service {
	s := &Service{
		serverIP:          serverIP,
		captivePortalPort: port,
		useHTTPS:          useHTTPS,
		skipSSLValidation: skipSSL,
		timeout:           timeout,
	}
	s.recreateHTTPClient()
	return s
}

// UpdateSettings updates the service configuration
func (s *Service) UpdateSettings(serverIP string, port int, useHTTPS, skipSSL bool) {
	sslChanged := s.skipSSLValidation != skipSSL
	s.serverIP = serverIP
	s.captivePortalPort = port
	s.useHTTPS = useHTTPS
	s.skipSSLValidation = skipSSL

	if sslChanged {
		s.recreateHTTPClient()
	}
}

// LoginAsync performs captive portal authentication
func (s *Service) LoginAsync(username, password string) (LoginResult, error) {
	loginURL := s.getCaptivePortalURL() + "/login"

	formData := url.Values{
		"username": {username},
		"password": {password},
		"mode":     {loginMode},
	}

	resp, err := s.httpClient.PostForm(loginURL, formData)
	if err != nil {
		return CreateFailure(StatusNoConnection, "Login failed: "+err.Error()), err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return CreateFailure(StatusNoConnection, "Failed to read response: "+err.Error()), err
	}

	return ParseCaptivePortalResponse(string(body)), nil
}

// LogoutAsync performs captive portal logout
func (s *Service) LogoutAsync(username string) (bool, error) {
	logoutURL := s.getCaptivePortalURL() + "/logout"

	formData := url.Values{
		"username": {username},
		"mode":     {logoutMode},
	}

	resp, err := s.httpClient.PostForm(logoutURL, formData)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300, nil
}

// CheckSessionAsync checks if the session is still active by testing internet connectivity.
// The username parameter is used for logging purposes only.
func (s *Service) CheckSessionAsync(username string) (bool, error) {
	resp, err := s.httpClient.Get(connectTestURL)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(body)) == connectTestExpected, nil
}

// GetCaptivePortalURL returns the captive portal base URL (exported for testing/logging)
func (s *Service) GetCaptivePortalURL() string {
	return s.getCaptivePortalURL()
}

// ServerIP returns the current server IP
func (s *Service) ServerIP() string {
	return s.serverIP
}

// CaptivePortalPort returns the current port
func (s *Service) CaptivePortalPort() int {
	return s.captivePortalPort
}

// UseHTTPS returns whether HTTPS is enabled
func (s *Service) UseHTTPS() bool {
	return s.useHTTPS
}

// SkipSSLValidation returns whether SSL validation is skipped
func (s *Service) SkipSSLValidation() bool {
	return s.skipSSLValidation
}

func (s *Service) getCaptivePortalURL() string {
	protocol := "http"
	if s.useHTTPS {
		protocol = "https"
	}
	return fmt.Sprintf("%s://%s:%d", protocol, s.serverIP, s.captivePortalPort)
}

func (s *Service) recreateHTTPClient() {
	transport := &http.Transport{}
	if s.skipSSLValidation {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	s.httpClient = &http.Client{
		Timeout:   s.timeout,
		Transport: transport,
	}
}

// Close cleans up the service resources
func (s *Service) Close() {
	if s.httpClient != nil {
		s.httpClient.CloseIdleConnections()
	}
}
