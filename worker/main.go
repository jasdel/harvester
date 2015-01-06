package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"log"
	"os"
	"time"
)

// Worker pulls for the Worker queue, crawls the URLs and enqueues any descendant URLs
// for further crawling. Results of crawling the URLs are also added to the origin Job
// URL's result set.
//
// Queues Used:
// Receives from Worker Queue:
// Work items will be received by the workers, and crawled
// If crawling the work item produces any URLs those URLs will be enqueued for processing
// or added directly to the origin Job URL's result based on the level depth from their origin.
//
// Publish to URL Queue:
// If crawling a work item produces any descendant URLs those URLs will be enqueued to be
// crawled, or added to the origin Job URL's results.
//
func main() {
	// Configuration file containing all basic configuration for a server instance to run
	cfgFilename := flag.String("config", "config.json", "The web server configuration file.")

	flag.Parse()
	cfg, err := LoadConfig(*cfgFilename)
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize the queue receiver of the filter URLs from the foreman.
	// URLs received from this queue will be crawled
	workQueueRecv, err := queue.NewReceiver(cfg.WorkQueueConfig)
	if err != nil {
		log.Fatalln("Worker Queue Receiver: initialization failed:", err)
	}
	defer workQueueRecv.Close()

	// Initialize the queue publisher for publishing descendants of
	// a previously queued URL to be queued for crawling
	urlQueuePub, err := queue.NewPublisher(cfg.URLQueueConfig)
	if err != nil {
		log.Fatalln("Worker Queue Publisher: initialization failed:", err)
	}
	defer urlQueuePub.Close()

	// Initialize the storage for determining the status of a URL,
	// updating URL values, and Job completeness status
	sc, err := storage.NewClient(cfg.StorageConfig)
	if err != nil {
		log.Fatalln("Worker Storage Client: initialization failed:", err)
	}
	defer sc.Close()

	crawler := NewCrawler(urlQueuePub, sc, cfg.MaxLevel)

	log.Println("Ready: Waiting for URL work items...")
	for {
		item := <-workQueueRecv.Receive()
		crawler.Crawl(item)

		<-time.After(cfg.WorkDelay)
	}
}

// Provides the Foreman's configuration information. For connecting to
// Queues, storage, and other runtime settings.
type Config struct {
	StorageConfig storage.ClientConfig `json:"storage"`

	// Queue to receive work from from the foreman(s). The URLQueueItems
	// will be pulled off of this queue and crawled.
	WorkQueueConfig queue.QueueConfig `json:"workQueue"`

	// Queue to publish URLs to be crawled which were found when scrapping
	// a previously queued work URLQueueItem
	URLQueueConfig queue.QueueConfig `json:"urlQueue"`

	// the maximum level the crawling should be allowed to travel
	MaxLevel int `json:"maxLevel"`

	// Delay before requesting additional work. Indented to prevent
	// flooding domain's with too many requests back to back.
	// time.Duration string formated value.
	// e.g: 1m23s for 1 minute and 23 seconds
	// See http://golang.org/pkg/time/#ParseDuration for formatting
	WorkDelayStr string `json:"workDelay"`

	// The WorkDelayStr will be parsed, and its value placed into the WorkDelay field.
	// Used to provide delay between accepting more work.
	WorkDelay time.Duration `json:"-"`
}

// Loads the configuration file from disk in as a JSON blob.
func LoadConfig(filename string) (Config, error) {
	cfg := Config{}

	file, err := os.Open(filename)
	if err != nil {
		return cfg, err
	}
	defer file.Close()

	if err = json.NewDecoder(file).Decode(&cfg); err != nil {
		return cfg, err
	}

	if cfg.WorkDelayStr != "" {
		cfg.WorkDelay, err = time.ParseDuration(cfg.WorkDelayStr)
		if err != nil {
			return cfg, fmt.Errorf("%s, %s", err.Error(), cfg.WorkDelayStr)
		} else if cfg.WorkDelay < 0 {
			return cfg, fmt.Errorf("Invalid work delay, must be positive", cfg.WorkDelayStr)
		}
	}

	return cfg, nil
}
