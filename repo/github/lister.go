package github

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v27/github"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"go.stevenxie.me/api/pkg/zero"
	"go.stevenxie.me/vaingogh/repo"
)

type (
	// A Lister can list GitHub repos containing Go for a particular
	// user / organization.
	Lister struct {
		client      *github.Client
		concurrency int

		user    string
		checked bool
		isOrg   bool
	}

	// A ListerConfig configures a Lister.
	ListerConfig struct {
		Concurrency int
	}
)

var _ repo.ListerService = (*Lister)(nil)

// NewLister creates a new Lister that lists repositories for the
// specified user.
func NewLister(
	c *github.Client,
	username string,
	opts ...func(*ListerConfig),
) *Lister {
	cfg := ListerConfig{
		Concurrency: 5,
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Lister{
		client:      c,
		user:        username,
		concurrency: cfg.Concurrency,
	}
}

// ListGoRepos lists all the repos that use Go.
func (l *Lister) ListGoRepos() ([]string, error) {
	if err := l.checkUser(); err != nil {
		return nil, errors.Wrap(err, "github: checking user type")
	}

	// List all repos by l.user.
	var (
		repos []*github.Repository
		err   error
	)
	if l.isOrg {
		repos, _, err = l.client.Repositories.ListByOrg(
			context.Background(),
			l.user,
			&github.RepositoryListByOrgOptions{Type: "public"},
		)
	} else {
		repos, _, err = l.client.Repositories.List(
			context.Background(),
			l.user,
			&github.RepositoryListOptions{
				Visibility:  "public",
				Affiliation: "owner",
				Sort:        "updated",
			},
		)
	}
	if err != nil {
		return nil, err
	}

	// Prepare to consolidate async work results.
	var (
		results = make(chan string)
		gorepos = make([]string, 0, len(repos))
		done    = make(chan zero.Struct)
	)
	go func(results <-chan string, done chan<- zero.Struct) {
		for result := range results {
			gorepos = append(gorepos, result)
		}
		done <- zero.Empty()
	}(results, done)

	// Create semaphore with a maximum weight of min(len(repos), rl.concurrency).
	numWorkers := l.concurrency
	if len(repos) < numWorkers {
		numWorkers = len(repos)
	}
	var (
		sem             = semaphore.NewWeighted(int64(numWorkers))
		group, groupctx = errgroup.WithContext(context.Background())
	)
	for _, repo := range repos {
		// Wait until a semaphore is acquired.
		if err = sem.Acquire(groupctx, 1); err != nil {
			return nil, errors.Wrap(err, "acquiring semaphore")
		}

		// Check repo languages to see if it contains 'Go'.
		func(repo *github.Repository, svc *github.RepositoriesService) {
			group.Go(func() error {
				defer sem.Release(1)

				// Make request with groupctx, so that it will be cancelled if the group
				// is cancelled.
				languages, _, err := svc.ListLanguages(
					groupctx,
					repo.GetOwner().GetLogin(),
					repo.GetName(),
				)
				if err != nil {
					return errors.Wrapf(err, "listing languages for '%s'",
						repo.GetFullName())
				}

				// Send repo name to results channel if it language analysis results
				// contain 'Go'.
				if _, ok := languages["Go"]; ok {
					results <- repo.GetFullName()
				}
				return nil
			})
		}(repo, l.client.Repositories)
	}

	// Wait for errgroup to finish.
	if err = group.Wait(); err != nil {
		return nil, errors.Wrap(err, "github: checking languages")
	}
	close(results)

	// Wait for results to finish consolidating.
	<-done

	return gorepos, nil
}

// DeriveRepoFullName derives the full name of a repo from a partial name.
func (l *Lister) DeriveRepoFullName(partial string) (repo string) {
	return fmt.Sprintf("%s/%s", l.user, partial)
}
