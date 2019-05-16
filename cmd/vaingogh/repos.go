package main

import (
	"fmt"

	"github.com/spf13/cobra"

	errors "golang.org/x/xerrors"

	"github.com/stevenxie/vaingogh/config"
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
		return errors.Errorf("loading config: %w", err)
	}
	if err = cfg.Validate(); err != nil {
		return errors.Errorf("invalid config: %w", err)
	}

	// Build and run repo lister.
	lister := cfg.BuildGithubRepoLister()
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
