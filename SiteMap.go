package main

import "sync"

// SiteMap is a container for a set of URLs for assets used by the site, and a
// Mutex to allow safe access over go routines.
type SiteMap struct {
	Sites map[string]*Site `json:sites`
	mutex sync.Mutex
}

// NewSiteMap is a constructor function used to return a SiteMap instance with
// an initialized sites map.
func NewSiteMap() *SiteMap {
	return &SiteMap{Sites: make(map[string]*Site)}
}

// SetOnce locks the SiteMap mutex and checks if a Site is already represented
// in the SiteMap. If it is not, the Site is set to the SiteMap and SetOnce
// returns true. Otherwise it does not set the Site to the SiteMap and returns
// false. In either case the Mutex is unlocked when the SetOnce returns.
func (m *SiteMap) SetOnce(s *Site) bool {
	link := s.URL.String()

	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, has := m.Sites[link]; has {
		return false
	}

	m.Sites[link] = s

	return true
}
