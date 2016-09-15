package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
)

func TestSiteMarshalJSON(t *testing.T) {
	siteURL, _ := url.Parse("http://the-site-address")
	linkURL, _ := url.Parse("http://link-address")
	nofollowLinkURL, _ := url.Parse("http://nofollow-link-address")
	scriptURL, _ := url.Parse("http://script-address")
	imageURL, _ := url.Parse("http://image-address")
	videosURL, _ := url.Parse("http://video-address")
	audioURL, _ := url.Parse("http://audio-address")
	cssURL, _ := url.Parse("http://css-address")

	site := Site{
		URL:           *siteURL,
		Links:         []url.URL{*linkURL},
		NoFollowLinks: []url.URL{*nofollowLinkURL},
		Scripts:       []url.URL{*scriptURL},
		Images:        []url.URL{*imageURL},
		Videos:        []url.URL{*videosURL},
		Audio:         []url.URL{*audioURL},
		CSS:           []url.URL{*cssURL},
	}

	serialized, _ := site.MarshalJSON()

	type unmarshalled struct {
		URL     string   `json:"url"`
		Links   []string `json:"links"`
		Scripts []string `json:"scripts"`
		Images  []string `json:"images"`
		Videos  []string `json:"videos"`
		Audio   []string `json:"audio"`
		CSS     []string `json:"css"`
	}

	var result unmarshalled

	err := json.Unmarshal(serialized, &result)

	if err != nil {
		t.Error(err)
	}

	expected := unmarshalled{
		URL:     "http://the-site-address",
		Links:   []string{"http://link-address", "http://nofollow-link-address"},
		Scripts: []string{"http://script-address"},
		Images:  []string{"http://image-address"},
		Videos:  []string{"http://video-address"},
		Audio:   []string{"http://audio-address"},
		CSS:     []string{"http://css-address"},
	}

	// TODO: Make this less brittle.
	if !reflect.DeepEqual(result, expected) {
		t.Error("Expected", result, "to be", serialized)
	}
}

func TestSiteCrawl(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./testPage.html")
	}))
	defer ts.Close()

	siteURL, _ := url.Parse(ts.URL)

	site := Site{URL: *siteURL}

	site.Crawl()

	expectedLinksURL, _ := url.Parse("http://some-link-address.html")

	if !reflect.DeepEqual(site.Links, []url.URL{*expectedLinksURL}) {
		t.Error("expected", site.Links, "to be", []url.URL{*expectedLinksURL})
	}

	expectedScriptsURL, _ := url.Parse("http://some-script-address.js")

	if !reflect.DeepEqual(site.Scripts, []url.URL{*expectedScriptsURL}) {
		t.Error("expected", site.Scripts, "to be", []url.URL{*expectedScriptsURL})
	}

	expectedImagesURL, _ := url.Parse("http://some-image-address.png")

	if !reflect.DeepEqual(site.Images, []url.URL{*expectedImagesURL}) {
		t.Error("expected", site.Images, "to be", []url.URL{*expectedImagesURL})
	}

	expectedVideosURLOne, _ := url.Parse("http://some-video-source.mp4")
	expectedVideosURLTwo, _ := url.Parse("http://some-other-video-source.mp4")

	if !reflect.DeepEqual(site.Videos, []url.URL{*expectedVideosURLOne, *expectedVideosURLTwo}) {
		t.Error("expected", site.Videos, "to be", []url.URL{*expectedVideosURLOne, *expectedVideosURLTwo})
	}

	expectedAudioURLOne, _ := url.Parse("http://some-audio-source.mp4")
	expectedAudioURLTwo, _ := url.Parse("http://some-other-audio-source.mp4")

	if !reflect.DeepEqual(site.Audio, []url.URL{*expectedAudioURLOne, *expectedAudioURLTwo}) {
		t.Error("expected", site.Audio, "to be", []url.URL{*expectedAudioURLOne, *expectedAudioURLTwo})
	}

	expectedCSSURL, _ := url.Parse("http://some-css-address.css")

	if !reflect.DeepEqual(site.CSS, []url.URL{*expectedCSSURL}) {
		t.Error("expected", site.CSS, "to be", []url.URL{*expectedCSSURL})
	}
}
