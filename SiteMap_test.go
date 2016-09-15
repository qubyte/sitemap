package main

import (
	"net/url"
	"testing"
)

func TestNewSiteMap(t *testing.T) {
	u, _ := url.Parse("http://some-place.com")

	sitemap := NewSiteMap(u)

	if sitemap.origin != u {
		t.Error("expected origin to be", u, "but was", sitemap.origin)
	}
}

func TestSetOnce(t *testing.T) {
	u, _ := url.Parse("http://some-place.com")

	sitemap := NewSiteMap(u)

	n, _ := url.Parse("http://new-place.com")
	m, _ := url.Parse("http://new-place.com")

	newsite := Site{URL: *n}
	anothernewsite := Site{URL: *m}

	sitemap.setOnce(&newsite)
	sitemap.setOnce(&anothernewsite)

	siteref := sitemap.Sites["http://new-place.com"]

	if siteref != &newsite || siteref == &anothernewsite {
		t.Error("Site should only be set once for a give URL.")
	}
}
