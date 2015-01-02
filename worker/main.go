package main

import (
	"fmt"
	"github.com/apcera/nats"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"github.com/jasdel/harvester/worker/scraper"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
)

func main() {
	queueSend, err := queue.NewPublisher(nats.DefaultURL, "url_queue")
	if err != nil {
		panic(err)
	}
	defer queueSend.Close()

	queueRecv, err := queue.NewReceiver(nats.DefaultURL, "work_queue")
	if err != nil {
		panic(err)
	}
	defer queueRecv.Close()

	sc := storage.NewClient()

	for {
		item := <-queueRecv.Receive()
		doWork(item, sc, queueSend)

		// Don't overload endpoints being crawled
		<-time.After(250 * time.Millisecond)
	}

}

func doWork(item *types.URLQueueItem, sc *storage.Client, sender queue.Publisher) {
	mime, urls, err := scraper.Scrape(item.URL, http.DefaultClient)
	fmt.Println("url", item.URL, "mime:", mime, "url count", len(urls), "err:", err)

	// Update mime type for the URL
	if err := sc.ForURL(item.URL).Update(mime, len(urls) > 0); err != nil {
		log.Println("Worker doWork, failed to add update URL's mime type", item.URL, mime, err)
	}

	for i := 0; i < len(urls); i++ {
		u := urls[i]
		kind := looksLikeImageURL(urls[i])
		if kind != "" {
			// Strip off the image so it isn't queued up
			urls = append(urls[:i], urls[i+1:]...)
		} else {
			kind = storage.DefaultURLMime
		}

		su := sc.ForURL(u)

		// Only add the URL if it is already not known
		if known, _ := su.KnownWithRefer(item.URL); !known {
			// set the descendant as known and from this work item's URL
			if err := su.Add(item.URL, kind); err != nil {
				log.Println("Worker doWork, failed to add image to know URLs", item.URL, u, err)
			}
		}

		if err := su.AddResult(item.Origin, item.URL, kind); err != nil {
			log.Println("Worker doWork, failed to add result", err, item.Origin, item.URL, u)
		}

		if item.Level+1 < 2 {
			q := &types.URLQueueItem{
				Origin: item.Origin,
				Refer:  item.URL,
				URL:    u,
				Level:  item.Level + 1,
			}
			if err := su.AddPending(q.Origin); err != nil {
				log.Println("Worker doWork, failed to add pending url", err)
			}

			sender.Send(q)
		}
	}

}

// Attempts to identify if the URL passed in is a image
// a non-empty string will be returned if the URL looks like
// an image URL.
func looksLikeImageURL(u string) string {
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

	}

	return ""
}
