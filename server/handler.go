package server

import (
	"io"
	"net/http"
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/sirupsen/logrus"
)

func (srv *Server) handler(log logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var status int
		if err := func() error {
			// Derive repo from URL.
			address := r.Host + r.URL.Path
			partial := strings.TrimPrefix(address, srv.baseURL)
			partial = strings.Trim(partial, "/")
			repo := srv.validator.DeriveRepoFullName(partial)

			// Ensure repository is valid.
			valid, err := srv.validator.IsRepoValid(repo)
			if err != nil {
				log.WithError(err).Error("Failure while checking repo validity.")
				return errors.Wrap(err, "checking repo validity")
			}
			if !valid {
				status = http.StatusNotFound
				return errors.New("invalid repo")
			}

			// Generate HTML page.
			html, err := srv.generator.GenerateHTML(address, repo)
			if err != nil {
				return errors.Wrap(err, "generating HTML")
			}

			// Send HTML response.
			w.Header().Set("Content-Type", "text/html; charset=UTF-8")
			if _, err := io.WriteString(w, html); err != nil {
				return errors.Wrap(err, "writing HTML response")
			}

			return nil
		}(); err != nil { // catch error and write as text
			w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
			if status != 0 {
				w.WriteHeader(status)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			io.WriteString(w, err.Error())
		}
	}
}