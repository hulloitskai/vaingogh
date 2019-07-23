package template

import (
	"html/template"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
)

type (
	// A Templator generalizes the process for generating vanity import pages
	// from a template.
	//
	// It is safe for concurrent use.
	Templator struct {
		mux     sync.Mutex
		tpl     *template.Template
		builder strings.Builder
	}

	// TemplatorConfig configures a Templator.
	TemplatorConfig struct {
		Template string // defaults to `defaultRawTpl`
	}

	// TemplatorData contains fields that can be used to fill out the
	// vanity mport page template.
	TemplatorData struct {
		Prefix        string
		Address       string
		VCSType       string
		ImportURL     string
		SourceURL     string
		SourceTreeURL string
		SourceBlobURL string
	}
)

// NewTemplator creates a new Templator. It is safe for concurrent use.
func NewTemplator(opts ...func(*TemplatorConfig)) (*Templator, error) {
	cfg := TemplatorConfig{
		Template: defaultTemplate,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	// Parse template.
	tpl, err := template.New("html").Parse(cfg.Template)
	if err != nil {
		return nil, errors.Wrap(err, "vanity: parsing HTML template")
	}

	return &Templator{tpl: tpl}, nil
}

// TemplateHTML generates an HTML page for a vanity import.
func (tplr *Templator) TemplateHTML(data TemplatorData) (html string,
	err error) {
	// Protect against concurrent access.
	tplr.mux.Lock()
	defer tplr.mux.Unlock()

	// Template HTML using data.
	defer tplr.builder.Reset()
	if err = tplr.tpl.Execute(&tplr.builder, &data); err != nil {
		return "", err
	}
	return tplr.builder.String(), nil
}
