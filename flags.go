package main

import (
	"errors"
	"flag"
	"net/url"
	"time"
)

// ParseFlags parses flags passed to the program, using these values
// to modify a pointer to a Crawler.
func ApplyFlags(c *Crawler) {
	// Load flags
	crawlRate := flag.Int("crawlrate", DefaultCrawlRate, "maximum number of requests per millisecond")
	crawlerTimeout := flag.Int("timeout", DefaultCrawlerTimeout, "maximum length of time to wait for new nodes (ms)")
	httpTimeout := flag.Int("httptimeout", DefaultHTTPTimeout, "maximum length of time to wait for any HTTP response (ms)")
	respectRobots := flag.Bool("respectrobots", DefaultRespectRobots, "respect robots.txt policy")
	maxRoutines := flag.Int("maxroutines", DefaultMaxRoutines, "maximum number of goroutines to spin up")
	maxPages := flag.Int("maxpages", DefaultMaxPages, "port for remote (default of 22 should usually work)")
	flag.Parse()

	c.Client.Timeout = time.Duration(*httpTimeout) * time.Millisecond
	c.CrawlRate = time.Duration(*crawlRate) * time.Millisecond
	c.Timeout = time.Duration(*crawlerTimeout) * time.Millisecond
	c.MaxRoutines = *maxRoutines
	c.MaxPages = *maxPages
	c.RespectRobots = *respectRobots
}

// GetURLCommand gets the first argument passed and checks that
// it's a valid URL that Crawler is going to be happy with.
func GetURLCommand() (url.URL, error) {
	args := flag.Args()

	if len(args) == 0 {
		return url.URL{}, errors.New("No command passed")
	}

	u, err := url.Parse(args[0])
	if err != nil {
		// Return an zero-valued struct as we don't know if 
		// it's safe to dereference u
		return url.URL{}, err
	}

	err = ValidateURL(*u)
	if err != nil {
		return *u, errors.New("The URL you entered is not valid")
	}

	return *u, nil
}
