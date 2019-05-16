package config

import (
	"github.com/stevenxie/vaingogh/imports"
)

// BuildRepoWatcher builds a preconfigured imports.RepoWatcher.
func (cfg *Config) BuildRepoWatcher(
	lister imports.RepoLister,
) *imports.RepoWatcher {
	return imports.NewRepoWatcher(lister, cfg.CheckInterval)
}
