package repo

type (
	// A ListerService lists the Go repositories from VCS platforms like GitHub.
	ListerService interface {
		ListGoRepos() ([]string, error)
		DeriveRepoFullName(partial string) (repo string)
	}

	// A ValidatorService can validate repos (in terms of whether or they not
	// they are an existing repo that contains Go).
	ValidatorService interface {
		IsRepoValid(repo string) (bool, error)
		DeriveRepoFullName(partial string) (repo string)
	}
)
