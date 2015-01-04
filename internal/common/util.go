package common

import (
	"log"
	"net/url"
	"path"
	"strings"
)

// Attempts to identify the content of the URL points to based on
// the URI path's extension.
func GuessURLsMime(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		log.Println("guessURLsMime failed to parse URL", u)
		return ""
	}

	// lower and trim the leading '.' from the extension
	ext := strings.ToLower(path.Ext(parsed.Path))

	switch ext {
	case ".gif":
		return "image/gif"
	case ".jpeg", ".jpg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".css":
		return "text/css"
	case ".js":
		return "text/javascript"
	case "":
		// Guessing that a path without an extension is
		// text/html. This could easily be wrong, but in
		// general it would be true.
		return "text/html"
	default:
		return ""
	}
}

// Returns if the content of the URL based on mime type
// can be ignored and doesn't need to be queued for crawling.
func CanSkipMime(mime string) bool {
	return strings.HasPrefix(mime, "image") ||
		mime == "text/css" ||
		mime == "text/javascript"
}
