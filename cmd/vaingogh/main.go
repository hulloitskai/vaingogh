package main

import (
	"fmt"
	"time"

	ess "github.com/unixpickle/essentials"

	"github.com/stevenxie/vaingogh/config"
)

func main() {
	// Load and validate config file.
	cfg, err := config.Load()
	if err != nil {
		ess.Die("Reading config:", err)
	}
	if err = cfg.Validate(); err != nil {
		ess.Die("Validating config:", err)
	}

	// Build watcher.
	var (
		lister  = cfg.GithubRepoLister()
		watcher = cfg.RepoWatcher(lister)
	)

	// Run watcher.
	go func() {
		if err := watcher.Run(); err != nil {
			ess.Die("Running watcher:", err)
		}
	}()

	time.Sleep(12 * time.Second)

	// List repositories.
	fmt.Println("Repositories:")
	for _, repo := range watcher.Repos() {
		fmt.Printf("\t%s\n", repo)
	}
}
