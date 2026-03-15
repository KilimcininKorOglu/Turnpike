package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewService(t *testing.T) {
	s := NewService("192.168.1.1", 8090, true, false, 30*time.Second)
	if s.serverIP != "192.168.1.1" {
		t.Errorf("expected serverIP 192.168.1.1, got %s", s.serverIP)
	}
	if s.captivePortalPort != 8090 {
		t.Errorf("expected port 8090, got %d", s.captivePortalPort)
	}
	if !s.useHTTPS {
		t.Error("expected useHTTPS to be true")
	}
	if s.skipSSLValidation {
		t.Error("expected skipSSLValidation to be false")
	}
	if s.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
}

func TestServiceGetCaptivePortalURL_HTTPS(t *testing.T) {
	s := NewService("172.16.100.2", 8090, true, false, 30*time.Second)
	url := s.GetCaptivePortalURL()
	expected := "https://172.16.100.2:8090"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestServiceGetCaptivePortalURL_HTTP(t *testing.T) {
	s := NewService("172.16.100.2", 8090, false, false, 30*time.Second)
	url := s.GetCaptivePortalURL()
	expected := "http://172.16.100.2:8090"
	if url != expected {
		t.Errorf("expected %s, got %s", expected, url)
	}
}

func TestServiceUpdateSettings(t *testing.T) {
	s := NewService("10.0.0.1", 8090, false, false, 30*time.Second)
	s.UpdateSettings("10.0.0.2", 8443, true, true)

	if s.serverIP != "10.0.0.2" {
		t.Errorf("expected serverIP 10.0.0.2, got %s", s.serverIP)
	}
	if s.captivePortalPort != 8443 {
		t.Errorf("expected port 8443, got %d", s.captivePortalPort)
	}
	if !s.useHTTPS {
		t.Error("expected useHTTPS to be true")
	}
	if !s.skipSSLValidation {
		t.Error("expected skipSSLValidation to be true")
	}
}

func TestServiceUpdateSettings_RecreatesClientOnSSLChange(t *testing.T) {
	s := NewService("10.0.0.1", 8090, false, false, 30*time.Second)
	originalClient := s.httpClient

	// No SSL change - client should remain same reference
	s.UpdateSettings("10.0.0.2", 8443, true, false)
	if s.httpClient != originalClient {
		t.Error("expected same httpClient when SSL didn't change")
	}

	// SSL change - client should be recreated
	s.UpdateSettings("10.0.0.2", 8443, true, true)
	if s.httpClient == originalClient {
		t.Error("expected new httpClient when SSL changed")
	}
}

func TestServiceLoginAsync_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/login" {
			t.Errorf("expected path /login, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		err := r.ParseForm()
		if err != nil {
			t.Fatal(err)
		}
		if r.FormValue("username") != "testuser" {
			t.Errorf("expected username testuser, got %s", r.FormValue("username"))
		}
		if r.FormValue("password") != "testpass" {
			t.Errorf("expected password testpass, got %s", r.FormValue("password"))
		}
		if r.FormValue("mode") != "191" {
			t.Errorf("expected mode 191, got %s", r.FormValue("mode"))
		}

		fmt.Fprint(w, "<status>Live</status>")
	}))
	defer server.Close()

	// Parse server URL to get host and port
	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port := 80
	fmt.Sscanf(parts[1], "%d", &port)

	s := NewService(host, port, false, false, 5*time.Second)
	result, err := s.LoginAsync("testuser", "testpass")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Success {
		t.Errorf("expected success, got failure: %s", result.Message)
	}
	if result.Status != StatusLoggedIn {
		t.Errorf("expected StatusLoggedIn, got %v", result.Status)
	}
}

func TestServiceLoginAsync_WrongPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Authentication failed - invalid credentials")
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port := 80
	fmt.Sscanf(parts[1], "%d", &port)

	s := NewService(host, port, false, false, 5*time.Second)
	result, err := s.LoginAsync("testuser", "wrongpass")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("expected failure for wrong password")
	}
	if result.Status != StatusWrongPassword {
		t.Errorf("expected StatusWrongPassword, got %v", result.Status)
	}
}

func TestServiceLoginAsync_ConnectionError(t *testing.T) {
	s := NewService("192.0.2.1", 1, false, false, 1*time.Second) // non-routable address
	result, err := s.LoginAsync("testuser", "testpass")

	if err == nil {
		t.Error("expected error for unreachable server")
	}
	if result.Success {
		t.Error("expected failure for connection error")
	}
	if result.Status != StatusNoConnection {
		t.Errorf("expected StatusNoConnection, got %v", result.Status)
	}
}

func TestServiceLogoutAsync_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/logout" {
			t.Errorf("expected path /logout, got %s", r.URL.Path)
		}

		err := r.ParseForm()
		if err != nil {
			t.Fatal(err)
		}
		if r.FormValue("mode") != "193" {
			t.Errorf("expected mode 193, got %s", r.FormValue("mode"))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port := 80
	fmt.Sscanf(parts[1], "%d", &port)

	s := NewService(host, port, false, false, 5*time.Second)
	success, err := s.LogoutAsync("testuser")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !success {
		t.Error("expected logout success")
	}
}

func TestServiceLogoutAsync_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port := 80
	fmt.Sscanf(parts[1], "%d", &port)

	s := NewService(host, port, false, false, 5*time.Second)
	success, err := s.LogoutAsync("testuser")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if success {
		t.Error("expected logout failure for server error")
	}
}

func TestServiceCheckSessionAsync_Active(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Microsoft Connect Test")
	}))
	defer server.Close()

	_ = NewService("localhost", 8090, false, false, 5*time.Second)
	// CheckSessionAsync uses a hardcoded URL (msftconnecttest.com)
	// so we validate the expected response content against our test server
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	body := make([]byte, 1024)
	n, _ := resp.Body.Read(body)
	content := strings.TrimSpace(string(body[:n]))
	if content != connectTestExpected {
		t.Errorf("expected %q, got %q", connectTestExpected, content)
	}
}

func TestServiceGetters(t *testing.T) {
	s := NewService("10.0.0.1", 9090, true, true, 30*time.Second)

	if s.ServerIP() != "10.0.0.1" {
		t.Errorf("expected ServerIP 10.0.0.1, got %s", s.ServerIP())
	}
	if s.CaptivePortalPort() != 9090 {
		t.Errorf("expected CaptivePortalPort 9090, got %d", s.CaptivePortalPort())
	}
	if !s.UseHTTPS() {
		t.Error("expected UseHTTPS true")
	}
	if !s.SkipSSLValidation() {
		t.Error("expected SkipSSLValidation true")
	}
}

func TestServiceClose(t *testing.T) {
	s := NewService("10.0.0.1", 8090, false, false, 30*time.Second)
	// Should not panic
	s.Close()
}
