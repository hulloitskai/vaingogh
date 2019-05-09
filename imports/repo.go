package imports

import "time"

// A RepoLister lists the Go repositories from VCS platforms like GitHub.
type RepoLister interface{ ListRepoNames() ([]string, error) }

// A RepoWatcher caches and updates a list of repos at regular intervals.
type RepoWatcher struct {
	repos         []string
	lister        RepoLister
	checkInterval time.Duration

	ticker *time.Ticker
	close  chan struct{}
}

// NewRepoWatcher creates a new RepoWatcher, which checks for new repos
// in accordance with checkInterval.
func NewRepoWatcher(
	lister RepoLister,
	checkInterval time.Duration,
) *RepoWatcher {
	return &RepoWatcher{
		lister:        lister,
		checkInterval: checkInterval,
		close:         make(chan struct{}),
	}
}

// Repos returns the last seen list of repos.
func (rw *RepoWatcher) Repos() []string { return rw.repos }

// Run starts the repo watch cycle. It blocks until the cycle finishes.
func (rw *RepoWatcher) Run() error {
	rw.ticker = time.NewTicker(rw.checkInterval)

	// Run initial list.
	repos, err := rw.lister.ListRepoNames()
	if err != nil {
		return err
	}
	rw.repos = repos

loop:
	for {
		select {
		case <-rw.ticker.C:
			if repos, err = rw.lister.ListRepoNames(); err != nil {
				return err
			}
			rw.repos = repos

		case <-rw.close:
			break loop
		}
	}

	return nil
}

// Stop stops the repo watch cycle.
func (rw *RepoWatcher) Stop() {
	rw.ticker.Stop()
	rw.close <- struct{}{} // trigger closing of watch loop
}
