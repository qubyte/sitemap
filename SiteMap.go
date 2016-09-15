package sitemap

import (
	"net/url"
	"sync"
)

// SiteMap is a container for a set of URLs for assets used by the site, and a
// Mutex to allow safe access over go routines.
type SiteMap struct {
	Sites  map[string]*Site
	mutex  sync.Mutex
	origin *url.URL
}

// NewSiteMap is a constructor function used to return a SiteMap instance with
// an initialized sites map.
func NewSiteMap(origin *url.URL) *SiteMap {
	return &SiteMap{
		Sites:  make(map[string]*Site),
		origin: origin,
	}
}

// worker accepts URLs through a channel, makes a Site instance using it, and
// adds the site to the sitemap. The Site instance is then instructed to crawl,
// and links it discovers are pushed into another channel. Decrements the
// WaitGroup when a site has been crawled.
func (m *SiteMap) worker(id int, wg *sync.WaitGroup, queue chan *Site) {
	for s := range queue {
		s.Crawl()

		// Find new links within this domain for the crawled site, and add them
		// to the queue.
		for _, u := range s.Links {
			if !isSameDomain(&u, m.origin) {
				continue
			}

			site := Site{URL: u}

			wasSet := m.setOnce(&site)

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

// Crawl begins the crawl. It hides the use of a WaitGroup behind a
// synchronization channel.
func (m *SiteMap) Crawl(workers int, done chan<- bool) {
	// Sites to crawl will be pushed into this stream.
	queue := make(chan *Site, workers)

	// The number of pages to crawl is initially unknown, so a WaitGroup is used
	// to block the completion of Crawl main until the WaitGroup is empty.
	var wg sync.WaitGroup

	// Create a worker pool to make requests and populate sites in the map.
	for w := 0; w < workers; w++ {
		go m.worker(w, &wg, queue)
	}

	// Since we don't know how large a sitemap will be, we use a WaitGroup. Each
	// new site to crawl will increment it, and once crawled will decrement it.
	// Start the site crawl by incrementing and pushing the given domain into
	// the sites channel.
	wg.Add(1)

	// Create a site to represent the origin and push it into the queue.
	site := Site{URL: *m.origin}
	m.setOnce(&site)
	queue <- &site

	// Wait until there are no sites left to crawl.
	wg.Wait()

	close(queue)

	done <- true
}

// SetOnce locks the SiteMap mutex and checks if a Site is already represented
// in the SiteMap. If it is not, the Site is set to the SiteMap and SetOnce
// returns true. Otherwise it does not set the Site to the SiteMap and returns
// false. In either case the Mutex is unlocked when the SetOnce returns.
func (m *SiteMap) setOnce(s *Site) bool {
	link := s.URL.String()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, has := m.Sites[link]; has {
		return false
	}

	m.Sites[link] = s

	return true
}
