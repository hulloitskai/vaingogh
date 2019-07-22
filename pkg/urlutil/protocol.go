package urlutil

import (
	"strings"
)

// StripProtocol removes the protocol portion of a URL string.
func StripProtocol(url string) string {
	protoi := strings.Index(url, "://")
	if protoi > -1 {
		return url[protoi+3:]
	}
	return url
}
