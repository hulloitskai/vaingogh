package github

import (
	"context"

	"golang.org/x/oauth2"
)

// SetAccessToken sets a personal access token to be used by the RepoLister.
//
// This will increase rate limits for GitHub requests.
func (rl *RepoLister) SetAccessToken(token string) {
	var (
		ts     = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		client = oauth2.NewClient(context.Background(), ts)
	)
	rl.SetHTTPClient(client)
}
