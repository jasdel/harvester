package main

import (
	"github.com/jasdel/harvester/internal/common"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"log"
)

// Provides filtering of the items before they are forwarded on to the worker queue.
type Foreman struct {
	// Queue to publish URL items to in order to be crawled.
	workQueuePub queue.Publisher

	// Queue to publish URL items to who's refer URL had already been crawled.
	urlQueuePub queue.Publisher

	// Storage client, for accessing, and manipulating the storage
	// For JobClients and URLClients
	sc *storage.Client

	// Maximum level already crawled items are allowed to have their descendants
	// queued.
	maxLevel int
}

// Creates a new instance of the foreman and returns it.  The foreman's methods
// are safe to be called across multiple go routines.
func NewForeman(workQueuePub queue.Publisher, urlQueuePub queue.Publisher, sc *storage.Client, maxLevel int) *Foreman {
	return &Foreman{
		workQueuePub: workQueuePub,
		urlQueuePub:  urlQueuePub,
		sc:           sc,
		maxLevel:     maxLevel,
	}
}

// Determines if a queued item should be filtered out because its already been crawled, or
// allowed to be sent to the worker queue. If the item was previously crawled it's descendants
// will be added to the queue if the maxLevel hasn't been reached yet.  If it has, the
// descendants will be just added to the job result list.
func (f *Foreman) ProcessQueueItem(item *common.URLQueueItem) {
	urlClient := f.sc.URLClient()
	log.Printf("Foreman: Queue URL: %s, from: %s, origin: %s, level: %d", item.URL, item.Refer, item.Origin, item.Level)

	url, err := urlClient.GetURLWithRefer(item.URL, item.Refer)
	if err != nil {
		log.Println("Failed to get URL", err)
		return
	}

	if url != nil {
		// If the item URL has already been crawled or a mime type
		// that can be skipped, use the cache instead.
		if url.Crawled || common.CanSkipMime(url.Mime) {
			f.processFromCache(item, url)
			return
		}
	} else {
		urlClient.Add(item.URL, item.Refer, storage.DefaultURLMime)
	}

	f.workQueuePub.Send(item)
}

// If an item is being processed from the cache this will determine if that item's descendants
// should be added the job results, or queued to be crawled them selves.
func (f *Foreman) processFromCache(item *common.URLQueueItem, url *storage.URL) {
	log.Println("URL already known and crawled, skipping, checking descendants", item.URL, item.Refer)
	urlClient := f.sc.URLClient()

	defer func() {
		// Make sure the Job is cleaned up even in if an error happens.
		if err := urlClient.DeletePending(item.URL, item.Origin); err != nil {
			log.Println("Failed to delete pending record for", item.URL, item.Origin)
		}

		// If there are no more pending entries for this origin, all jobs which contain that
		// origin which are not already complete can be marked as complete.
		if complete, err := urlClient.UpdateJobURLIfComplete(item.Origin); err != nil {
			log.Println("Failed to update if Job URL is complete", item.Origin, err)
		} else if complete {
			log.Println("Marked Job URL as complete", item.Origin)
		}
	}()

	// Only add items to the result if they are greater than the first layer
	// because the first layer is the URLs that are used to start a job,
	// so they do not make sense to be inserted into the results without a refer.
	if item.Level > 0 {
		if err := f.addToResults(item, url); err != nil {
			log.Println("Failed to add already crawled item to results", item.URL, err)
		}
	}

	if err := f.processDescendants(item); err != nil {
		log.Println("Failed to process known queued item's descendants", item.URL, err)
	}
}

// Adds the item to all Jobs associated with the item's origin results
func (f *Foreman) addToResults(item *common.URLQueueItem, knownURL *storage.URL) error {
	urlClient := f.sc.URLClient()

	jobIdsForOrigin, err := urlClient.GetJobIdsForURL(item.Origin)
	if err != nil {
		return err
	}
	for _, id := range jobIdsForOrigin {
		if err := urlClient.AddResult(id, item.URL, item.Refer, knownURL.Mime); err != nil {
			return err
		}
	}
	return nil
}

// Processes descendants of a URL which is both known and already crawled.
// The descendants will be either added to the urlQueue if the maxLevel hasn't
// been reached yet, or will be just added as results to
func (f *Foreman) processDescendants(item *common.URLQueueItem) error {
	urlClient := f.sc.URLClient()

	// Get all URLs where this item is a refer to, so that they can be queued
	// for crawling.
	urls, err := urlClient.GetAllURLsWithRefer(item.URL)
	if err != nil {
		log.Println("Failed to get URL descendants of", item.URL, err)
		return err
	}

	// Get all URLs where this URL is the refer, and enqueue them. But if the
	// level would exceed the max, just add the descendants to the results.
	if item.Level+1 < f.maxLevel {
		log.Println("enqueuing descendants")
		if err := f.enqueueURLs(item, urls); err != nil {
			log.Println("Failed to enqueue URLs", err)
			return err
		}
	} else {
		log.Println("Adding descendants to results")
		// Since the URLs won't be enqueued
		jobIds, err := urlClient.GetJobIdsForURL(item.Origin)
		if err != nil {
			log.Println("Failed to get associated JobIds", item.Origin, err)
			return err
		}
		if err := urlClient.AddURLsToResults(jobIds, item.URL, urls); err != nil {
			log.Println("Failed to add Job URL results", err)
			return err
		}
	}

	return nil
}

// Enqueue a list of URLs with a single refer.  The URLs are added to both the
// pending Job, and urlQueue.
func (f *Foreman) enqueueURLs(refer *common.URLQueueItem, urls []*storage.URL) error {
	urlClient := f.sc.URLClient()

	for _, u := range urls {
		// if common.CanSkipMime(u.Mime) {
		// 	// If the URL is a type that can be skipped and doesn't need to be
		// 	// crawled,
		// 	continue
		// }

		q := &common.URLQueueItem{
			Origin: refer.Origin,
			Refer:  refer.URL,
			URL:    u.URL,
			Level:  refer.Level + 1,
		}
		if err := urlClient.AddPending(u.URL, q.Origin); err != nil {
			return err
		}

		f.urlQueuePub.Send(q)
	}

	return nil
}
