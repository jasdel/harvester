package main

import (
	"fmt"
	"github.com/apcera/nats"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
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

		su := sc.ForURL(item.URL)

		if known, _ := su.Known(); !known {
			su.Add(item.Refer, storage.DefaultURLMime)
		} else {

		}

		fmt.Println(item)
		send.Send(item)
	}
}
