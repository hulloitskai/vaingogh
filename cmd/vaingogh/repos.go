package main

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"github.com/stevenxie/vaingogh/config"
	"github.com/stevenxie/vaingogh/repo"
	"github.com/stevenxie/vaingogh/repo/github"
)

var reposCmd = &cobra.Command{
	Use:          "repos",
	Short:        "List all your Go repositories (sanity check).",
	RunE:         execRepos,
	Args:         cobra.NoArgs,
	SilenceUsage: true,
}

func execRepos(*cobra.Command, []string) error {
	// Load and validate config file.
	cfg, err := config.Load()
	if err != nil {
		return errors.Wrap(err, "loading config")
	}
	if err = cfg.Validate(); err != nil {
		return errors.Wrap(err, "invalid config")
	}

	// Build and run repo lister.
	var lister repo.GoLister
	{
		client, err := github.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating GitHub client")
		}

		cfg := cfg.GitHub
		lister = github.NewGoLister(
			client,
			cfg.Username,
			func(glc *github.GoListerConfig) {
				glc.Concurrency = cfg.Lister.Concurrency
			},
		)
	}

	repos, err := lister.ListGoRepos()
	if err != nil {
		return err
	}

	// Print repos, one on each line.
	for _, repo := range repos {
		fmt.Println(repo)
	}
	return nil
}
