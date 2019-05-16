package config

import (
	"github.com/stevenxie/vaingogh/imports/github"
)

// BuildGithubRepoLister builds a preconfigured github.RepoLister.
func (cfg *Config) BuildGithubRepoLister() *github.RepoLister {
	var (
		gh = &cfg.Github
		rl = github.NewRepoLister(gh.Username)
	)
	rl.SetIsOrg(gh.IsOrg)
	rl.SetAccessToken(gh.Token)
	rl.SetConcurrency(gh.Concurrency)
	return rl
}
