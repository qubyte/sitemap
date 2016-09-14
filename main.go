package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"

	"github.com/PuerkitoBio/purell"
)

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

	normalizedOriginString, err := purell.NormalizeURLString(*domain, purell.FlagsSafe)

	if err != nil {
		log.Fatal("Unable to normalize domain.")
	}

	normalizedURL, err := url.Parse(normalizedOriginString)

	if err != nil || !normalizedURL.IsAbs() {
		log.Fatal("domain must be a fully qualified URL.")
	}

	return normalizedURL, *jobs
}

func main() {
	originURL, jobs := checkFlags()

	log.Println("Domain:", originURL.String())
	log.Println("Workers:", jobs)

	done := make(chan bool)

	sitemap := NewSiteMap(jobs, originURL)

	go sitemap.Crawl(jobs, done)

	<-done

	result, _ := json.Marshal(sitemap)

	fmt.Println(string(result))
}
