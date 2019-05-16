package github

import (
	"context"
	"encoding/json"
	"net/http"

	errors "golang.org/x/xerrors"
)

type langCheckJob struct {
	Repo    string
	LangURL string
}

type langCheckResult struct {
	Repo  string
	HasGo bool
}

func (rl *RepoLister) langCheckWorker(
	ctx context.Context,
	jobs <-chan langCheckJob,
	results chan<- langCheckResult,
) error {
	for job := range jobs {
		// Create request with the worker context, so that it will be cancelled
		// if the worker context is cancelled.
		req, err := http.NewRequest("GET", job.LangURL, nil)
		if err != nil {
			return err
		}
		req = req.WithContext(ctx)

		res, err := rl.httpc.Get(job.LangURL)
		if err != nil {
			return err
		}
		defer res.Body.Close()

		var (
			dec   = json.NewDecoder(res.Body)
			langs map[string]int
		)
		if err = dec.Decode(&langs); err != nil {
			return errors.Errorf("decoding response as JSON: %w", err)
		}
		if err = res.Body.Close(); err != nil {
			return errors.Errorf("closing response body", err)
		}

		result := langCheckResult{Repo: job.Repo}
		_, result.HasGo = langs["Go"]
		results <- result
	}
	return nil
}
