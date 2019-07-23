package main

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"

	"go.stevenxie.me/vaingogh/config"
	"go.stevenxie.me/vaingogh/repo"
	"go.stevenxie.me/vaingogh/repo/github"
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

	// Build repo service.
	var lister repo.ListerService
	{
		client, err := github.NewClient()
		if err != nil {
			return errors.Wrap(err, "creating GitHub client")
		}

		cfg := cfg.Lister
		lister = github.NewLister(
			client,
			cfg.GitHub.Username,
			func(glc *github.ListerConfig) {
				glc.Concurrency = cfg.Concurrency
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
