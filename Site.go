package main

import (
	"encoding/json"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
)

// Site contains information about a site, including its URL, and the URLs of
// other sites it links to, scripts, and images.
type Site struct {
	URL           url.URL
	Links         []url.URL
	NoFollowLinks []url.URL
	Scripts       []url.URL
	Images        []url.URL
	Videos        []url.URL
	Audio         []url.URL
	CSS           []url.URL
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
		Videos  []string `json:"videos"`
		Audio   []string `json:"audio"`
		CSS     []string `json:"css"`
	}{
		URL:     s.URL.String(),
		Links:   links,
		Scripts: *urlsToStrings(&s.Scripts),
		Images:  *urlsToStrings(&s.Images),
		Videos:  *urlsToStrings(&s.Videos),
		Audio:   *urlsToStrings(&s.Audio),
		CSS:     *urlsToStrings(&s.CSS),
	})
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

	openMediaTags := []string{}

	for {
		tt := reader.Next()

		switch tt {

		// The document has been read to the end. Push all of the resolved links
		// into the urls channel.
		case html.ErrorToken:
			return

		case html.SelfClosingTagToken:
			token := reader.Token()

			switch {
			case token.Data == "link":
				if findAttribute(&token, "type") == "text/css" {
					u, err := findParseAndNormalizeLinkAttribute(&s.URL, &token, "href")

					if err == nil {
						s.CSS = append(s.CSS, *u)
					}
				}

			case token.Data == "img":
				u, err := findParseAndNormalizeLinkAttribute(&s.URL, &token, "src")

				if err == nil {
					s.Images = append(s.Images, *u)
				}

			case token.Data == "source":
				length := len(openMediaTags)

				if length == 0 {
					break
				}

				lastOpenMediaTag := openMediaTags[length-1]
				u, err := findParseAndNormalizeLinkAttribute(&s.URL, &token, "src")

				if err != nil {
					break
				}

				switch lastOpenMediaTag {
				case "video":
					s.Videos = append(s.Videos, *u)

				case "audio":
					s.Audio = append(s.Audio, *u)
				}
			}

		case html.StartTagToken:
			token := reader.Token()

			switch token.Data {
			case "a":
				u, err := findParseAndNormalizeLinkAttribute(&s.URL, &token, "href")

				if err != nil {
					break
				}

				if follow(&token) {
					s.Links = append(s.Links, *u)
				} else {
					s.NoFollowLinks = append(s.NoFollowLinks, *u)
				}

			case "script":
				u, err := findParseAndNormalizeLinkAttribute(&s.URL, &token, "src")

				if err == nil {
					s.Scripts = append(s.Scripts, *u)
				}

			case "video":
				openMediaTags = append(openMediaTags, "video")

				u, err := findParseAndNormalizeLinkAttribute(&s.URL, &token, "src")

				var empty url.URL

				if err == nil && *u != empty {
					s.Videos = append(s.Videos, *u)
				}

			case "audio":
				openMediaTags = append(openMediaTags, "audio")

				u, err := findParseAndNormalizeLinkAttribute(&s.URL, &token, "src")

				var empty url.URL

				if err == nil && *u != empty {
					s.Audio = append(s.Audio, *u)
				}
			}

		case html.EndTagToken:
			token := reader.Token()

			if token.Data == "video" || token.Data == "audio" {
				openMediaTags = openMediaTags[:len(openMediaTags)-1]
			}
		}
	}
}
