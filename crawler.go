package main

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// A Crawler is the main interface to gocrawl, called from main.go.
type Crawler struct {
	// BannedURLs specifies URLs prohibited by the domain's robots policy.
	BannedURLs []url.URL

	// Client stores the client we should use to make requests. Client can
	// theoretically be replaced with any type that embeds http.Client.
	Client http.Client

	// CrawlRate sets the number of crawls per millisecond.
	CrawlRate time.Duration

	// Host sets the domain the user wants to crawl.
	Host string

	// Map stores all the pages visited in a tree structure - or technically
	// it stores the top node. See map.go for implementation details.
	Map Map

	// The Parser (defined in parser.go) does two jobs: parsing the robots file,
	// and parsing for links.
	Parser Parser

	// RateTicker is used to limit the crawl rate, by default to 1 req/ms.
	RateTicker *time.Ticker

	// StartURL specifies the URL or URLs that Crawler should enter from.
	StartURL url.URL

	// NewURLs contains pointers to Nodes in c.Map to be crawled from. If a Node has
	// no links or all its links are in VisitedURLs, it won't be modified.
	NewNodes chan *Node

	// Timeout specifies the length of time Crawler should wait for a
	// response from any given page.
	Timeout time.Duration

	// MaxRoutines stores the number of goroutines Crawler is allowed to spin up.
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
	VisitedURLsMutex sync.Mutex
	VisitedURLs      []url.URL
}

// Crawl loads the start URL and robots policy, then launches
// a lot of goroutines to map the site. 
func (c *Crawler) Crawl(u url.URL) (m Map, err error) {
	// Store the hostname of the URL so we don't crawl external links.
	c.Host = u.Host

	// If the user has chosen to respect the site's robots policy,
	// load robots.txt and store disallowed paths in c.BannedURLs. 
	if c.RespectRobots {
		err = c.LoadRobotsPolicy(u)
		if err != nil {
			return
		}
	}

	// Initialise our Ticker to limit the crawl rate (we need
	// this for concurrent crawling).
	c.RateTicker = time.NewTicker(c.CrawlRate)

	// Create a channel to receive kill signals from workers.
	kill := make(chan bool)

	// Spin up worker goroutines to get pointers to new Nodes
	// from c.NewNodes, crawl them for links, then add these child
	// nodes to the Map (and send pointers to NewNodes).
	for i := 0; i < c.MaxRoutines; i++ {
		go c.worker(kill)
	}

	// Create the first Node in our Map and send
	// to be processed by the workers we just launched.
	c.Map.Start = &Node{URL: u}
	go c.addNodes(c.Map.Start)

	select {
	// Wait to receive a kill signal from worker goroutines
	// indicating the channel has not been written to for the
	// period specified in c.Timeout (default is 5sec). 
	case _ = <-kill:
		// Return a copy of c.Map - fine since it only contains one pointer.
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

// LoadRobotsPolicy works out the path to robots.txt from the supplied
// URL, then parses the page and stores any disallowed paths in c.BannedURLs.
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

// addNodes adds one or several Nodes to Crawler's NewNodes channel. This 
// is wrapped in a function so it can easily be called as a goroutine.
func (c *Crawler) addNodes(nodes ...*Node) {
	for _, n := range nodes {
		c.NewNodes <- n
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

// compareHosts checks if two URLs are equal. This is a surprisingly
// complicated task 
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

// crawlNode crawls from a particular Node in Crawler's Map and
// creates new Nodes from the results. Designed to be called by workers.
func (c *Crawler) crawlNode(n *Node, kill *chan bool) {
	if !compareHosts(n.URL.Host, c.Host) {
		return
	}

	if checkSliceForURL(n.URL, c.BannedURLs) {
		return
	}

	if checkSliceForURL(n.URL, c.VisitedURLs) {
		return
	}

	if c.MaxPages > -1 {
		if len(c.VisitedURLs) > c.MaxPages {
			*kill <- true
			return
		}
	}

	// Request the page and check we didn't fail (nb this only checks for
	// a failure to connect to the server, not for a bad response).
	res, err := c.request(n.URL)
	if err != nil {
		return
	}

	// Get all of the links out of the HTML document we just fetched.
	urls, err := c.Parser.ParseLinks(res)
	if err != nil {
		return
	}

	// Go through the array and remove any links to external pages.
	urls = c.excludeExternalLinks(urls)

	nodes := CreateNodesFromURLs(urls)

	// Write nodes to channel to be processed
	for _, nn := range nodes {
		node := nn
		c.NewNodes <- &node
		n.Children = append(n.Children, &node)
	}

	c.writeToVisitedURLs(n.URL)
}

// excludeExternalLinks to a slice of URLs and returns
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

// Read from NewNodes channel until it's empty, writing any
// links we find back to this channel.
func (c *Crawler) worker(kill chan bool) {
	for {
		// Wrap our worker in a RateTicker to limit the number of crawls
		// to whatever the user specifies.
		select {
		case <-c.RateTicker.C:
			select {
			case n, ok := <-c.NewNodes:

				// Kill workers and stop crawling if channel is 
				// closed (this should never happen).
				if !ok {
					kill <- true
					return
				}

				c.crawlNode(n, &kill)

			case <-time.After(c.Timeout):
				// Send signal for c.Crawl to return if reading from
				// channel blocks for a certain length of time.
				kill <- true
				return
			}
		}

	}
}

// writeToVisitedURLs is a safe interface to write a value to Crawler's
// VisitedURLs array. The append method is easiest but not thread-safe
// since it creates temporary variables.
func (c *Crawler) writeToVisitedURLs(u url.URL) {
	c.VisitedURLsMutex.Lock()
	defer c.VisitedURLsMutex.Unlock()

	c.VisitedURLs = append(c.VisitedURLs, u)
}
