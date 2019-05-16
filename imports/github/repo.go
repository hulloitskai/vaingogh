package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/sync/errgroup"
	errors "golang.org/x/xerrors"
)

const (
	baseURL  = "https://api.github.com"
	usersURL = baseURL + "/users"
	orgsURL  = baseURL + "/orgs"
)

// A RepoLister can list GitHub repos for a particular user / organization.
type RepoLister struct {
	httpc    *http.Client
	username string
	isOrg    bool

	concurrency int
}

// NewRepoLister creates a new RepoLister that lists repositories for the
// specified username.
//
// If username corresponds to an organization account, then
// (*RepoLister).SetIsOrg(true) should be called in order to specify the
// appropriate endpoint.
func NewRepoLister(username string) *RepoLister {
	return &RepoLister{
		httpc:       http.DefaultClient,
		username:    username,
		concurrency: 5,
	}
}

// SetIsOrg sets whether or not the username used by the RepoLister belongs
// to an organization.
func (rl *RepoLister) SetIsOrg(isOrg bool) { rl.isOrg = isOrg }

// SetHTTPClient sets the RepoLister's internal http.Client.
func (rl *RepoLister) SetHTTPClient(hc *http.Client) { rl.httpc = hc }

// SetConcurrency sets the concurrency at which RepoLister will make network
// requests.
func (rl *RepoLister) SetConcurrency(concurrency int) {
	rl.concurrency = concurrency
}

//revive:disable-line
func (rl *RepoLister) ListGoRepos() ([]string, error) {
	base := usersURL
	if rl.isOrg {
		base = orgsURL
	}

	res, err := rl.httpc.Get(fmt.Sprintf("%s/%s/repos", base, rl.username))
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var (
		dec   = json.NewDecoder(res.Body)
		repos []struct {
			FullName     string `json:"full_name"`
			LanguagesURL string `json:"languages_url"`
		}
	)
	if err = dec.Decode(&repos); err != nil {
		return nil, errors.Errorf("github: decoding response as JSON: %w", err)
	}
	if err = res.Body.Close(); err != nil {
		return nil, errors.Errorf("github: closing response body: %w", err)
	}

	// Init channels.
	var (
		jobs    = make(chan langCheckJob)
		results = make(chan langCheckResult, len(repos))
	)

	// Start min(len(repos), rl.concurrency) workers.
	var (
		numWorkers      = rl.concurrency
		group, groupctx = errgroup.WithContext(context.Background())
	)
	if len(repos) < numWorkers {
		numWorkers = len(repos)
	}
	for i := 0; i < numWorkers; i++ {
		group.Go(func() error {
			return rl.langCheckWorker(groupctx, jobs, results)
		})
	}

	// Give workers jobs.
	// This will block if there are fewer workers than jobs.
	for _, repo := range repos {
		jobs <- langCheckJob{Repo: repo.FullName, LangURL: repo.LanguagesURL}
	}
	close(jobs)

	// Wait for workers to finish.
	if err = group.Wait(); err != nil {
		return nil, errors.Errorf("github: checking repo languages: %w", err)
	}
	close(results)

	// Ensure all jobs were acknowledged.
	if len(results) != len(repos) {
		return nil, errors.Errorf(
			"github: requested language checks for %d repos, but got %d results",
			len(repos), len(results),
		)
	}

	// Build repository names list.
	names := make([]string, 0, len(repos))
	for result := range results {
		if result.HasGo {
			names = append(names, result.Repo)
		}
	}
	return names, nil
}
