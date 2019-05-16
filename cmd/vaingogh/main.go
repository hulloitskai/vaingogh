package main

import (
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	ess "github.com/unixpickle/essentials"

	"github.com/stevenxie/vaingogh/config"
	"github.com/stevenxie/vaingogh/imports"
	"github.com/stevenxie/vaingogh/internal/info"
)

var (
	app = &cobra.Command{
		Use:     info.Namespace,
		Short:   info.Namespace + " is a vanity URL generator for your Go modules.",
		Version: info.Version,
	}

	logger = buildLogger()
)

func main() {
	// Initialization.
	prepareEnv()
	configureApp()

	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}

// prepareEnv loads envvars from .env files.
func prepareEnv() {
	if err := godotenv.Load(".env", ".env.local"); err != nil {
		if !strings.Contains( // unknown error
			err.Error(),
			"no such file or directory",
		) {
			ess.Die("Error reading '.env' file:", err)
		}
	}
}

func configureApp() {
	app.AddCommand(reposCmd)
	app.AddCommand(completionCmd)

	// Disable help command.
	app.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

// buildLogger builds an application-level zerolog.Logger.
func buildLogger() zerolog.Logger {
	var logger zerolog.Logger
	if os.Getenv("GOENV") == "production" {
		logger = zerolog.New(os.Stdout)
	} else {
		logger = zerolog.New(zerolog.NewConsoleWriter())
	}
	return logger.With().Timestamp().Logger()
}

func buildWatcher(cfg *config.Config, l zerolog.Logger) *imports.RepoWatcher {
	var (
		lister  = cfg.BuildGithubRepoLister()
		watcher = cfg.BuildRepoWatcher(lister)
	)
	watcher.SetLogger(l.With().Str("component", "watcher").Logger())
	return watcher
}
