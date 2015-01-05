package main

import (
	"fmt"
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
	log.Printf("Foreman: Queue URL: %s, from: %s, origin: %s, level: %d", item.URLId, item.ReferId, item.OriginId, item.Level)

	urlRec, err := urlClient.GetURLById(item.URLId)
	if err != nil || urlRec == nil {
		log.Println("Foreman: Failed to get URL", item.URLId, err)
		return
	}

	// If the item URL has already been crawled or a mime type
	// that can be skipped, use the cache instead.
	if urlRec.Crawled || common.CanSkipMime(urlRec.Mime) {
		f.processFromCache(item, urlRec)
		return
	}

	f.workQueuePub.Send(item)
}

// If an item is being processed from the cache this will determine if that item's descendants
// should be added the job results, or queued to be crawled them selves.
func (f *Foreman) processFromCache(item *common.URLQueueItem, urlRec *storage.URL) {
	log.Println("Foreman: URL already known and crawled, skipping, checking descendants", item.URLId, item.ReferId)
	urlClient := f.sc.URLClient()

	defer func() {
		// Make sure the Job is cleaned up even in if an error happens.
		if err := urlClient.DeletePending(item.URLId, item.OriginId); err != nil {
			log.Println("Foreman: Failed to delete pending record for", item.URLId, item.OriginId)
		}

		// If there are no more pending entries for this origin, all jobs which contain that
		// origin which are not already complete can be marked as complete.
		if complete, err := urlClient.UpdateJobURLIfComplete(item.OriginId); err != nil {
			log.Println("Foreman: Failed to update if Job URL is complete", item.OriginId, err)
		} else if complete {
			log.Println("Foreman: Marked Job URL as complete", item.OriginId)
		}
	}()

	// Get the job Ids which are associated with the origin of this item
	// so that items can be added to their results
	jobIds, err := urlClient.GetJobIdsForURLById(item.OriginId)
	if err != nil {
		log.Println("Foreman: Failed to get job ids associated with  origin", item.OriginId)
		return
	}

	// Only add items to the result if they are greater than the first layer
	// because the first layer is the URLs that are used to start a job,
	// so they do not make sense to be inserted into the results without a refer.
	if item.Level > 0 {
		urlClient.AddURLsToResults(jobIds, item.ReferId, []*storage.URL{urlRec})
	}

	if err := f.processDescendants(jobIds, item); err != nil {
		log.Println("Foreman: Failed to process known queued item's descendants", item.URLId, err)
		return
	}
}

// Processes descendants of a URL which is both known and already crawled.
// The descendants will be either added to the urlQueue if the maxLevel hasn't
// been reached yet, or will be just added as results to
func (f *Foreman) processDescendants(jobIds []common.JobId, item *common.URLQueueItem) error {
	urlClient := f.sc.URLClient()

	// Get all URLs where this item is a refer to, so that they can be queued
	// for crawling.
	urlRecs, err := urlClient.GetAllURLsWithReferById(item.URLId)
	if err != nil {
		return fmt.Errorf("Failed to get URL descendants of", item.URLId, err)
	}

	// Get all URLs where this URL is the refer, and enqueue them. But if the
	// level would exceed the max, just add the descendants to the results.
	if item.Level+1 < f.maxLevel {
		log.Println("enqueue descendants")
		if err := f.enqueueURLs(item, urlRecs); err != nil {
			return fmt.Errorf("Failed to enqueue URLs", err)
		}
	} else {
		log.Println("Adding descendants to results")
		urlClient.AddURLsToResults(jobIds, item.URLId, urlRecs)
	}

	return nil
}

// Enqueue a list of URLs with a single refer.  The URLs are added to both the
// pending Job, and urlQueue.
func (f *Foreman) enqueueURLs(refer *common.URLQueueItem, urls []*storage.URL) error {
	urlClient := f.sc.URLClient()

	for _, u := range urls {
		q := &common.URLQueueItem{
			OriginId: refer.OriginId,
			ReferId:  refer.URLId,
			URLId:    u.Id,
			Level:    refer.Level + 1,
		}
		if err := urlClient.AddPending(u.Id, q.OriginId); err != nil {
			return err
		}

		f.urlQueuePub.Send(q)
	}

	return nil
}
