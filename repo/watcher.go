package repo

import (
	"sync"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"

	"github.com/stevenxie/api/pkg/stream"
	"github.com/stevenxie/api/pkg/zero"
)

type (
	// A Watcher caches and updates a list of repos at regular intervals.
	// It is safe for concurrent use.
	Watcher struct {
		svc      ListerService
		streamer stream.Streamer
		log      logrus.FieldLogger

		mux   sync.Mutex
		repos []string
		err   error
	}

	// A WatcherConfig configures a Watcher.
	WatcherConfig struct {
		Logger logrus.FieldLogger
	}
)

var (
	_ ListerService    = (*Watcher)(nil)
	_ ValidatorService = (*Watcher)(nil)
)

// NewWatcher creates a new Watcher, which keeps an updated list of
// Go repositories, and checks for updates at regular intervals.
func NewWatcher(
	svc ListerService,
	interval time.Duration,
	opts ...func(*WatcherConfig),
) *Watcher {
	cfg := WatcherConfig{
		Logger: zero.Logger(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	w := &Watcher{
		svc: svc,
		streamer: stream.NewPoller(
			func() (zero.Interface, error) {
				return svc.ListGoRepos()
			},
			interval,
		),
		log: cfg.Logger,
	}
	go w.run()
	return w
}

// ListGoRepos returns the last seen list of Go repos.
func (w *Watcher) ListGoRepos() ([]string, error) {
	w.mux.Lock()
	defer w.mux.Unlock()
	repos := make([]string, len(w.repos))
	copy(repos, w.repos)
	return repos, w.err
}

// IsRepoValid returns true if repo is found in the list of Go repos, and false
// otherwise.
func (w *Watcher) IsRepoValid(repo string) (bool, error) {
	repos, err := w.ListGoRepos()
	if err != nil {
		return false, errors.Wrap(err, "repo: listing Go repos")
	}
	for _, r := range repos {
		if r == repo {
			return true, nil
		}
	}
	return false, nil
}

// DeriveRepoFullName derives the full name of a repo from a partial name.
func (w *Watcher) DeriveRepoFullName(partial string) (repo string) {
	return w.svc.DeriveRepoFullName(partial)
}

func (w *Watcher) run() {
	w.log.Info("Watching for repo list changes.")
	for result := range w.streamer.Stream() {
		var (
			repos []string
			err   error
		)

		switch v := result.(type) {
		case error:
			err = v
			w.log.WithError(err).Error("Failed to list latest Go repos.")
		case []string:
			repos = v
		default:
			w.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Newf("imports: unexpected upstream value '%s'", v)
		}

		w.log.WithFields(logrus.Fields{
			"numRepos": len(repos),
			"err":      err,
		}).Debug("Received updated repo list.")

		w.mux.Lock()
		w.repos = repos
		w.err = err
		w.mux.Unlock()
	}
}

// Stop stops the repo watch cycle.
func (w *Watcher) Stop() { w.streamer.Stop() }
