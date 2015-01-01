package main

import (
	"database/sql"
	"fmt"
	"github.com/jasdel/harvester/worker/scraper"
	"github.com/lib/pq"
	"net/http"
	"time"
)

const queryPopURLQueueTop = `select id,url from url_queue where not processed and pg_try_advisory_xact_lock(id) for update limit 1`

func main() {
	connInfo := `user=docker password=docker dbname=docker host=localhost port=49154 sslmode=disable`

	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}

	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			fmt.Println(err)
		}
	}

	listener := pq.NewListener(connInfo, 10*time.Second, time.Minute, reportProblem)
	if err := listener.Listen("url_queue_watchers"); err != nil {
		panic(err)
	}

	fmt.Println("Entering main loop")
	for {
		// process all available work before waiting for notifications again.
		getURL(db)
		waitForNotification(listener)
	}
}

func waitForNotification(l *pq.Listener) {
	for {
		select {
		case <-l.Notify:
			fmt.Println("received notification, new work available")
			return
		case <-time.After(90 * time.Second):
			go func() {
				l.Ping()
			}()
			// Check if there is work available. Just in case the listener
			// has not noticed a connection loss, and needs to reconnect.
			fmt.Println("received no work for 90 seconds, checking for new work")
			return
		}
	}
}

func getURL(db *sql.DB) {
	for {
		var id sql.NullInt64
		var url sql.NullString
		if err := db.QueryRow(queryPopURLQueueTop).Scan(&id, &url); err != nil {
			if err != sql.ErrNoRows {
				fmt.Println("Failed to query for url queue", err)
			}
			time.Sleep(10 * time.Second)
			return
		}
		if !id.Valid || !url.Valid {
			fmt.Println("No more work to do")
			return
		}

		if _, err := db.Exec("update url_queue set processed = true where id = $1", id); err != nil {
			fmt.Println("Failed to update processed", id, err)
		}

		go doWork(url.String)
	}
}

func doWork(url string) {
	mime, urls, err := scraper.Scrape(url, http.DefaultClient)

	fmt.Println("mime:", mime, "url count", len(urls), "err:", err)
}
