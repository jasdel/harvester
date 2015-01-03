package main

import (
	"github.com/apcera/nats"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"log"
)

func main() {
	recv, err := queue.NewReceiver(nats.DefaultURL, "url_queue")
	if err != nil {
		panic(err)
	}
	send, err := queue.NewPublisher(nats.DefaultURL, "work_queue")
	if err != nil {
		panic(err)
	}

	sc := storage.NewClient()

	for {
		item := <-recv.Receive()
		log.Printf("Foreman: Queue URL: %s, from: %s, origin: %s, level: %d", item.URL, item.Refer, item.Origin, item.Level)

		su := sc.ForURL(item.URL)
		if crawled, _ := su.Crawled(); crawled {
			log.Println("Foreman: URL already known and crawled, skipping", item.URL, item.Origin, item.Level)
			su.DeletePending(item.Origin)
			continue
		} else if known, _ := su.Known(); !known {
			su.Add(item.Refer, storage.DefaultURLMime)
		}

		send.Send(item)
	}
}
