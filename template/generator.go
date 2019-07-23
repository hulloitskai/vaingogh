package template

import (
	"strings"

	"go.stevenxie.me/vaingogh/pkg/urlutil"
)

// An Generator can generate an vanity imports HTML page.
type Generator interface {
	GenerateHTML(prefix, address, repo string) (html string, err error)
}

// WithSanitizer wraps a Generator with an input-sanitization layer.
func WithSanitizer(g Generator) Generator {
	return sanitizedGenerator{g}
}

type sanitizedGenerator struct {
	Generator
}

func (sg sanitizedGenerator) GenerateHTML(prefix, address, repo string) (
	html string, err error) {
	prefix = urlutil.StripProtocol(prefix)
	address = urlutil.StripProtocol(address)
	repo = strings.Trim(repo, "/")
	return sg.Generator.GenerateHTML(prefix, address, repo)
}
