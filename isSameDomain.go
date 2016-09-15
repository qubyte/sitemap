package sitemap

import (
	"net/url"
	"strings"
)

// isSameDomain is a crude check that a given linkURL is in the same domain or a
// subdomain of domainURL.
func isSameDomain(linkURL, domainURL *url.URL) bool {
	// Relative links are for the same domain.
	if linkURL.Host == "" {
		return true
	}

	if linkURL.Host == domainURL.Host {
		return true
	}

	// Links to subdomains of this domain still count.
	if strings.HasSuffix(linkURL.Host, "."+domainURL.Host) {
		return true
	}

	return false
}
