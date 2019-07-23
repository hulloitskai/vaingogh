# vaingogh

_A vanity URL generator for your Go packages._

[![Git Tag][tag-img]][tag]
[![Drone: Status][drone-img]][drone]
[![Go Report Card][grp-img]][grp]
[![GoDoc][godoc-img]][godoc]
[![Microbadger][microbadger-img]][microbadger]

## Introduction

This is the lazy man's vanity Go URL generator—a fully-automatic, set-it-and-forget-it solution for people who don't want to have to update a
static file every time they create a new Go repo.

### How It Works

`vaingogh` is a server that continually watches your GitHub account for
an updated list of repos that contain Go. When a request is made to the server,
a repo name is derived, and checked against the list of valid repos; if the
check succeeds, a vanity imports page is generated (see [
`template/default.go`](./template/default.go)) in order to handle both `go get`
and user visits to that webpage (real users will be redirected to the
[GoDoc](http://godoc.org)).

## Usage

```bash
# Create a config file.
$ cat <<EOF > config.yaml
server:
  baseURL: localhost:3000

lister:
  github:
    username: stevenxie
EOF

# Run the server.
$ docker run \
    --rm \
    -v $(PWD)/config.yaml:/etc/vaingogh/config.yaml \
    -p 3000:3000 \
    stevenxie/vaingogh

# Try loading a repo page!
$ curl http://localhost:3000/vaingogh

# (response)
<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="go-import" content="localhost:3000/vaingogh git https://github.com/stevenxie/vaingogh">
    <meta name="go-source" content="localhost:3000/vaingogh https://github.com/stevenxie/vaingogh https://github.com/stevenxie/vaingogh/tree/master{/dir} https://github.com/stevenxie/vaingogh/blob/master{/dir}/{file}#L{line}">
    <meta http-equiv="refresh" content="0; url=https://godoc.org/localhost:3000/vaingogh">
  </head>
  <body>
    Nothing to see here; <a href="https://godoc.org/localhost:3000/vaingogh">move along</a>.
  </body>
</html>
```

Of course, none of these imports will actually work until you run this on an
actual server behind a valid externally-reachable domain. Spin up a server and
change the `server.baseURL` in the config to `go.${YOURDOMAIN}.com` or
somethin', and try it out!

### Authenticated Requests and Rate Limits

In order to increase the rate of requests to the Github API (i.e. with a small
enough `watcher.checkInterval`), authentication must be enabled.

To enable authentication, ensure that the environment variable `GITHUB_TOKEN`
is set with a
[personal access token](https://help.github.com/en/articles/creating-a-personal-access-token-for-the-command-line):

```bash
$ docker run \
    -e GITHUB_TOKEN=... \
    # ...other flags...
    stevenxie/vaingogh
```

[tag]: https://github.com/stevenxie/vaingogh/releases
[tag-img]: https://img.shields.io/github/tag/stevenxie/vaingogh.svg
[drone]: https://ci.stevenxie.me/stevenxie/vaingogh
[drone-img]: https://ci.stevenxie.me/api/badges/stevenxie/vaingogh/status.svg
[grp]: https://goreportcard.com/report/go.stevenxie.me/vaingogh
[grp-img]: https://goreportcard.com/badge/go.stevenxie.me/vaingogh
[godoc]: https://godoc.org/go.stevenxie.me/vaingogh
[godoc-img]: https://godoc.org/go.stevenxie.me/vaingogh?status.svg
[microbadger]: https://microbadger.com/images/stevenxie/vaingogh
[microbadger-img]: https://images.microbadger.com/badges/image/stevenxie/vaingogh.svg
