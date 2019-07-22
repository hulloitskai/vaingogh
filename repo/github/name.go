package github

import "fmt"

// DeriveRepoFullName derives the full name of a repo from a partial name.
func (l *Lister) DeriveRepoFullName(partial string) (repo string) {
	return fmt.Sprintf("%s/%s", l.user, partial)
}
