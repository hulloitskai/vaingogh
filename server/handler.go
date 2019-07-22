package server

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/sirupsen/logrus"
	"github.com/stevenxie/vaingogh/internal/info"
	serverinfo "github.com/stevenxie/vaingogh/server/internal/info"
)

func (srv *Server) handler(log logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var status int
		if err := func() error {
			// Respond with server info upon root path request.
			if r.URL.Path == "/" {
				w.Header().Set("Content-Type", "application/json")
				err := json.NewEncoder(w).Encode(struct {
					Name        string `json:"name"`
					Version     string `json:"version"`
					Environment string `json:"environment,omitempty"`
				}{
					Name:        serverinfo.Name,
					Version:     info.Version,
					Environment: os.Getenv("GOENV"),
				})
				return errors.Wrap(err, "encoding info response")
			}

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
