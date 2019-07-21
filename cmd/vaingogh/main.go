package main

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/stevenxie/api/pkg/cmdutil"
	"github.com/stevenxie/vaingogh/internal/info"
)

var (
	app = &cobra.Command{
		Use:     info.Namespace,
		Short:   info.Namespace + " is a vanity URL generator for your Go modules.",
		Version: info.Version,
	}

	log = buildLogger()
)

func main() {
	// Perform app initialization.
	cmdutil.PrepareEnv()
	configureApp()

	if err := app.Execute(); err != nil {
		os.Exit(1)
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
