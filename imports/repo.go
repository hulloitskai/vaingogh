package imports

import (
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// A RepoLister lists the Go repositories from VCS platforms like GitHub.
type RepoLister interface{ ListGoRepos() ([]string, error) }

// A RepoWatcher caches and updates a list of repos at regular intervals.
type RepoWatcher struct {
	repos []string
	mux   sync.Mutex

	lister        RepoLister
	checkInterval time.Duration

	ticker *time.Ticker
	close  chan struct{}

	logger zerolog.Logger
}

// NewRepoWatcher creates a new RepoWatcher, which keeps an updated list of
// Go repositories, and checks for updates in interals.
func NewRepoWatcher(
	lister RepoLister,
	checkInterval time.Duration,
) *RepoWatcher {
	return &RepoWatcher{
		lister:        lister,
		checkInterval: checkInterval,
		close:         make(chan struct{}),
		logger:        zerolog.Nop(),
	}
}

// Repos returns the last seen list of repos.
func (rw *RepoWatcher) Repos() []string {
	rw.mux.Lock()
	defer rw.mux.Unlock()
	return rw.repos
}

// SetLogger sets RepoWatcher's logger.
func (rw *RepoWatcher) SetLogger(l zerolog.Logger) { rw.logger = l }

// Start begins the repo watch cycle. It blocks for the first repo-listing
// operation, and runs the rest asynchronously.
//
// If the initial repo-listing operation fails, it will return the corresponding
// error.
//
// Future listing operation errors will be logged.
func (rw *RepoWatcher) Start() error {
	rw.ticker = time.NewTicker(rw.checkInterval)

	// Run initial list.
	repos, err := rw.lister.ListGoRepos()
	if err != nil {
		return err
	}
	rw.repos = repos

	go func() {
	loop:
		for {
			select {
			case <-rw.ticker.C:
				repos, err := rw.lister.ListGoRepos()
				if err != nil {
					rw.l().Err(err).Msg("Failed to list Go repositories.")
				}

				rw.mux.Lock()
				rw.repos = repos
				rw.mux.Unlock()

			case <-rw.close:
				break loop
			}
		}
	}()

	return nil
}

// Stop stops the repo watch cycle.
func (rw *RepoWatcher) Stop() {
	rw.ticker.Stop()
	rw.close <- struct{}{} // trigger closing of watch loop
}

func (rw *RepoWatcher) l() *zerolog.Logger { return &rw.logger }
