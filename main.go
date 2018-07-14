package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	c := Crawler{
		Client:        http.Client{},
		NewNodes:      make(chan *Node),
		UserAgent:     DefaultUserAgent,
	}

	// Disguise Crawler by default, or we won't be able to
	// crawl most websites.
	c.DisguiseCrawler()

	// Apply any command-line flags (or otherwise the defaults
	// we set in config.go).
	ApplyFlags(&c)

	// Get URL from command line and parse it.
	u, err := GetURLCommand()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("\nAbout to crawl %s. Hold tight.\n\n", u.String())

	// Crawl all internal links.
	m, err := c.Crawl(u)
	if err != nil {
		log.Fatal(err)
	}

	// Print the map we've created of this website.
	PrintMap(m)
}
