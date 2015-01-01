package main

import (
	"github.com/apcera/nats"
	"strings"
	// "github.com/lib/pq"
	// "database/sql"
	"fmt"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/worker/scraper"
	"net/http"
)

func main() {
	// connInfo := `user=docker password=docker dbname=docker host=localhost port=24001 sslmode=disable`

	// db, err := sql.Open("postgres", connInfo)
	// if err != nil {
	// 	panic(err)
	// }
	queueRecv, err := queue.NewReceiver(nats.DefaultURL, "work_queue")
	if err != nil {
		panic(err)
	}

	for {
		item := <-queueRecv.Receive()
		doWork(item)
	}

}

func doWork(item *types.URLQueueItem) {
	mime, urls, err := scraper.Scrape(url, http.DefaultClient)

	if strings.HasPrefix(mime, "image") {
		// TODO insert into url table as image
	}

	fmt.Println("mime:", mime, "url count", len(urls), "err:", err)
}
