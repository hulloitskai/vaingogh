package server

import (
	"context"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/stevenxie/api/pkg/zero"

	"github.com/stevenxie/vaingogh/pkg/urlutil"
	"github.com/stevenxie/vaingogh/repo"
	"github.com/stevenxie/vaingogh/template"
)

type (
	// Server responds to vanity import URL requests.
	Server struct {
		httpsrv *http.Server
		log     logrus.FieldLogger

		generator template.Generator
		validator repo.ValidatorService
		baseURL   string
	}

	// Config configures a Server.
	Config struct {
		HTTPServer *http.Server
		Logger     logrus.FieldLogger
	}
)

// New creates a new Server.
func New(
	generator template.Generator,
	validator repo.ValidatorService,
	baseURL string,
	opts ...func(*Config),
) (*Server, error) {
	cfg := Config{
		HTTPServer: new(http.Server),
		Logger:     zero.Logger(),
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Normalize baseURL.
	baseURL = urlutil.StripProtocol(baseURL)
	baseURL = strings.Trim(baseURL, "/")

	return &Server{
		generator: generator,
		validator: validator,
		httpsrv:   cfg.HTTPServer,
		log:       cfg.Logger,
		baseURL:   baseURL,
	}, nil
}

// ListenAndServe listens and serves respones to network requests on the
// specified address.
func (srv *Server) ListenAndServe(addr string) error {
	var (
		httpsrv = srv.httpsrv
		hlog    = srv.log.WithField("component", "handler")
	)

	// Configure HTTP server.
	httpsrv.Handler = srv.handler(hlog)
	httpsrv.Addr = addr

	srv.log.WithField("addr", addr).Info("Listening for connections...")
	return httpsrv.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (srv *Server) Shutdown(ctx context.Context) error {
	return srv.httpsrv.Shutdown(ctx)
}
