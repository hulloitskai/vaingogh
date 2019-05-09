package github

import (
	"encoding/json"
	"fmt"
	"net/http"

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
}

// NewRepoLister creates a new RepoLister that lists repositories for the
// specified username.
//
// If username corresponds to an organization account, then
// (*RepoLister).SetIsOrg(true) should be called in order to specify the
// appropriate endpoint.
func NewRepoLister(username string) *RepoLister {
	return &RepoLister{
		httpc:    http.DefaultClient,
		username: username,
	}
}

// SetIsOrg sets whether or not the username used by the RepoLister belongs
// to an organization.
func (rl *RepoLister) SetIsOrg(isOrg bool) { rl.isOrg = isOrg }

// SetHTTPClient sets the RepoLister's internal http.Client.
func (rl *RepoLister) SetHTTPClient(hc *http.Client) { rl.httpc = hc }

//revive:disable-line
func (rl *RepoLister) ListRepoNames() ([]string, error) {
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

	names := make([]string, 0, len(repos))
	for _, repo := range repos {
		// Filter only repos that have Go in it.
		if res, err = rl.httpc.Get(repo.LanguagesURL); err != nil {
			return nil, errors.Errorf(
				"github: get languages for '%s': %w",
				repo.FullName,
				err,
			)
		}

		dec = json.NewDecoder(res.Body)
		var langs map[string]int
		if err = dec.Decode(&langs); err != nil {
			return nil, errors.Errorf("github: decoding respose as JSON: %w", err)
		}
		if _, ok := langs["Go"]; !ok {
			continue
		}

		names = append(names, "github.com/"+repo.FullName)
	}
	return names, nil
}
