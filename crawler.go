package main

import (
	"bytes"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
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

	// Maximum number of pages Crawler should visit.
	MaxPages int

	// RespectRobots specifies whether Crawler should skip links that are
	// disallowed by this site's robots.txt file.
	RespectRobots bool

	// UserAgent is the user agent string with which Crawler identifies itself
	// to the domain.
	UserAgent string

	// VisitedURLs contains all the URLs Crawler has visited.
	VisitedURLs []url.URL
	VisitedURLsMutex sync.Mutex
}

// Crawl loads the start URL and robots policy, then spins up
// goroutines to map the site.
func (c *Crawler) Crawl(u url.URL) (m Map, err error) {
	c.Host = u.Host

	// Start ticker to rate-limit visits
	if c.RespectRobots {
		err = c.LoadRobotsPolicy(u)
		if err != nil {
			return
		}
	}

	c.Ticker = time.NewTicker(time.Second)

	kill := make(chan bool)

	for i := 0; i < c.MaxRoutines; i++ {
		go c.worker(kill)
	}

	n := &Node{URL: u}
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

var counter int
// Read from NewNodes channel until it's empty, writing any
// links we find back to this channel.
func (c *Crawler) worker(kill chan bool) {
	for {
	Start:
		select {
		case n, ok := <-c.NewNodes:
			// Check for errors
			switch {

			case !ok,
				!compareHosts(n.URL.Host, c.Host),
				checkSliceForURL(n.URL, c.BannedURLs),
				checkSliceForURL(n.URL, c.VisitedURLs):
				goto Start

			case c.MaxPages > -1 && len(c.VisitedURLs) > c.MaxPages:
				kill <- true
				return
			}

			// Make request to server and check response is not nil
			res, err := c.request(n.URL)
			if err != nil {
				continue
			}

			urls, err := c.Parser.ParseLinks(res)
			urls = c.excludeExternalLinks(urls)

			nodes := CreateNodesFromURLs(urls)

			// Write nodes to channel to be processed
			var arr []*Node

			for _, n := range nodes {
				node := n
				c.NewNodes <- &node
				arr = append(arr, &node)
			}

			n.Children = arr

			// Append to VisitedURLs
			c.writeToVisitedURLs(n.URL)
		case <-time.After(10*time.Second):
			// Send signal for c.Crawl to return
			kill <- true
			return
		}
	}
}

// checkSliceForURL checks a slice for a URL.
func checkSliceForURL(u url.URL, s []url.URL) bool {
	for _, o := range s {
		if o == u {
			return true
		}
	}

	return false
}

func compareHosts(h1 string, h2 string) bool {
	if h1 == h2 {
		return true
	}

	if strings.Contains(h1, h2) {
		return true
	}

	if strings.Contains(h2, h1) {
		return true
	}

	return false
}

func (c *Crawler) excludeExternalLinks(urls []url.URL) []url.URL {
	internalURLs := make([]url.URL, 0, len(urls))

	for _, u := range urls {
		if compareHosts(u.Host, c.Host) {
			internalURLs = append(internalURLs, u)
		}
	}

	return internalURLs
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

// An interface to c.Client.Do, applying all the settings (user agent,
// settings, etc.) specified in c.CrawlerConfig.
func (c *Crawler) request(url url.URL) (*http.Response, error) {
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", c.UserAgent)

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// writeToVisitedURLs is a safe interface to write a value to Crawler's
// VisitedURLs array without race conditions causing double writes.
func (c *Crawler) writeToVisitedURLs(u url.URL) {
	c.VisitedURLsMutex.Lock()
	c.VisitedURLs = append(c.VisitedURLs, u)
	c.VisitedURLsMutex.Unlock()
}
