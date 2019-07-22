package vanity

// An HTMLGenerator can generator an HTML page for a vanity import address.
type HTMLGenerator interface {
	GenerateHTML(prefix, address, repo string) (html string, err error)
}
