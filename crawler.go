package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

/*** Errors ***/

var (
	// Robots.txt file forbids crawling this page
	DisallowedError = errors.New("Error: Disallowed by robots.txt")

	// Server returned response code 404
	NotFoundError = errors.New("Error: Page not found")

	// Server returned response code 500
	ServerError = errors.New("Error: Server could not load page")
)

// A Crawler is the main interface to gocrawl, called from main.go.
type Crawler struct {
	// BannedURLs specifies URLs prohibited by the domain's robots policy.
	BannedURLs []url.URL

	// Client stores the client we should use to make requests. Client can
	// theoretically be replaced with any type that embeds http.Client.
	Client http.Client

	// Host specifies the domain the user wants to crawl.
	Host string

	// Map stores all the pages visited in a tree structure.
	Map Map

	// The Parser (defined in parser.go) does two jobs: parsing the robots file,
	// and parsing for links.
	Parser Parser

	// StartURL specifies the URL or URLs that Crawler should enter from.
	StartURL url.URL

	// Ticker is used to limit the crawl rate if the user specifies this.
	Ticker *time.Ticker

	// NewURLs contains pointers to Nodes to be crawled from. If a Node has no links
	// or all its links are in VisitedURLs, it won't be modified.
	NewNodes chan *Node

	// Timeout specifies the length of time Crawler should wait for a 
	// response from any given page.
	Timeout time.Duration

	// Maximum number of goroutines Crawler is allowed to spin up
	MaxRoutines int

	// RespectRobots specifies whether Crawler should skip links that are
	// disallowed by this site's robots.txt file.
	RespectRobots bool

	// UserAgent is the user agent string with which Crawler identifies itself
	// to the domain.
	UserAgent string

	// VisitedURLs contains all the URLs Crawler has visited.
	VisitedURLs []url.URL
}

// Crawl loads the start URL and robots policy, then spins up
// goroutines to map the site.
func (c *Crawler) Crawl(s string) (m Map, err error) {
	u, err := url.Parse(s)
	if err != nil {
		return
	}

	// Start ticker to rate-limit visits
	if c.RespectRobots {
		err = c.LoadRobotsPolicy(*u)
		if err != nil {
			return
		}
	}

	c.Ticker = time.NewTicker(time.Second)

	kill := make(chan bool)

	for i := 0; i < c.MaxRoutines; i++ {
		go c.consumeChan(kill)
	}

	n := &Node{URL: *u}
	c.NewNodes <- n
	c.Map.Start = n

	select {
	case _ = <-kill:
		m = c.Map
		return
	}
}

// DisguiseCrawler adopts Googlebot user agent (it's too difficult to disguise
// ourselves as a genuine web browser) and sets RespectRobots to false.
func (c *Crawler) DisguiseCrawler() {
	c.RespectRobots = false
	c.UserAgent = DecoyUserAgent
}

func (c *Crawler) LoadRobotsPolicy(u url.URL) error {
	p, err := makeRobotsPath(u)
	if err != nil {
		return err
	}

	r, err := c.request(*p)
	if err != nil {
		return err
	}

	defer r.Body.Close()

	urls, err := c.Parser.ParseRobots(r)
	if err != nil {
		return err
	}

	c.BannedURLs = urls
	return nil
}

/*** Internal funcs ***/

func (c *Crawler) addNodePtrsToChannel(n *Node) {
	// TODO: Accept variadic args
	c.NewNodes <- n
}

// Read from NewNodes channel until it's empty, writing any
// links we find back to this channel.
func (c *Crawler) consumeChan(kill chan bool) {
	for {
		select {
		case n, ok := <-c.NewNodes:
			if !ok {
				break
			}

			if c.checkSliceForURL(n.URL, c.VisitedURLs) {
				break
			}

			if c.checkSliceForURL(n.URL, c.BannedURLs) {
				break
			}

			// Make request to server and check response is not nil
			res, err := c.request(n.URL)
			if err != nil {
				fmt.Println(err)
				break
			}

			urls, err := c.Parser.ParseLinks(res)
			nodes := CreateNodesFromURLs(urls)

			// Write nodes to channel to be processed
			for _, n := range nodes {
				c.NewNodes <- &n
				n.Children = append(n.Children, &n)
			}
		case <-time.After(10*time.Second):
			// Send signal for c.Crawl to return
			kill <- true
			return
		}
	}
}

// checkSliceForURL checks a slice for a URL.
func (c *Crawler) checkSliceForURL(u url.URL, s []url.URL) bool {
	for _, o := range s {
		if o == u {
			return true
		}
	}

	return false
}

// An interface to c.Client.Do, applying all the settings (user agent,
// settings, etc.) specified in c.CrawlerConfig.
func (c *Crawler) request(url url.URL) (*http.Response, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)

	fmt.Println("made req")
	res, err := c.Client.Do(req)
	fmt.Println("done req")
	if err != nil {
		return nil, err
	}

	return res, nil
}

func makeRobotsPath(u url.URL) (*url.URL, error) {
	var buf bytes.Buffer
	var path = []string{u.Scheme, "://", u.Host, "/", "robots.txt"}

	for _, s := range path {
		buf.WriteString(s)
	}

	return url.Parse(buf.String())
}

func (c *Crawler) makeRootURL(u url.URL) (*url.URL, error) {
	var buf bytes.Buffer
	var path = []string{u.Scheme, "://", u.Host, "/"}

	for _, s := range path {
		buf.WriteString(s)
	}

	return url.Parse(buf.String())
}
