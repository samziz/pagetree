package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// The Parser does two things for a Crawler: scans its Robots
// file into a slice of banned URLs and scans each page for links.
type Parser struct {

}

// ParseLinks takes an http.Response struct and returns an array of all the links
// in that page (restricted to href links, not JavaScript).
func (*Parser) ParseLinks(r *http.Response) ([]url.URL, error) {
	var links []url.URL

	body, err := safelyReadBody(r)
	if err != nil {
		return links, err
	}

	linkStrings, err := extractAllLinks(body)
	if err != nil {
		return links, err
	}

	for _, str := range linkStrings {
		u, err := makeURL(*r.Request.URL, str)

		if err != nil {
			continue
		}

		err = ValidateURL(u)
		if err != nil {
			continue
		}

		links = append(links, u)
	}

	links = dedupeURLArray(links)
	return links, nil
}

// ParseRobots reads a robots.txt file and writes the results to Crawler's
// Policies list. We pass a pointer to p to avoid corrupting its Visits count.
func (*Parser) ParseRobots(r *http.Response) ([]url.URL, error) {
	var urls []url.URL

	bytes, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		return nil, err
	}

	body := string(bytes)

	for _, l := range strings.Split(body, "\n") {
		exp := regexp.MustCompile(`^(Allow|Disallow):\s?([^\s]*)\s*?$`)

		if matches := exp.FindAllStringSubmatch(l, -1); len(matches) > 0 {

			// Bundle all our failure conditions together
			switch {
			case len(matches) == 0:
			case len(matches[0]) < 2:
			case matches[0][1] == "Disallow":
				continue
			}

			path := matches[0][2]

			u, err := makeURL(*r.Request.URL, path)
			if err != nil {
				continue
			}

			urls = append(urls, u)
		}
	}

	return urls, nil
}

// ValidateURL makes sure that a URL is valid in a stronger sense
// than url.Parse does. It checks that it conforms to the HTTP scheme
// and leads to an HTML page we can crawl rather than e.g. a file.
func ValidateURL(u url.URL) error {
	_, err := url.ParseRequestURI(u.String())
	if err != nil {
		return err
	}

	exp := regexp.MustCompile(`http(s)?://(www.)?.*..*/`)
	if !exp.MatchString(u.String()) {
		return errors.New("Invalid URL or wrong scheme")
	}

	exp2 := regexp.MustCompile(`.jp(e?)g$|.css$|.ico$`)
	if exp2.MatchString(u.String()) {
		return errors.New("Bad document type")
	}

	return nil
}


/*** Functions for internal use ***/

// dedupeURLArray takes an array of URLs and returns
// that array with all duplicates removed.
func dedupeURLArray(urls []url.URL) []url.URL {
	uniqs := make([]url.URL, 0, len(urls))

	for _, u := range urls {
		exists := false

		// Check if we already added u to our new array.
		for _, un := range uniqs {
			if u == un {
				exists = true
			}
		}

		// If we didn't, add it now.
		if !exists {
			uniqs = append(uniqs, u)
		}
	}

	return uniqs
}

// extractAllLinks takes a Reader (strictly a ReadCloser) with HTML
// contents and returns an array of all its links in string form.
func extractAllLinks(body io.ReadCloser) ([]string, error) {
	links := make([]string, 0, 20)

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return links, err
	}

	exp := regexp.MustCompile(`(?:document.location(?:.href)?|href)\s?=\s?(?:"|')([^"]*)(?:"|')`)

	// Strip 'key' values returned by FindAllStringSubmatch
	for _, arr := range exp.FindAllStringSubmatch(string(b), -1) {
		links = append(links, arr[1])
	}

	return links, nil
}

// MakeURL takes a root URL and a relative path (in string form)
// and constructs an absolute URL path.
func makeURL(root url.URL, path string) (url.URL, error) {
	path = strings.TrimSpace(path)

	if strings.HasPrefix(path, "/") {
		root.Path = path
		return root, nil
	}

	url, err := url.Parse(path)
	if err != nil {
		return root, err
	}

	return *url, nil
}

// safelyReadBody reads from an http.Response body and returns it how it was
// found so that other functions, e.g. user callbacks, can read it properly.
func safelyReadBody(r *http.Response) (io.ReadCloser, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	rdr := ioutil.NopCloser(bytes.NewBuffer(buf))
	rdr2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	r.Body = rdr2
	return rdr, nil
}