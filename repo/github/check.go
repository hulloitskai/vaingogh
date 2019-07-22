package github

import (
	"context"

	"github.com/cockroachdb/errors"
)

func (l *Lister) checkUser() error {
	if l.checked {
		return nil
	}

	{
		user, _, err := l.client.Users.Get(context.Background(), l.user)
		if err != nil {
			return errors.Wrap(err, "getting user details")
		}
		if user != nil {
			goto Checked
		}
	}

	{
		org, _, err := l.client.Organizations.Get(context.Background(), l.user)
		if err != nil {
			return errors.Wrap(err, "getting org details")
		}
		if org != nil {
			l.isOrg = true
			goto Checked
		}
	}

	return ErrUserNotExists

Checked:
	l.checked = true
	return nil
}

// ErrUserNotExists is returned by a RepoLister when it is unable to locate
// either a user or organization with a given username.
var ErrUserNotExists = errors.New("no such user or organization exists")
