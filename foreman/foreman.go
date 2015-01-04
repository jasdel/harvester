package main

import (
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"log"
)

type Foreman struct {
	queuePub queue.Publisher
	sc       *storage.Client
}

func NewForeman(queuePub queue.Publisher, sc *storage.Client) *Foreman {
	return &Foreman{
		queuePub: queuePub,
		sc:       sc,
	}
}

func (f *Foreman) ProcessQueueItem(item *types.URLQueueItem) {
	urlClient := f.sc.URLClient()

	url, err := urlClient.GetURLWithRefer(item.URL, item.Refer)
	if err != nil {
		log.Println("Failed to get URL", err)
		return
	}

	if url != nil {
		if url.Crawled {
			log.Println("URL already known and crawled, skipping, adding descendants", item.URL, item.Refer)
			urlClient.DeletePending(item.URL, item.Origin)

			// If this item has already been crawled and it has a refer,
			// add the item to the results for all not completed origin job URLs.
			// Not having a refer is an origin URL from a job.
			if item.Refer != "" {
				if err := addToResults(urlClient, item, url); err != nil {
					log.Println("Failed to add already crawled item to results", item.URL, err)
				}
			}
			// TODO need to check if this was the last job, and if so mark as complete
			// Get all URLs where this URL is the refer, and enqueue them, if none, run check for origin complete
			return
		}
	} else {
		urlClient.Add(item.URL, item.Refer, storage.DefaultURLMime)
	}

	f.queuePub.Send(item)
}

// Adds the item to all Jobs associated with the item's origin results
func addToResults(urlClient *storage.URLClient, item *types.URLQueueItem, knownURL *storage.URL) error {
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
