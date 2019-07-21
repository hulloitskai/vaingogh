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
	Watcher struct {
		streamer stream.Streamer
		log      *logrus.Logger

		mux   sync.Mutex
		repos []string
		err   error
	}

	// A WatcherConfig configures a Watcher.
	WatcherConfig struct {
		Logger *logrus.Logger
	}

	// A GoLister lists the Go repositories from VCS platforms like GitHub.
	GoLister interface {
		ListGoRepos() ([]string, error)
	}
)

// NewWatcher creates a new Watcher, which keeps an updated list of
// Go repositories, and checks for updates at regular intervals.
func NewWatcher(
	lister GoLister,
	interval time.Duration,
	opts ...func(*WatcherConfig),
) *Watcher {
	cfg := WatcherConfig{
		Logger: zero.Logger(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}
	return &Watcher{
		streamer: stream.NewPoller(
			func() (zero.Interface, error) {
				return lister.ListGoRepos()
			},
			interval,
		),
		log: cfg.Logger,
	}
}

// Repos returns the last seen list of repos.
func (rw *Watcher) Repos() ([]string, error) {
	rw.mux.Lock()
	defer rw.mux.Unlock()
	return rw.repos, rw.err
}

func (rw *Watcher) run() {
	for result := range rw.streamer.Stream() {
		var (
			repos []string
			err   error
		)

		switch v := result.(type) {
		case error:
			err = v
			rw.log.WithError(err).Error("Failed to list latest repos.")
		case []string:
			repos = v
		default:
			rw.log.WithField("value", v).Error("Unexpected value from upstream.")
			err = errors.Newf("imports: unexpected upstream value '%s'", v)
		}

		rw.mux.Lock()
		rw.repos = repos
		rw.err = err
		rw.mux.Unlock()
	}
}

// Stop stops the repo watch cycle.
func (rw *Watcher) Stop() { rw.streamer.Stop() }
