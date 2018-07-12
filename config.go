package main

import "time"

const (
	// Crawler will use a lot of goroutines by default.
	DefaultMaxRoutines = 10000

	// Crawler will obey robots.txt file by default.
	DefaultRespectRobots = true

	// Crawler will log by default.
	DefaultShouldLog = true

	// DefaultUserAgent is the default user agent string.
	DefaultUserAgent = "Gocrawl"
	DecoyUserAgent = "Mozilla/5.0 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"

	// Default timeout for HTTP requests
	DefaultTimeout = 2 * time.Second
)