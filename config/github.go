package config

import (
	"github.com/stevenxie/vaingogh/imports/github"
)

// GithubRepoLister builds a github.RepoLister.
func (cfg *Config) GithubRepoLister() *github.RepoLister {
	var (
		gh = &cfg.Github
		rl = github.NewRepoLister(gh.Username)
	)
	rl.SetIsOrg(gh.IsOrg)
	rl.SetAccessToken(gh.Token)
	return rl
}
