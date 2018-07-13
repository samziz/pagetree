package main

import (
	"net/http"
	"net/url"
	"testing"
	)

// We reuse the same Crawler for all tests (or most tests)
// to keep config as simple as possible.
var c Crawler = Crawler{
		Client: http.Client{
			Timeout: DefaultTimeout,
		},
		NewNodes: 		 make(chan *Node),
		RespectRobots:   DefaultRespectRobots,
		Timeout:         DefaultTimeout,
		UserAgent:       DefaultUserAgent,
		MaxPages:		 10,
	}

// Use the same URL (Wikipedia) for all requests since we know
// it's likely to remain up and that it has a robots policy.
var s string = "https://en.wikipedia.org/wiki/Battle_of_Actium"

func TestCrawl(t *testing.T) {
	if testing.Short() {
        t.Skip()
    }

	c.DisguiseCrawler()

	u, err := url.Parse(s)
	
	_, err = c.Crawl(*u)
	if err != nil {
		t.Error(err)
	}
}

func TestLoadRobotsPolicy(t *testing.T) {
	u, _ := url.Parse(s)

	err := c.LoadRobotsPolicy(*u)
	if err != nil {
		t.Error(err)
	}
}

func TestRequest(t *testing.T) {
	u, _ := url.Parse(s)

	_, err := c.request(*u)
	if err != nil {
		t.Error(err)
	}
}