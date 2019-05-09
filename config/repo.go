package config

import (
	"github.com/stevenxie/vaingogh/imports"
)

// RepoWatcher builds an imports.RepoWatcher, configured using Config.
func (cfg *Config) RepoWatcher(lister imports.RepoLister) *imports.RepoWatcher {
	return imports.NewRepoWatcher(lister, cfg.CheckInterval)
}
