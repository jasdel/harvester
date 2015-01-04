package main

import (
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"github.com/jasdel/harvester/internal/util"
	"log"
	"net/http"
	"time"
)

type Crawler struct {
	urlQueuePub queue.Publisher
	sc          *storage.Client
	maxLevel    int
}

func NewCrawler(urlQueuePub queue.Publisher, sc *storage.Client, maxLevel int) *Crawler {
	return &Crawler{
		urlQueuePub: urlQueuePub,
		sc:          sc,
		maxLevel:    maxLevel,
	}
}

func (c *Crawler) Crawl(item *types.URLQueueItem) {
	startedAt := time.Now()
	urlClient := c.sc.URLClient()
	defer func() {
		// Make sure the Job is cleaned up even in if an error happens.
		if err := urlClient.DeletePending(item.URL, item.Origin); err != nil {
			log.Println("crawl: Failed to delete pending record for", item.URL, item.Origin)
		}

		// If there are no more pending entries for this origin, all jobs which contain that
		// origin which are not already complete can be marked as complete.
		if complete, err := urlClient.UpdateJobURLIfComplete(item.Origin); err != nil {
			log.Println("crawl: Failed to update if Job URL is complete", item.Origin, err)
		} else if complete {
			log.Println("crawl: Marked Job URL as complete", item.Origin)
		}

		log.Println("crawl: Finished crawling of", item.URL, item.Level, "duration", time.Now().Sub(startedAt).String())
	}()

	mime, urls, err := Scrape(item.URL, http.DefaultClient)
	if err != nil {
		log.Println("crawl: Failed to request and scrape", item.URL, err)
		return
	}

	log.Println("crawl: Request and Scape complete URL", item.URL, "mime:", mime, "level", item.Level, "descendants", len(urls), "duration", time.Now().Sub(startedAt).String(), "error", err)

	// Update mime type for the URL
	if err := urlClient.Update(item.URL, mime, true); err != nil {
		log.Println("crawl: failed to add update URL's mime type", item.URL, mime, err)
		return
	}

	// Collect the job ids for this origin so their results can be updated
	jobIdsForOrigin, err := urlClient.GetJobIdsForURL(item.Origin)
	if err != nil {
		log.Println("crawl: origin has no associated jobId", item.Origin)
		return
	}

	// Only add items to the result if they are greater than the first layer
	// because the first layer is the URLs that are used to start a job,
	// so they do not make sense to be inserted into the results without a refer.
	if item.Level > 0 {
		for _, id := range jobIdsForOrigin {
			if err := urlClient.AddResult(id, item.URL, item.Refer, mime); err != nil {
				log.Println("crawl: failed to add result", err, item.Origin, item.Refer, item.URL)
			}
		}
	}

	for i := 0; i < len(urls); i++ {
		u := urls[i]

		kind := util.GuessURLsMime(u)
		if url, _ := urlClient.GetURLWithRefer(u, item.URL); url == nil {
			// Only add the URL if it is already not known set the descendant as known
			// and from this work item's URL, but it will be marked as not-crawled by by default.
			if err := urlClient.Add(u, item.URL, kind); err != nil {
				log.Println("crawl: failed to add image to know URLs", item.URL, u, err)
			}
		}
		// For any URL that will not be enqueued, add it as a result instead
		for _, id := range jobIdsForOrigin {
			if err := urlClient.AddResult(id, u, item.URL, kind); err != nil {
				log.Println("crawl: failed to add result", err, item.Origin, item.URL, u)
			}
		}

		if util.CanSkipMime(kind) {
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
			if err := urlClient.AddPending(u, q.Origin); err != nil {
				log.Println("crawl: failed to add pending URL", err)
			}

			c.urlQueuePub.Send(q)
		}
	}
}
