package main

import (
	"net/url"
	"testing"

	"golang.org/x/net/html"
)

func TestFindAttribute(t *testing.T) {
	type findAttributeSpec struct {
		token         html.Token
		attributeName string
		expected      string
		failMessage   string
	}

	var findAttributeSpecs = []findAttributeSpec{
		{
			html.Token{
				Type: html.StartTagToken,
				Attr: []html.Attribute{
					html.Attribute{Key: "src", Val: "blah"},
				},
			},
			"src",
			"blah",
			"Should return a match when one attribute matches.",
		},
		{
			html.Token{
				Type: html.StartTagToken,
				Attr: []html.Attribute{
					html.Attribute{Key: "src", Val: "first"},
					html.Attribute{Key: "src", Val: "second"},
					html.Attribute{Key: "src", Val: "third"},
				},
			},
			"src",
			"first",
			"Should return the first match when many can match.",
		},
		{
			html.Token{
				Type: html.StartTagToken,
				Attr: []html.Attribute{
					html.Attribute{Key: "src", Val: "blah"},
				},
			},
			"href",
			"",
			"Should return an empty string when no matching attributes are found.",
		},
	}

	for _, spec := range findAttributeSpecs {
		attribute := findAttribute(&spec.token, spec.attributeName)

		if attribute != spec.expected {
			t.Error(spec.failMessage, "Got:", attribute)
		}
	}
}

func TestFindParseAndNormalizeLinkAttribute(t *testing.T) {
	var originURL, _ = url.Parse("http://origin.com")

	type findParseAndNormalizeLinkAttributeSpec struct {
		attributeName string
		expected      string
		err           bool
	}

	var findParseAndNormalizeLinkAttributeSpecs = []findParseAndNormalizeLinkAttributeSpec{
		{"src-one", "http://somewhere.com/", false},
		{"src-two", "http://somewhere.com/a/", false},
		{"src-three", "http://origin.com/a/", false},
		{"src-four", "", true},
		{"src-five", "", true},
	}

	var token = html.Token{
		Type: html.StartTagToken,
		Attr: []html.Attribute{
			html.Attribute{Key: "src-one", Val: "http://somewhere.com/"},
			html.Attribute{Key: "src-two", Val: "http://somewhere.com/a/b/../"},
			html.Attribute{Key: "src-three", Val: "/a/b/../"},
			html.Attribute{Key: "src-four", Val: ""},
			html.Attribute{Key: "src-five", Val: "#$%#$"},
		},
	}

	for _, spec := range findParseAndNormalizeLinkAttributeSpecs {
		u, err := findParseAndNormalizeLinkAttribute(originURL, &token, spec.attributeName)

		if !spec.err && err != nil {
			t.Error("Expected URL for", u, "but got error", err)
		} else if spec.err && err == nil {
			t.Error("Expected error but got URL for", u)
		} else if u.String() != spec.expected {
			t.Error("Expected URL for", spec.expected, "but got", u.String())
		}
	}
}

func TestFollow(t *testing.T) {
	followToken := html.Token{
		Type: html.StartTagToken,
		Attr: []html.Attribute{
			html.Attribute{Key: "rel", Val: "x"},
			html.Attribute{Key: "rel", Val: "y"},
			html.Attribute{Key: "rel", Val: "z"},
		},
	}

	nofollowToken := html.Token{
		Type: html.StartTagToken,
		Attr: []html.Attribute{
			html.Attribute{Key: "rel", Val: "x"},
			html.Attribute{Key: "rel", Val: "nofollow"},
			html.Attribute{Key: "rel", Val: "z"},
		},
	}

	if !follow(&followToken) {
		t.Error("Expected token with no rel=nofollow to result in true")
	}

	if follow(&nofollowToken) {
		t.Error("Expected token with rel=nofollow to result in false")
	}
}
