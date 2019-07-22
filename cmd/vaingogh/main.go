package main

import (
	"os"

	"github.com/sirupsen/logrus"
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

func init() {
	app.AddCommand(reposCmd)
	app.AddCommand(serveCmd)
	app.AddCommand(completionCmd)

	// Disable help command.
	app.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}

func main() {
	cmdutil.PrepareEnv()
	if err := app.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildLogger builds an application-level logger, which also captures errors
// using Sentry.
func buildLogger() *logrus.Logger {
	log := logrus.New()
	log.SetOutput(os.Stdout)

	// Set logger level.
	if os.Getenv("GOENV") == "development" {
		log.SetLevel(logrus.DebugLevel)
	}
	return log
}
