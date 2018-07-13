package main

import (
	"net/url"
	"testing"
	)

// We reuse the same Parser for all tests. We don't need to do any
// configuration: Parser doesn't contain any fields, only methods.
var p Parser

func TestParseLinks(t *testing.T) {
	u, _ := url.Parse(s)

	res, err := c.request(*u)
	if err != nil {
		t.Error("Error with unrelated func crawler.request:", err)
	}

	links, err := p.ParseLinks(res)
	
	if err != nil {
		t.Error(err)
	}

	if len(links) == 0 {
		// Anyone who's played the Wikipedia game knows that
		// this page should contain some links!
		t.Error(err)
	}
}

func TestParseRobots(t *testing.T) {
	u, _ := url.Parse(s)

	r, err := makeRobotsPath(*u)
	if err != nil {
		t.Error("Error with unrelated func crawler.makeRobotsPath:", err)
	}

	res, err := c.request(*r)
	if err != nil {
		t.Error("Error with unrelated func crawler.LoadRobotsPolicy:", err)
	}

	links, err := p.ParseRobots(res)
	if err != nil {
		t.Error(err)
	}

	if len(links) == 0 {
		// We don't know for sure that our robots policy is always
		// going to disallow any paths
		t.Error("WARNING:", err)
	}
}