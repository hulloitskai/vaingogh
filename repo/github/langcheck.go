package github

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v27/github"
)

type langCheckResult struct {
	Repo  *github.Repository
	HasGo bool
}

func (gl *GoLister) langCheckWorker(
	ctx context.Context,
	repos <-chan *github.Repository,
	results chan<- langCheckResult,
) error {
	for repo := range repos {
		// Make request with the worker context, so that it will be cancelled if the
		// worker context is cancelled.
		languages, _, err := gl.client.Repositories.ListLanguages(
			ctx,
			repo.GetOwner().GetLogin(),
			repo.GetName(),
		)
		if err != nil {
			return errors.Wrapf(err, "listing languages for '%s'", repo.GetFullName())
		}

		result := langCheckResult{Repo: repo}
		_, result.HasGo = languages["Go"]
		results <- result
	}
	return nil
}
