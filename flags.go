package main

import (
	"errors"
	"flag"
	"net/url"
)

// ParseFlags parses flags passed to the program, using these values
// to modify a pointer to a Crawler.
func ApplyFlags(c *Crawler) {
	// Load flags
	maxRoutines := flag.Int("maxroutines", DefaultMaxRoutines, "maximum number of goroutines to spin up")
	maxPages := flag.Int("maxpages", DefaultMaxPages, "port for remote (default of 22 should usually work)")
	flag.Parse()

	c.MaxRoutines = *maxRoutines
	c.MaxPages = *maxPages
}

func GetURLCommand() (url.URL, error) {
	args := flag.Args()
	
	if len(args) == 0 {
		return url.URL{}, errors.New("No command passed")
	}
	
	u, err := url.Parse(args[0])
	if err != nil {
		return url.URL{}, err
	}

	return *u, nil
}