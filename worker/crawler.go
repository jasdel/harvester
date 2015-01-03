package main

import (
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

type Crawler struct {
	queuePub queue.Publisher
	sc       *storage.Client
	maxLevel int
}

func NewCrawler(queuePub queue.Publisher, sc *storage.Client, maxLevel int) *Crawler {
	return &Crawler{
		queuePub: queuePub,
		sc:       sc,
		maxLevel: maxLevel,
	}
}

func (c *Crawler) crawl(item *types.URLQueueItem) {
	startedAt := time.Now()
	mime, urls, err := Scrape(item.URL, http.DefaultClient)

	log.Println("Worker crawl: Scape URL", item.URL, "mime:", mime, "level", item.Level, "descendants", len(urls), "duration", time.Now().Sub(startedAt).String(), "error", err)

	// Update mime type for the URL
	if err := c.sc.ForURL(item.URL).Update(mime, true); err != nil {
		log.Println("Worker crawl: failed to add update URL's mime type", item.URL, mime, err)
	}

	// Collect the job ids for this origin so their results can be updated
	jobIdsForOrigin, err := c.sc.ForURL(item.Origin).GetJobIds()
	if err != nil {
		log.Println("Worker crawl: origin has no associated jobId", item.Origin)
		return
	}

	startedAt = time.Now()
	for i := 0; i < len(urls); i++ {
		u := urls[i]

		kind := guessURLsMime(u)

		su := c.sc.ForURL(u)
		for _, id := range jobIdsForOrigin {
			if err := su.AddResult(id, item.Origin, item.URL, kind); err != nil {
				log.Println("Worker crawl: failed to add result", err, item.Origin, item.URL, u)
			}
		}

		if canSkipMime(kind) {
			urls = append(urls[:i], urls[i+1:]...)
			i-- // Step back on to pick up what would of been the item at the next index.
			continue
		}

		if item.Level+1 < c.maxLevel {
			q := &types.URLQueueItem{
				Origin: item.Origin,
				Refer:  item.URL,
				URL:    u,
				Level:  item.Level + 1,
			}
			if err := su.AddPending(q.Origin); err != nil {
				log.Println("Worker crawl: failed to add pending URL", err)
			}

			c.queuePub.Send(q)
		} else if known, _ := su.KnownWithRefer(item.URL); !known {
			// Only add the URL if it is already not known set the descendant as known
			// and from this work item's URL, but it will be marked as not-crawled by by default.
			if err := su.Add(item.URL, kind); err != nil {
				log.Println("Worker doWork, failed to add image to know URLs", item.URL, u, err)
			}
		}
	}

	if err := c.sc.ForURL(item.URL).DeletePending(item.Origin); err != nil {
		log.Println("Worker crawl: failed to delete pending record for", item.URL, item.Origin)
	}

	origSU := c.sc.ForURL(item.Origin)
	if pending, _ := origSU.HasPending(); !pending {
		log.Println("Worker crawl: marking origin as complete", item.Origin)
		if err := origSU.MarkJobURLComplete(); err != nil {
			log.Println("Worker doWork, failed to mark jobs completed for", item.Origin, err)
		}
	}

	log.Println("Finished crawling of", item.URL, "duration", time.Now().Sub(startedAt).String())

}

// Attempts to identify the content of the URL points to based on
// the URI path's extension.
func guessURLsMime(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		log.Println("Worker looksLikeImageURL failed to parse URL", u)
		return ""
	}

	ext := path.Ext(parsed.Path)
	if len(ext) == 0 {
		return ""
	}

	// lower and trim the leading '.' from the extension
	ext = strings.ToLower(ext[1:])

	switch ext {
	case "gif":
		return "image/gif"
	case "jpeg", "jpg":
		return "image/jpeg"
	case "png":
		return "image/png"
	case "css":
		return "text/css"
	case "js":
		return "text/javascript"
	default:
		return storage.DefaultURLMime
	}
}

// Returns if the content of the URL based on mime type
// can be ignored and doesn't need to be queued for crawling.
func canSkipMime(mime string) bool {
	return strings.HasPrefix(mime, "image") ||
		mime == "text/css" ||
		mime == "text/javascript"
}
