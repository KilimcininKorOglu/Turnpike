package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/KilimcininKorOglu/SophosXG-User-Client/internal/auth"
	"github.com/KilimcininKorOglu/SophosXG-User-Client/internal/config"
)

// AppVersion is the application version string, injected at build time via ldflags.
var AppVersion = "dev"

const (
	// Exit codes
	ExitSuccess   = 0
	ExitError     = 1
	ExitAuthError = 2
)

// Options holds all CLI command options
type Options struct {
	Login     bool
	Logout    bool
	Status    bool
	Version   bool
	Username  string
	Password  string
	Server    string
	Port      int
	UseHTTPS  bool
	SkipSSL   bool
	UseConfig bool
	Stdout    *os.File // for testing output redirection
	Stderr    *os.File // for testing error redirection
}

// Run executes the CLI command based on options and returns an exit code
func Run(opts Options) int {
	if opts.Stdout == nil {
		opts.Stdout = os.Stdout
	}
	if opts.Stderr == nil {
		opts.Stderr = os.Stderr
	}

	if opts.Version {
		return runVersion(opts)
	}
	if opts.Login {
		return runLogin(opts)
	}
	if opts.Logout {
		return runLogout(opts)
	}
	if opts.Status {
		return runStatus(opts)
	}

	fmt.Fprintln(opts.Stderr, "No command specified. Use --login, --logout, --status, or --version.")
	return ExitError
}

func runLogin(opts Options) int {
	// Load from saved config if requested
	if opts.UseConfig {
		creds := loadSavedCredentials()
		if creds != nil {
			if opts.Username == "" {
				opts.Username = creds.Username
			}
			if opts.Password == "" {
				opts.Password = creds.Password
			}
			if opts.Server == "" {
				opts.Server = creds.ServerIP
			}
			if opts.Port == config.DefaultCaptivePortalPort {
				opts.Port = creds.CaptivePortalPort
			}
			opts.UseHTTPS = creds.UseHTTPS
			opts.SkipSSL = creds.SkipSSLValidation
		}
	}

	// Validate required fields
	if opts.Server == "" {
		fmt.Fprintln(opts.Stderr, "Error: Server IP is required (-s)")
		return ExitError
	}
	if opts.Username == "" {
		fmt.Fprintln(opts.Stderr, "Error: Username is required (-u)")
		return ExitError
	}

	// Prompt for password if not provided
	if opts.Password == "" {
		prompted, err := PromptPassword()
		if err != nil {
			fmt.Fprintf(opts.Stderr, "Error reading password: %v\n", err)
			return ExitError
		}
		opts.Password = prompted
	}

	service := auth.NewService(opts.Server, opts.Port, opts.UseHTTPS, opts.SkipSSL, config.HTTPRequestTimeout)
	defer service.Close()

	result, err := service.LoginAsync(opts.Username, opts.Password)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "Connection error: %v\n", err)
		return ExitError
	}

	if result.Success {
		fmt.Fprintf(opts.Stdout, "✓ Logged in as %s on %s:%d\n", opts.Username, opts.Server, opts.Port)
		return ExitSuccess
	}

	fmt.Fprintf(opts.Stdout, "✕ Login failed: %s\n", result.Message)
	return ExitAuthError
}

func runLogout(opts Options) int {
	// Load from saved config if requested
	if opts.UseConfig {
		creds := loadSavedCredentials()
		if creds != nil {
			if opts.Username == "" {
				opts.Username = creds.Username
			}
			if opts.Server == "" {
				opts.Server = creds.ServerIP
			}
			if opts.Port == config.DefaultCaptivePortalPort {
				opts.Port = creds.CaptivePortalPort
			}
			opts.UseHTTPS = creds.UseHTTPS
			opts.SkipSSL = creds.SkipSSLValidation
		}
	}

	if opts.Server == "" {
		fmt.Fprintln(opts.Stderr, "Error: Server IP is required (-s)")
		return ExitError
	}
	if opts.Username == "" {
		fmt.Fprintln(opts.Stderr, "Error: Username is required (-u)")
		return ExitError
	}

	service := auth.NewService(opts.Server, opts.Port, opts.UseHTTPS, opts.SkipSSL, config.HTTPRequestTimeout)
	defer service.Close()

	success, err := service.LogoutAsync(opts.Username)
	if err != nil {
		fmt.Fprintf(opts.Stderr, "Connection error: %v\n", err)
		return ExitError
	}

	if success {
		fmt.Fprintf(opts.Stdout, "✓ Logged out from %s:%d\n", opts.Server, opts.Port)
		return ExitSuccess
	}

	fmt.Fprintln(opts.Stdout, "✕ Logout failed")
	return ExitError
}

func runStatus(opts Options) int {
	if opts.Server == "" && opts.UseConfig {
		creds := loadSavedCredentials()
		if creds != nil {
			opts.Server = creds.ServerIP
			opts.Username = creds.Username
		}
	}

	service := auth.NewService(
		opts.Server, opts.Port, opts.UseHTTPS, opts.SkipSSL,
		10*time.Second, // shorter timeout for status check
	)
	defer service.Close()

	isActive, err := service.CheckSessionAsync(opts.Username)
	if err != nil || !isActive {
		fmt.Fprintln(opts.Stdout, "○ Disconnected")
		return ExitError
	}

	fmt.Fprintln(opts.Stdout, "● Connected (session active)")
	return ExitSuccess
}

func runVersion(opts Options) int {
	fmt.Fprintf(opts.Stdout, "Sophos XG User Login v%s (%s/%s/%s)\n",
		AppVersion, runtime.Compiler, runtime.GOOS, runtime.GOARCH)
	return ExitSuccess
}

func loadSavedCredentials() *config.Credentials {
	exePath, err := os.Executable()
	if err != nil {
		return nil
	}
	settingsDir := filepath.Dir(exePath)
	cm := config.NewCredentialManager(settingsDir)
	creds := cm.GetCredentials()
	if creds.Username == "" {
		return nil
	}
	return &creds
}
