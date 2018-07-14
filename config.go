package main

const (
	// The default number of milliseconds to wait between one page crawl
	// and another. By default this is 1ms.
	DefaultCrawlRate = 1

	// Crawler has a DefaultMaxPages of -1, which is interpreted
	// as *no limit to the number of pages*.
	DefaultMaxPages = -1

	// Crawler will use a lot of goroutines by default.
	DefaultMaxRoutines = 10000

	// Crawler will obey robots.txt file by default.
	DefaultRespectRobots = true

	// DefaultUserAgent is the default user agent string.
	DefaultUserAgent = "Gocrawl"
	DecoyUserAgent   = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

	// Default timeout for HTTP requests (ms)
	DefaultHTTPTimeout = 2000

	// Default timeout waiting for new nodes to crawl (ms)
	DefaultCrawlerTimeout = 5000
)
