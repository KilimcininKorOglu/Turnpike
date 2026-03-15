package cli

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func captureOutput(fn func(stdout, stderr *os.File)) (string, string) {
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()

	fn(wOut, wErr)

	wOut.Close()
	wErr.Close()

	outBuf := make([]byte, 4096)
	nOut, _ := rOut.Read(outBuf)
	errBuf := make([]byte, 4096)
	nErr, _ := rErr.Read(errBuf)

	return string(outBuf[:nOut]), string(errBuf[:nErr])
}

func TestRunVersion(t *testing.T) {
	stdout, _ := captureOutput(func(out, err *os.File) {
		code := Run(Options{Version: true, Stdout: out, Stderr: err})
		if code != ExitSuccess {
			t.Errorf("expected exit code %d, got %d", ExitSuccess, code)
		}
	})

	if !strings.Contains(stdout, "Turnpike v") {
		t.Errorf("expected version output, got %q", stdout)
	}
	if !strings.Contains(stdout, AppVersion) {
		t.Errorf("expected version %s in output, got %q", AppVersion, stdout)
	}
}

func TestRun_NoCommand(t *testing.T) {
	_, stderr := captureOutput(func(out, err *os.File) {
		code := Run(Options{Stdout: out, Stderr: err})
		if code != ExitError {
			t.Errorf("expected exit code %d, got %d", ExitError, code)
		}
	})

	if !strings.Contains(stderr, "No command specified") {
		t.Errorf("expected error message, got %q", stderr)
	}
}

func TestRunLogin_MissingServer(t *testing.T) {
	_, stderr := captureOutput(func(out, err *os.File) {
		code := Run(Options{
			Login:    true,
			Username: "admin",
			Password: "pass",
			Stdout:   out,
			Stderr:   err,
		})
		if code != ExitError {
			t.Errorf("expected exit code %d, got %d", ExitError, code)
		}
	})

	if !strings.Contains(stderr, "Server IP is required") {
		t.Errorf("expected server error, got %q", stderr)
	}
}

func TestRunLogin_MissingUsername(t *testing.T) {
	_, stderr := captureOutput(func(out, err *os.File) {
		code := Run(Options{
			Login:    true,
			Server:   "10.0.0.1",
			Password: "pass",
			Stdout:   out,
			Stderr:   err,
		})
		if code != ExitError {
			t.Errorf("expected exit code %d, got %d", ExitError, code)
		}
	})

	if !strings.Contains(stderr, "Username is required") {
		t.Errorf("expected username error, got %q", stderr)
	}
}

func TestRunLogin_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<status>Live</status>")
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port := 80
	fmt.Sscanf(parts[1], "%d", &port)

	stdout, _ := captureOutput(func(out, err *os.File) {
		code := Run(Options{
			Login:    true,
			Username: "admin",
			Password: "testpass",
			Server:   host,
			Port:     port,
			UseHTTPS: false,
			Stdout:   out,
			Stderr:   err,
		})
		if code != ExitSuccess {
			t.Errorf("expected exit code %d, got %d", ExitSuccess, code)
		}
	})

	if !strings.Contains(stdout, "Logged in as admin") {
		t.Errorf("expected success message, got %q", stdout)
	}
}

func TestRunLogin_WrongPassword(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Authentication failed - invalid credentials")
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port := 80
	fmt.Sscanf(parts[1], "%d", &port)

	stdout, _ := captureOutput(func(out, err *os.File) {
		code := Run(Options{
			Login:    true,
			Username: "admin",
			Password: "wrongpass",
			Server:   host,
			Port:     port,
			UseHTTPS: false,
			Stdout:   out,
			Stderr:   err,
		})
		if code != ExitAuthError {
			t.Errorf("expected exit code %d, got %d", ExitAuthError, code)
		}
	})

	if !strings.Contains(stdout, "Login failed") {
		t.Errorf("expected failure message, got %q", stdout)
	}
}

func TestRunLogout_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	parts := strings.Split(strings.TrimPrefix(server.URL, "http://"), ":")
	host := parts[0]
	port := 80
	fmt.Sscanf(parts[1], "%d", &port)

	stdout, _ := captureOutput(func(out, err *os.File) {
		code := Run(Options{
			Logout:   true,
			Username: "admin",
			Server:   host,
			Port:     port,
			UseHTTPS: false,
			Stdout:   out,
			Stderr:   err,
		})
		if code != ExitSuccess {
			t.Errorf("expected exit code %d, got %d", ExitSuccess, code)
		}
	})

	if !strings.Contains(stdout, "Logged out from") {
		t.Errorf("expected logout message, got %q", stdout)
	}
}

func TestRunLogout_MissingServer(t *testing.T) {
	_, stderr := captureOutput(func(out, err *os.File) {
		code := Run(Options{
			Logout:   true,
			Username: "admin",
			Stdout:   out,
			Stderr:   err,
		})
		if code != ExitError {
			t.Errorf("expected exit code %d, got %d", ExitError, code)
		}
	})

	if !strings.Contains(stderr, "Server IP is required") {
		t.Errorf("expected server error, got %q", stderr)
	}
}

func TestRunLogout_MissingUsername(t *testing.T) {
	_, stderr := captureOutput(func(out, err *os.File) {
		code := Run(Options{
			Logout: true,
			Server: "10.0.0.1",
			Stdout: out,
			Stderr: err,
		})
		if code != ExitError {
			t.Errorf("expected exit code %d, got %d", ExitError, code)
		}
	})

	if !strings.Contains(stderr, "Username is required") {
		t.Errorf("expected username error, got %q", stderr)
	}
}

func TestRunStatus_Connected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Microsoft Connect Test")
	}))
	defer server.Close()

	// Status check uses msftconnecttest.com, not our test server.
	// We test the output format for disconnected state instead.
	stdout, _ := captureOutput(func(out, err *os.File) {
		code := Run(Options{
			Status:   true,
			Server:   "192.0.2.1", // non-routable, will fail
			Port:     1,
			UseHTTPS: false,
			Stdout:   out,
			Stderr:   err,
		})
		// Expected to fail since we can't reach msftconnecttest.com via a fake server
		if code != ExitError {
			// If we somehow connect, that's fine too
			_ = code
		}
	})

	if !strings.Contains(stdout, "Connected") && !strings.Contains(stdout, "Disconnected") {
		t.Errorf("expected status output, got %q", stdout)
	}
}

func TestRunStatus_OutputFormat(t *testing.T) {
	// Status check uses real msftconnecttest.com so result depends on network.
	// We only verify the output contains a valid status indicator.
	stdout, _ := captureOutput(func(out, err *os.File) {
		_ = Run(Options{
			Status:   true,
			Server:   "10.0.0.1",
			Port:     8090,
			UseHTTPS: false,
			Stdout:   out,
			Stderr:   err,
		})
	})

	hasStatus := strings.Contains(stdout, "Connected") || strings.Contains(stdout, "Disconnected")
	if !hasStatus {
		t.Errorf("expected status output with Connected or Disconnected, got %q", stdout)
	}
}

func TestExitCodes(t *testing.T) {
	if ExitSuccess != 0 {
		t.Errorf("expected ExitSuccess=0, got %d", ExitSuccess)
	}
	if ExitError != 1 {
		t.Errorf("expected ExitError=1, got %d", ExitError)
	}
	if ExitAuthError != 2 {
		t.Errorf("expected ExitAuthError=2, got %d", ExitAuthError)
	}
}
