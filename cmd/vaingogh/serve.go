package main

import (
	"fmt"

	"github.com/stevenxie/vaingogh/server"
	"github.com/stevenxie/vaingogh/vanity"
	"github.com/stevenxie/vaingogh/vanity/template"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/stevenxie/vaingogh/config"
	"github.com/stevenxie/vaingogh/repo"
	"github.com/stevenxie/vaingogh/repo/github"
)

var (
	serveCmd = &cobra.Command{
		Use:          "serve",
		Short:        "Serve a vanity imports redirection server.",
		RunE:         execServe,
		Args:         cobra.NoArgs,
		SilenceUsage: true,
	}

	serveOpts struct {
		Port int
	}
)

func init() {
	serveCmd.Flags().IntVarP(
		&serveOpts.Port,
		"port", "p",
		3000,
		"The port to listen on.",
	)
}

func execServe(*cobra.Command, []string) error {
	// Load and validate config file.
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "loading config")
	}
	if err = cfg.Validate(); err != nil {
		return errors.Wrap(err, "invalid config")
	}

	// Build repo service.
	var lister repo.ListerService
	{
		// Initiate GitHub client.
		client, err := github.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating GitHub client")
		}

		// Init service using GitHub client.
		cfg := cfg.GitHub
		lister = github.NewLister(
			client,
			cfg.Username,
			func(lc *github.ListerConfig) {
				lc.Concurrency = cfg.Lister.Concurrency
			},
		)
	}

	// Build validator service.
	var validator repo.ValidatorService
	{
		cfg := cfg.Watcher
		validator = repo.NewWatcher(
			lister,
			cfg.CheckInterval,
			func(wc *repo.WatcherConfig) {
				wc.Logger = log.WithField("component", "repo.Watcher")
			},
		)
	}

	// Build HTML generator.
	var generator vanity.HTMLGenerator
	{
		if generator, err = template.NewHTMLGenerator(); err != nil {
			return errors.Wrap(err, "building HTML generator")
		}
	}

	// Build and run server.
	var srv *server.Server
	{
		cfg := cfg.Server
		if srv, err = server.New(
			generator, validator,
			cfg.BaseURL,
			func(c *server.Config) { c.Logger = log },
		); err != nil {
			return errors.Wrap(err, "creating server")
		}
	}

	// Start server on the specified port.
	err = srv.ListenAndServe(fmt.Sprintf(":%d", serveOpts.Port))
	return errors.Wrap(err, "starting server")
}
