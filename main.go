package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	c := &Crawler{
		Client: http.Client{
			Timeout: DefaultTimeout,
		},
		MaxRoutines:   	 DefaultMaxRoutines,
		NewNodes: 		 make(chan *Node),
		RespectRobots:   DefaultRespectRobots,
		Timeout:         DefaultTimeout,
		UserAgent:       DefaultUserAgent,
	}

	c.DisguiseCrawler()
	m, err := c.Crawl("https://stackoverflow.com/questions/46836534/golang-native-http-client-hangs-on-particular-uri")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(m)
}