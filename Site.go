package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"

	"github.com/PuerkitoBio/purell"
)

type normalizeError struct {
	link string
}

func (e normalizeError) Error() string {
	return fmt.Sprintf("Link could not be parsed and normalized: %s", e.link)
}

// parseAndNormalizeLinkAttribute searches for an attribute by name in a Token,
// which is then parsed, resolved against the given origin (for relative URLs)
// and normalized.
func parseAndNormalizeLinkAttribute(origin *url.URL, token *html.Token, attributeName string) (*url.URL, error) {
	link := ""

	for _, a := range token.Attr {
		if a.Key == attributeName {
			link = a.Val
		}
	}

	var u url.URL

	if link == "" {
		return &u, normalizeError{link}
	}

	linkURL, err := url.Parse(link)

	if err != nil {
		return &u, normalizeError{link}
	}

	return url.Parse(purell.NormalizeURL(origin.ResolveReference(linkURL), purell.FlagsSafe))
}

// Site contains information about a site, including its URL, and the URLs of
// other sites it links to, scripts, and images.
type Site struct {
	URL           url.URL
	Links         []url.URL
	NoFollowLinks []url.URL
	Scripts       []url.URL
	Images        []url.URL
}

// urlsToStrings takes a list or URL instances and returns a list of URL strings
// which represent them.
func urlsToStrings(urls *[]url.URL) *[]string {
	var links []string

	for _, u := range *urls {
		links = append(links, u.String())
	}

	return &links
}

// MarshalJSON is a helper method which makes Site instances more JSON friendly.
func (s *Site) MarshalJSON() ([]byte, error) {
	links := append(*urlsToStrings(&s.Links), *urlsToStrings(&s.NoFollowLinks)...)

	return json.Marshal(&struct {
		URL     string   `json:"url"`
		Links   []string `json:"links"`
		Scripts []string `json:"scripts"`
		Images  []string `json:"images"`
	}{
		URL:     s.URL.String(),
		Links:   links,
		Scripts: *urlsToStrings(&s.Scripts),
		Images:  *urlsToStrings(&s.Images),
	})
}

func follow(token *html.Token) bool {
	for _, a := range token.Attr {
		if a.Key == "rel" && a.Val == "nofollow" {
			return false
		}
	}

	return true
}

// Crawl populates the resources of this Site instance by loading the associated
// link and processing the page. After processing the document, links are pushed
// into a channel for processing into the sitemap.
func (s *Site) Crawl() {
	res, err := http.Get(s.URL.String())

	if err != nil {
		return
	}

	defer res.Body.Close()

	reader := html.NewTokenizer(res.Body)

	for {
		tt := reader.Next()

		switch {

		// The document has been read to the end. Push all of the resolved links
		// into the urls channel.
		case tt == html.ErrorToken:

			return

		case tt == html.StartTagToken:
			token := reader.Token()

			switch {
			case token.Data == "a":
				u, err := parseAndNormalizeLinkAttribute(&s.URL, &token, "href")

				if err != nil {
					break
				}

				if follow(&token) {
					s.Links = append(s.Links, *u)
				} else {
					s.NoFollowLinks = append(s.NoFollowLinks, *u)
				}

				break

			case token.Data == "script":
				u, err := parseAndNormalizeLinkAttribute(&s.URL, &token, "src")

				if err == nil {
					s.Scripts = append(s.Scripts, *u)
				}

				break

			case token.Data == "img":
				u, err := parseAndNormalizeLinkAttribute(&s.URL, &token, "src")

				if err == nil {
					s.Images = append(s.Images, *u)
				}

				break
			}
		}
	}
}
