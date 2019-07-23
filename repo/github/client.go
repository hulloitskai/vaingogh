package github

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v27/github"
	"golang.org/x/oauth2"
)

// Namespace is the package namespace, used for things like envvars.
const Namespace = "github"

// NewClient creates a new GitHub client.
//
// If the environment contains a 'GITHUB_TOKEN' variable, then an authenticated
// client will be created; otherwise, the client will be unauthenticated.
func NewClient(opts ...func(*ClientConfig)) (*github.Client, error) {
	client := new(http.Client)

	// Create authenticated http.Client if 'GITHUB_TOKEN' is set.
	if token := os.Getenv(strings.ToUpper(Namespace) + "_TOKEN"); token != "" {
		source := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		client = oauth2.NewClient(context.Background(), source)
	}

	// Create config and apply opts.
	cfg := ClientConfig{
		HTTPClient: client,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	return github.NewClient(cfg.HTTPClient), nil
}

// ClientConfig configures a github.Client.
type ClientConfig struct {
	HTTPClient *http.Client
}
