package template

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/vaingogh/template"
)

const defaultBaseURL = "https://github.com"

type (
	// A Generator can generate an HTML page for a vanity import whose source
	// originates from GitHub.
	Generator struct {
		templator *template.Templator
		baseURL   string
	}

	// GeneratorConfig configures a Generator.
	GeneratorConfig struct {
		Template string
		BaseURL  string // defaults to "https://github.com"
	}
)

var _ template.Generator = (*Generator)(nil)

// NewGenerator creates a new Generator. It is safe for concurrent use.
func NewGenerator(opts ...func(*GeneratorConfig)) (template.Generator, error) {
	cfg := GeneratorConfig{
		BaseURL: defaultBaseURL,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Build templator.
	templator, err := template.NewTemplator(
		func(tc *template.TemplatorConfig) {
			if cfg.Template != "" {
				tc.Template = cfg.Template
			}
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "github: building templator")
	}

	return template.WithSanitizer(Generator{
		templator: templator,
		baseURL:   cfg.BaseURL,
	}), nil
}

// GenerateHTML generates an HTML page for a vanity import.
func (gen Generator) GenerateHTML(prefix, address, repo string) (html string,
	err error) {
	sourceURL := fmt.Sprintf("%s/%s", gen.baseURL, repo)

	// Fill out template data.
	data := template.TemplatorData{
		Prefix:        prefix,
		Address:       address,
		VCSType:       "git",
		ImportURL:     sourceURL,
		SourceURL:     sourceURL,
		SourceTreeURL: fmt.Sprintf("%s/tree/master{/dir}", sourceURL),
		SourceBlobURL: fmt.Sprintf("%s/blob/master{/dir}/{file}#L{line}",
			sourceURL),
	}

	// Execute template using data.
	return gen.templator.TemplateHTML(data)
}
