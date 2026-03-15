package main

import (
	"flag"
	"os"

	"github.com/KilimcininKorOglu/Turnpike/internal/cli"
	"github.com/KilimcininKorOglu/Turnpike/internal/config"
	"github.com/KilimcininKorOglu/Turnpike/internal/ui"
)

func main() {
	// CLI command flags
	login := flag.Bool("login", false, "Perform login")
	logout := flag.Bool("logout", false, "Perform logout")
	status := flag.Bool("status", false, "Check session status")
	version := flag.Bool("version", false, "Show version")

	// GUI flags
	guiMode := flag.Bool("gui", false, "Launch GUI mode explicitly")
	minimized := flag.Bool("minimized", false, "Start GUI minimized (system tray)")

	// CLI options
	user := flag.String("u", "", "Username")
	pass := flag.String("p", "", "Password")
	server := flag.String("s", "", "Server IP")
	port := flag.Int("P", config.DefaultCaptivePortalPort, "Captive portal port")
	useHTTPS := flag.Bool("https", config.DefaultUseHTTPS, "Use HTTPS")
	skipSSL := flag.Bool("skip-ssl", config.DefaultSkipSSLValidation, "Skip SSL certificate validation")
	useConfig := flag.Bool("config", false, "Use saved credentials")

	flag.Parse()

	// CLI mode: any command flag triggers CLI
	if *login || *logout || *status || *version {
		exitCode := cli.Run(cli.Options{
			Login:     *login,
			Logout:    *logout,
			Status:    *status,
			Version:   *version,
			Username:  *user,
			Password:  *pass,
			Server:    *server,
			Port:      *port,
			UseHTTPS:  *useHTTPS,
			SkipSSL:   *skipSSL,
			UseConfig: *useConfig,
		})
		os.Exit(exitCode)
	}

	// GUI mode: default (no args) or explicit --gui
	_ = guiMode // always GUI if no CLI command
	app := ui.NewApplication(*minimized)
	app.Run()
}
