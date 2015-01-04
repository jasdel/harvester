package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Requests, and scrapes the content of a URL. The URL's content will only be scrapped
// if its returned Content-Type (mime) is text/html. The list of URLs will also be
// de-duped preventing duplicate entries.
func Scrape(tgtURL string, client *http.Client) (mime string, urls []string, err error) {
	var body []byte
	mime, body, err = requestContent(client, tgtURL)
	if err != nil {
		return "", nil, err
	}

	if body == nil || mime != "text/html" {
		// Only valid body responses, or HTML documents are scrapped
		return mime, []string{}, nil
	}

	tgtURLParsed, _ := url.Parse(tgtURL)
	foundUrls := findHTMLDocURLs(body)

	urlMap := make(map[string]struct{})
	urls = []string{}
	for _, u := range foundUrls {
		if u, err := normalizeURL(tgtURLParsed, u); err != nil {
			// Drop URL if it is unable to be normalized, because it means
			// they are not valid URLs
			continue
		} else if _, ok := urlMap[u]; !ok {
			// Prevent duplicate entries
			urls = append(urls, u)
		}
	}

	return mime, urls, nil
}

// Requests content from a URL and returns the properties of that content along with its body.
// a body will only be returned if the content type of the response is a text/*
func requestContent(client *http.Client, tgtURL string) (mime string, body []byte, err error) {
	var resp *http.Response
	resp, err = client.Get(tgtURL)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	return validateContent(resp)
}

// Validates the content of the response to determine if it is text, and can be
// parsed
func validateContent(resp *http.Response) (mime string, body []byte, err error) {
	mime = resp.Header.Get("Content-Type")
	if mime == "" {
		mime = "application/octet-stream"
	}
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}

	if !strings.HasPrefix(mime, "text") {
		// If this is not a text document there is no point reading the body
		return mime, nil, nil
	}

	buf := bytes.Buffer{}
	// TODO this should be limited to a sane max length
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return "", nil, err
	}

	return mime, buf.Bytes(), nil
}

// Inspects the URL provided and normalizes it so that it contains
// a scheme and host.  If the scheme or host are missing from the
// URL the origin's values will be substituted in their place.
// If the URL has a relative path the path of the origin will be pre-pended.
func normalizeURL(origin *url.URL, u string) (string, error) {
	if strings.HasPrefix(u, "data:") {
		return "", fmt.Errorf("not URL link")
	}

	normURL, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	if normURL.Host == "" {
		normURL.Host = origin.Host
	}
	if normURL.Scheme == "" {
		normURL.Scheme = origin.Scheme
	}
	if !strings.HasPrefix(normURL.Path, "/") {
		// Need to store if the path has a trailing slash, because it will be stripped
		// off via path.Join
		hasTraillingSlash := strings.HasSuffix(normURL.Path, "/")
		normURL.Path = path.Join(origin.Path, normURL.Path)
		if hasTraillingSlash {
			normURL.Path += "/"
		}
	}

	return normURL.String(), nil
}
