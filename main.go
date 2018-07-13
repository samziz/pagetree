package main

import (
	"log"
	"net/http"
)

func main() {
	c := Crawler{
		Client: http.Client{
			Timeout: DefaultTimeout,
		},
		NewNodes: 		 make(chan *Node),
		RespectRobots:   DefaultRespectRobots,
		Timeout:         DefaultTimeout,
		UserAgent:       DefaultUserAgent,
	}

	// Disguise Crawler by default, or we won't be able to 
	// crawl most websites.
	c.DisguiseCrawler()

	// Apply settings specified by user.
	ApplyFlags(&c)

	// Get URL from command line
	u, err := GetURLCommand()
	if err != nil {
		log.Fatal(err)
	}

	m, err := c.Crawl(u)
	if err != nil {
		log.Fatal(err)
	}

	PrintMap(m)
}