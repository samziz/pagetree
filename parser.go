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

var (
	// HTML body cannot be parsed at all
	InvalidHTMLError = errors.New("Error: Could not parse as HTML document")

	// Link is not valid error
	InvalidLinkError = errors.New("Error: Link is invalid")

	// Cannot parse as robots.txt document
	InvalidRobotsError = errors.New("Error: Could not parse as robots.txt document")
)

// The Parser does two things for a Crawler: scans its Robots
// file into a slice of banned URLs and scans each page for links. 
type Parser struct {
	
}


/*** Exported Parser functions ***/

// ParseLinks takes an http.Response struct and returns an array of all the links
// in that page (restricted to href links, not JavaScript).
func (p *Parser) ParseLinks(r *http.Response) ([]url.URL, error) {
	var links []url.URL

	body, err := safelyReadBody(r)
	if err != nil {
		return links, err
	}

	linkStrings := extractAllLinks(body)

	for _, str := range linkStrings {
		u, err := makeURL(*r.Request.URL, str)

		if err != nil {
			continue
		}

		err = validateLink(u)
		if err != nil {
			continue
		}

		links = append(links, u)
	}


	return links, nil
}

// ParseRobots reads a robots.txt file and writes the results to Crawler's
// Policies list. We pass a pointer to p to avoid corrupting its Visits count.
func (p *Parser) ParseRobots(r *http.Response) ([]url.URL, error) {
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

	urls = dedupeURLArray(urls)
	return urls, nil
}



/*** Unexported functions ***/

func extractAllLinks(body io.ReadCloser) (links []string) {
	b, _ := ioutil.ReadAll(body)
	exp := regexp.MustCompile(`(?:document.location(?:.href)?|href)\s?=\s?(?:"|')([^"]*)(?:"|')`)

	// Strip 'key' values returned by FindAllStringSubmatch
	for _, arr := range exp.FindAllStringSubmatch(string(b), -1) {
		links = append(links, arr[1])
	}

	return
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

func dedupeURLArray(urls []url.URL) []url.URL {
	uniqs := make([]url.URL, 0, len(urls))

	for _, u := range urls {
		for _, un := range uniqs {
			if u == un {
				continue
			}
		}

		uniqs = append(uniqs, u)
	}

	return uniqs
}

// safelyReadBody reads from an http.Response body and returns it how it was
// found so that other functions, e.g. user callbacks, can read it properly.
func safelyReadBody(r *http.Response) (io.ReadCloser, error) {
	buf, _ := ioutil.ReadAll(r.Body)
	rd := ioutil.NopCloser(bytes.NewBuffer(buf))
	rd2 := ioutil.NopCloser(bytes.NewBuffer(buf))

	r.Body = rd2
	return rd, nil
}

func validateLink(u url.URL) error {
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
