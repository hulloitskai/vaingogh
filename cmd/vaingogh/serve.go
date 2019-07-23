package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"go.stevenxie.me/api/pkg/cmdutil"
	"go.stevenxie.me/vaingogh/config"
	"go.stevenxie.me/vaingogh/server"

	"go.stevenxie.me/vaingogh/repo"
	repogh "go.stevenxie.me/vaingogh/repo/github"
	"go.stevenxie.me/vaingogh/template"
	tplgh "go.stevenxie.me/vaingogh/template/github"
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
		client, err := repogh.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating GitHub client")
		}

		// Init service using GitHub client.
		cfg := cfg.Lister
		lister = repogh.NewLister(
			client,
			cfg.GitHub.Username,
			func(lc *repogh.ListerConfig) {
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

	// Build page generator.
	var generator template.Generator
	{
		if generator, err = tplgh.NewGenerator(); err != nil {
			return errors.Wrap(err, "building generator")
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
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return errors.Wrap(err, "starting server")
	}
	return nil
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
