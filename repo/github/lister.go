package github

import (
	"context"

	"go.stevenxie.me/vaingogh/repo"

	"go.stevenxie.me/api/pkg/zero"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/v27/github"
	"golang.org/x/sync/errgroup"
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

	// A ListerConfig configures a Service.
	ListerConfig struct {
		Concurrency int
	}
)

var _ repo.ListerService = (*Lister)(nil)

// NewLister creates a new Service that lists repositories for the
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

	// Init channels.
	var (
		jobs    = make(chan *github.Repository)
		results = make(chan langCheckResult)
	)

	// Start min(len(repos), rl.concurrency) workers.
	var (
		numWorkers      = l.concurrency
		group, groupctx = errgroup.WithContext(context.Background())
	)
	if len(repos) < numWorkers {
		numWorkers = len(repos)
	}
	for i := 0; i < numWorkers; i++ {
		group.Go(func() error {
			return l.langCheckWorker(groupctx, jobs, results)
		})
	}

	// Prepare to collect results into names.
	var (
		names    = make([]string, 0, len(repos))
		done     = make(chan zero.Struct)
		nresults int
	)
	go func(results <-chan langCheckResult, done chan<- zero.Struct) {
		for result := range results {
			nresults++
			if result.HasGo {
				names = append(names, result.Repo.GetFullName())
			}
		}
		done <- zero.Empty()
	}(results, done)

	// Pass repos to language-check workers. This will block if there are fewer
	// workers than repos.
	for _, repo := range repos {
		jobs <- repo
	}
	close(jobs)

	// Wait for workers to finish.
	if err = group.Wait(); err != nil {
		return nil, errors.Wrap(err, "github: checking languages")
	}
	close(results)

	// Wait for results to finish collecting.
	<-done

	// Ensure all jobs were accounted for.
	if nresults != len(repos) {
		return nil, errors.Newf(
			"github: requested language checks for %d repos, but got %d results",
			len(repos), nresults,
		)
	}
	return names, nil
}
