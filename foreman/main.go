package main

import (
	"fmt"
	"github.com/apcera/nats"
	"github.com/jasdel/harvester/internal/queue"
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

	for {
		item := <-recv.Receive()
		fmt.Println(item)
		send.Send(item)
	}
}
