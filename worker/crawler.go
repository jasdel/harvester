package main

import (
	"fmt"
	"github.com/jasdel/harvester/internal/common"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"log"
	"net/http"
	"time"
)

// Searches for and extracts URLs from a page. Those URLs are then queued up for recursive
// crawling with maximum depth of the passed in max level.
type Crawler struct {
	urlQueuePub queue.Publisher
	sc          *storage.Client
	maxLevel    int
}

// Creates a new instance of the Crawler. The crawler is save to be run across multiple
// go-routines.
func NewCrawler(urlQueuePub queue.Publisher, sc *storage.Client, maxLevel int) *Crawler {
	return &Crawler{
		urlQueuePub: urlQueuePub,
		sc:          sc,
		maxLevel:    maxLevel,
	}
}

// Retrieves the content of the item URL scrapes it for URLs.  Those descendant URLs
// are then either added back into the URL queue or added directly to a job's results.
// The URLs will be added to the URL queue if when the passed in item's Level is incremented
// and won't breach the Max Level of distance from the origin URL.
//
// When a crawl is complete the associated pending URL with this item will be removed,
// and a check to determine if there are anymore pending URLs for the item's Origin
// will be made. If there are no longer any pending URLs the Origin's Job URL entry
// will be marked as completed.
func (c *Crawler) Crawl(item *common.URLQueueItem) {
	startedAt := time.Now()
	urlClient := c.sc.URLClient()

	defer func() {
		// Make sure the Job is cleaned up even in if an error happens.
		if err := urlClient.DeletePending(item.URLId, item.OriginId); err != nil {
			log.Println("crawl: Failed to delete pending record for", item.URLId, item.OriginId)
		}
		log.Println("crawl: Finished crawling of", item.URLId, item.Level, "duration", time.Now().Sub(startedAt).String())

		// If there are no more pending entries for this origin, all jobs which contain that
		// origin which are not already complete can be marked as complete.
		if complete, err := urlClient.UpdateJobURLIfComplete(item.OriginId); err != nil {
			log.Println("crawl: Failed to update if Job URL is complete", item.OriginId, err)
		} else if complete {
			log.Println("crawl: Marked Job URL as complete", item.OriginId)
		}

	}()

	urlRec, err := c.sc.URLClient().GetURLById(item.URLId)
	if err != nil || urlRec == nil {
		log.Println("Failed to get URL record for URLId", item.URLId)
		return
	}

	mime, urls, err := Scrape(urlRec.URL, http.DefaultClient)
	if err != nil {
		log.Println("crawl: Failed to request and scrape", item.URLId, urlRec.URL, err)
		return
	}

	log.Println("crawl: Request and Scape complete URL", item.URLId, urlRec.URL, "mime:", mime, "level", item.Level, "descendants", len(urls), "duration", time.Now().Sub(startedAt).String(), "error", err)

	// Update mime type for the URL
	if err := urlClient.MarkCrawled(item.URLId, mime); err != nil {
		log.Println("crawl: failed to add update URL's mime type", item.URLId, mime, err)
		return
	}
	// Update the local urlRec mime value so don't need to re-query for it.
	urlRec.Mime = mime

	// Collect the job ids for this origin so their results can be updated
	jobIds, err := urlClient.GetJobIdsForURLById(item.OriginId)
	if err != nil {
		log.Println("crawl: origin has no associated jobId", item.OriginId)
		return
	}

	// Only add items to the result if they are greater than the first layer
	// because the first layer is the URLs that are used to start a job,
	// so they do not make sense to be inserted into the results without a refer.
	if item.Level > 0 {
		urlClient.AddURLsToResults(jobIds, item.ReferId, []*storage.URL{urlRec})
	}

	if err := c.processURLDescendants(jobIds, item, urls); err != nil {
		log.Println("crawl: failed to process descendants", err)
	}
}

// Iterates over the raw URLs fond on the page. These URLs will be added back into the
// URL Queue if the max level distance from the origin hasn't been reached yet. If the
// level has been reached the URLs will be just added to the Origin's Job URL result.
func (c *Crawler) processURLDescendants(jobIds []common.JobId, referItem *common.URLQueueItem, urls []string) error {
	urlClient := c.sc.URLClient()

	for i := 0; i < len(urls); i++ {
		u := urls[i]

		kind := common.GuessURLsMime(u)
		urlRec, err := urlClient.GetOrAddURLByURL(u, kind)
		if err != nil {
			return fmt.Errorf("Failed to get or add URL", u)
		}

		// Link the descendant with the refer, Ignore errors about duplicates
		urlClient.AddLink(urlRec.Id, referItem.URLId)

		// Only process the URLs for queue, or skipping, if the max level would
		// wouldn't be reached yet.
		if referItem.Level+1 < c.maxLevel {
			if common.CanSkipMime(kind) {
				urlClient.AddURLsToResults(jobIds, referItem.URLId, []*storage.URL{urlRec})
			}

			q := &common.URLQueueItem{
				OriginId:   referItem.OriginId,
				ReferId:    referItem.URLId,
				URLId:      urlRec.Id,
				Level:      referItem.Level + 1,
				ForceCrawl: referItem.ForceCrawl,
			}
			if err := urlClient.AddPending(urlRec.Id, q.OriginId); err != nil {
				log.Println("crawl: failed to add pending URL", err)
			}

			c.urlQueuePub.Send(q)
		} else {
			// For any URL that will not be enqueued, add it as a result instead
			urlClient.AddURLsToResults(jobIds, referItem.URLId, []*storage.URL{urlRec})
		}
	}

	return nil
}
