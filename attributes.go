package main

import (
	"errors"
	"net/url"

	"github.com/PuerkitoBio/purell"
	"golang.org/x/net/html"
)

// FindAttribute finds all attributes of a token which match a given attribute
// name.
func findAttribute(token *html.Token, attributeName string) string {
	for _, a := range token.Attr {
		if a.Key == attributeName {
			return a.Val
		}
	}

	return ""
}

// parseAndNormalizeLinkAttribute searches for an attribute by name in a Token,
// which is then parsed, resolved against the given origin (for relative URLs)
// and normalized.
func findParseAndNormalizeLinkAttribute(origin *url.URL, token *html.Token, attributeName string) (*url.URL, error) {
	link := findAttribute(token, attributeName)

	var u url.URL

	if link == "" {
		return &u, errors.New("Unable to parse empty link.")
	}

	linkURL, err := url.Parse(link)

	if err != nil {
		return &u, err
	}

	return url.Parse(purell.NormalizeURL(origin.ResolveReference(linkURL), purell.FlagsSafe))
}

// follow searches the attributes of a token for rel=nofollow.
func follow(token *html.Token) bool {
	for _, a := range token.Attr {
		if a.Key == "rel" && a.Val == "nofollow" {
			return false
		}
	}

	return true
}
