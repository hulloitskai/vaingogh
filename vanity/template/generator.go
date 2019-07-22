package template

import (
	"fmt"
	"html/template"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/stevenxie/vaingogh/pkg/urlutil"
	"github.com/stevenxie/vaingogh/vanity"
)

type (
	// An HTMLGenerator can generator an HTML page for a vanity import address.
	HTMLGenerator struct {
		mux sync.Mutex
		tpl *template.Template

		importURLPrefix string
		sourceURLPrefix string
	}

	// HTMLGeneratorConfig configures a HTMLGenerator.
	HTMLGeneratorConfig struct {
		HTMLTemplate    string // defaults to `defaultRawTpl`
		ImportURLPrefix string // defaults to https://github.com
		SourceURLPrefix string // defaults to https://github.com
	}

	// HTMLGeneratorContext contains data fields that can be used by HTMLGenerator
	// templates.
	HTMLGeneratorContext struct {
		Prefix    string
		Address   string
		ImportURL string
		SourceURL string
	}
)

var _ vanity.HTMLGenerator = (*HTMLGenerator)(nil)

// NewHTMLGenerator creates a new HTMLGenerator. It is safe for concurrent use.
//
// Username is used to construct import and source URLs.
func NewHTMLGenerator(opts ...func(*HTMLGeneratorConfig)) (*HTMLGenerator,
	error) {
	var (
		repoPrefix = "https://github.com"
		cfg        = HTMLGeneratorConfig{
			HTMLTemplate:    defaultRawTpl,
			ImportURLPrefix: repoPrefix,
			SourceURLPrefix: repoPrefix,
		}
	)
	for _, opt := range opts {
		opt(&cfg)
	}

	// Parse template.
	tpl, err := template.New("html").Parse(cfg.HTMLTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "template: parsing HTML template")
	}

	return &HTMLGenerator{
		tpl:             tpl,
		importURLPrefix: cfg.ImportURLPrefix,
		sourceURLPrefix: cfg.SourceURLPrefix,
	}, nil
}

// GenerateHTML generates an HTML page for a vanity import.
func (gen *HTMLGenerator) GenerateHTML(prefix, address, repo string) (
	html string, err error) {
	// Normalize prefix and address (strip protocol).
	prefix = urlutil.StripProtocol(prefix)
	address = urlutil.StripProtocol(address)

	// Protect internal values access.
	gen.mux.Lock()
	defer gen.mux.Unlock()

	// Create context.
	ctx := HTMLGeneratorContext{
		Prefix:    prefix,
		Address:   address,
		ImportURL: gen.generateURL(gen.importURLPrefix, repo),
		SourceURL: gen.generateURL(gen.sourceURLPrefix, repo),
	}

	// Build HTML using template and context.
	var builder strings.Builder
	if err = gen.tpl.Execute(&builder, ctx); err != nil {
		return "", errors.Wrap(err, "template: building HTML from template")
	}
	return builder.String(), nil
}

func (gen *HTMLGenerator) generateURL(prefix, repo string) string {
	return fmt.Sprintf("%s/%s", prefix, repo)
}
