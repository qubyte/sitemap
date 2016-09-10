package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/purell"
	"golang.org/x/net/html"
)

// sitemap = {
//   https://qubyte.codes: {
//     link: {
//       https://qubyte.codes: true
//     },
//     script: {
//       ...
//     }
//   }
// }
var sitemap = make(map[string]map[string]map[string]bool)

// Determines if a link belongs to a domain.
func isSameDomain(linkURL, domainURL url.URL) bool {
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

func findAttribute(token html.Token, attributeName string) string {
	for _, a := range token.Attr {
		if a.Key == attributeName { // Haven't found the href yet... Continue.
			return a.Val
		}
	}

	return ""
}

func handleLink(domainURL url.URL, pageLink string, token html.Token) {
	href := findAttribute(token, "href")

	if href == "" {
		return
	}

	linkURL, err := url.Parse(href)

	if err != nil {
		return
	}

	if !isSameDomain(*linkURL, domainURL) { // Mismatching domains. Break.
		appendResource(pageLink, "link", href)

		return
	}

	resolved := purell.NormalizeURL(domainURL.ResolveReference(linkURL), purell.FlagsSafe)

	appendResource(pageLink, "link", resolved)

	if _, ok := sitemap[resolved]; !ok {
		// println(resolved)
		findLinks(resolved) // A new link! Crawl it.
	}

	return
}

func handleScript(domainURL url.URL, pageLink string, token html.Token) {
	src := findAttribute(token, "src")

	if src == "" {
		return
	}

	linkURL, err := url.Parse(src)

	if err != nil {
		return
	}

	resolved := domainURL.ResolveReference(linkURL).String()

	appendResource(pageLink, "script", resolved)
}

func handleImage(domainURL url.URL, pageLink string, token html.Token) {
	src := findAttribute(token, "src")

	if src == "" {
		return
	}

	linkURL, err := url.Parse(src)

	if err != nil {
		return
	}

	resolved := domainURL.ResolveReference(linkURL).String()

	appendResource(pageLink, "image", resolved)
}

func appendResource(page, resourceType, resource string) {
	if _, ok := sitemap[page][resourceType]; ok == false {
		sitemap[page][resourceType] = make(map[string]bool)
	}

	sitemap[page][resourceType][resource] = true
}

// Consumes and tokenizes an HTTP response body.
func findLinks(pageLink string) {
	domainURL, err := url.Parse(pageLink)

	if err != nil {
		return
	}

	if domainURL.IsAbs() == false {
		return
	}

	sitemap[pageLink] = make(map[string]map[string]bool)

	res, err := http.Get(pageLink)

	if err != nil {
		return
	}

	defer res.Body.Close()

	reader := html.NewTokenizer(res.Body)

	for {
		tt := reader.Next()

		switch {
		case tt == html.ErrorToken:
			return

		case tt == html.StartTagToken:
			token := reader.Token()

			switch {
			case token.Data == "a":
				handleLink(*domainURL, pageLink, token)
				break

			case token.Data == "script":
				handleScript(*domainURL, pageLink, token)
				break

			case token.Data == "img":
				handleImage(*domainURL, pageLink, token)
				break
			}
		}
	}
}

func printSiteMap() {
	for key := range sitemap {
		fmt.Println(key)

		for resourceType, resources := range sitemap[key] {
			fmt.Println("\t", resourceType)

			for resource := range resources {
				fmt.Println("\t\t", resource)
			}
		}
	}
}

func main() {
	domain := flag.String("domain", "", "A fully qualified URL for the domain to crawl.")
	flag.Parse()

	if *domain == "" {
		fmt.Println("A domain to crawl is required.")
	}

	fmt.Printf("Domain: %s\n", *domain)

	findLinks(*domain)

	printSiteMap()
}
