package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strings"
	"sync"
)

var waiting = 0

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
		log.Println("SUBDOMAIN:", linkURL.Host)
		return true
	}

	return false
}

// worker accepts URLs through a channel, makes a Site instance using it, and
// adds the site to the sitemap. The Site instance is then instructed to crawl,
// and links it discovers are pushed into another channel. Decrements the
// WaitGroup when a site has been crawled.
func worker(id int, startDomain *url.URL, sitemap *SiteMap, wg *sync.WaitGroup, queue chan *Site) {
	for s := range queue {
		s.Crawl()

		// Find new links within this domain for the crawled site, and add them to the queue.
		for _, u := range s.Links {
			if !isSameDomain(&u, startDomain) {
				continue
			}

			site := Site{URL: u}

			wasSet := sitemap.SetOnce(&site)

			if !wasSet {
				continue
			}

			wg.Add(1)

			go func() {
				queue <- &site
			}()
		}

		wg.Done()
	}
}

func checkFlags() (*url.URL, int) {
	domain := flag.String("domain", "", "A fully qualified URL for the domain to crawl.")
	jobs := flag.Int("jobs", 1, "Number of simultaneous requests to allow.")
	flag.Parse()

	if *domain == "" {
		log.Fatalln("A domain to crawl is required.")
	}

	if *jobs < 1 {
		log.Fatal("The job count cannot be less than 1.")
	}

	linkURL, err := url.Parse(*domain)

	if err != nil || !linkURL.IsAbs() {
		log.Fatal("domain must be a fully qualified URL.")
	}

	return linkURL, *jobs
}

func main() {
	linkURL, jobs := checkFlags()

	log.Println("Domain:", linkURL.String())
	log.Println("Workers:", jobs)

	// Sites to crawl will be pushed into this stream.
	sites := make(chan *Site, jobs)

	// The number of pages to crawl is initially unknown, so a WaitGroup is used
	// to block the completion of main until the WaitGroup is empty.
	var wg sync.WaitGroup

	sitemap := NewSiteMap()

	// Start a bunch of workers.
	for w := 0; w < jobs; w++ {
		go worker(w, linkURL, sitemap, &wg, sites)
	}

	// Since we don't know how large a sitemap will be, we use a WaitGroup. Each
	// new site to crawl will increment it, and once crawled will decrement it.
	// Start the site crawl by incrementing and pushing the given domain into
	// the sites channel.
	wg.Add(1)

	site := Site{URL: *linkURL}
	sitemap.SetOnce(&site)
	sites <- &site

	// Wait until there are no sites left to crawl.
	wg.Wait()

	close(sites)

	result, _ := json.Marshal(sitemap)

	fmt.Println(string(result))
}
