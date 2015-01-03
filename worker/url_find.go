package main

import (
	"regexp"
	"strings"
)

const (
	// Generic URL pattern for <scheme>://domain/path
	htmlURLRegexp    = `href="([^"]+)|src="([^"]+)`
	cssURLRegexp     = `url\(['"](.+?)['"]\)`
	genericURLRegexp = `(https?:\/\/[\w.\/=&?:-]+)|(\/\/[\w.\/=&?:-]+)`
)

var htmlURLRegexpComp *regexp.Regexp
var cssURLRegexpComp *regexp.Regexp
var genericURLregexpComp *regexp.Regexp

func init() {
	htmlURLRegexpComp = regexp.MustCompile(htmlURLRegexp)
	cssURLRegexpComp = regexp.MustCompile(cssURLRegexp)
	genericURLregexpComp = regexp.MustCompile(genericURLRegexp)
}

// Searches through the HTML document for for strings which look or are used like URLs
func findHTMLDocURLs(doc []byte) []string {
	return findURLs(doc, htmlURLRegexpComp)
}

// Searches through a CSS document for strings which look or are used as URLs
func findCSSDocURLs(doc []byte) []string {
	return findURLs(doc, cssURLRegexpComp)
}

// Searches through a generic document for things which look like URLs
func findGenericDocURLs(doc []byte) []string {
	return findURLs(doc, genericURLregexpComp)
}

// Searches through the document searching for matches, and returns those
func findURLs(doc []byte, reg *regexp.Regexp) []string {
	urls := []string{}

	matches := reg.FindAllSubmatch(doc, -1)
	for i := 0; i < len(matches); i++ {
		matchGroup := matches[i]
		for j := 1; j < len(matchGroup); j++ {
			// Skip the first index since it is the full matched phrase, not the sub match
			if len(matchGroup[j]) > 0 {
				urls = append(urls, strings.TrimSpace(string(matchGroup[j])))
			}
		}
	}

	return urls
}
