package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/stevenxie/api/pkg/cmdutil"

	"github.com/stevenxie/vaingogh/config"
	"github.com/stevenxie/vaingogh/repo"
	"github.com/stevenxie/vaingogh/repo/github"
	"github.com/stevenxie/vaingogh/server"
	"github.com/stevenxie/vaingogh/vanity"
	"github.com/stevenxie/vaingogh/vanity/template"
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

	// Finalizers should be run before the program terminates.
	var finalizers cmdutil.Finalizers
	defer func() {
		if len(finalizers) == 0 {
			return
		}

		// Run finalizers in reverse order.
		log.Info("Running finalizers...")
		for i := len(finalizers) - 1; i >= 0; i-- {
			if ferr := finalizers[i](); ferr != nil {
				log.WithError(ferr).Error("A finalizer failed.")
				if err == nil {
					err = errors.New("one or more finalizers failed")
				}
			}
		}
	}()

	// Build repo service.
	var lister repo.ListerService
	{
		// Initiate GitHub client.
		client, err := github.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating GitHub client")
		}

		// Init service using GitHub client.
		cfg := cfg.Lister
		lister = github.NewLister(
			client,
			cfg.GitHub.Username,
			func(lc *github.ListerConfig) {
				lc.Concurrency = cfg.Concurrency
			},
		)
	}

	// Build validator service.
	var validator repo.ValidatorService
	{
		var (
			cfg     = cfg.Watcher
			watcher = repo.NewWatcher(
				lister,
				cfg.CheckInterval,
				func(wc *repo.WatcherConfig) {
					wc.Logger = log.WithField("component", "repo.Watcher")
				},
			)
		)
		validator = watcher
		finalizers = append(finalizers, func() error {
			watcher.Stop()
			return nil
		})
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

	// Shut down server gracefully upon interrupt.
	go shutdownServerUponInterrupt(srv, log, cfg.Server.ShutdownTimeout)

	// Start server on the specified port.
	err = srv.ListenAndServe(fmt.Sprintf(":%d", serveOpts.Port))
	return errors.Wrap(err, "starting server")
}

func shutdownServerUponInterrupt(
	srv *server.Server,
	log *logrus.Logger,
	timeout *time.Duration,
) {
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	// Wait for interrupt signal.
	<-sig

	const msg = "Received interrupt signal; shutting down."
	if timeout != nil {
		log.WithField("timeout", timeout.String()).Info(msg)
	} else {
		log.Info(msg)
	}

	// Prepare shutdown context.
	ctx := context.Background()
	if timeout != nil {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), *timeout)
		defer cancel()
	}

	if err := srv.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server didn't shut down correctly.")
	}
}
