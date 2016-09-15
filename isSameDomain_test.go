package sitemap

import (
	"net/url"
	"testing"
)

func TestIsSameDomain(t *testing.T) {
	type isSameDomainSpec struct {
		link     string
		expected bool
	}

	var isSameDomainSpecs = []isSameDomainSpec{
		{"/relative/path", true},
		{"http://origin.com/some/path", true},
		{"http://sub.origin.com/some/path", true},
		{"http://another-origin.com", false},
	}

	var originURL, _ = url.Parse("http://origin.com")

	for _, spec := range isSameDomainSpecs {
		testURL, _ := url.Parse(spec.link)

		if isSameDomain(testURL, originURL) != spec.expected {
			t.Error(
				"For", spec.link,
				"expected", spec.expected,
			)
		}
	}
}
