package github

import (
	"context"

	"github.com/cockroachdb/errors"
)

func (gl *GoLister) checkUser() error {
	if gl.checked {
		return nil
	}

	{
		user, _, err := gl.client.Users.Get(context.Background(), gl.user)
		if err != nil {
			return errors.Wrap(err, "getting user details")
		}
		if user != nil {
			goto Checked
		}
	}

	{
		org, _, err := gl.client.Organizations.Get(context.Background(), gl.user)
		if err != nil {
			return errors.Wrap(err, "getting org details")
		}
		if org != nil {
			gl.isOrg = true
			goto Checked
		}
	}

	return ErrUserNotExists

Checked:
	gl.checked = true
	return nil
}

// ErrUserNotExists is returned by a RepoLister when it is unable to locate
// either a user or organization with a given username.
var ErrUserNotExists = errors.New("no such user or organization exists")
