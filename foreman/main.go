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

// Foreman is an intermediate Queue filter which filters out URLs which have previously
// been crawled.
//
// Queues Used:
// Receive from URL Queue:
// The URL Queue provides URLQueueItems that will be filtered
// and sent to the worker for processing. A check against the cache will be made to ensure
// the same URL is not crawled multiple times.
//
// Publish to URL Queue:
// If an item is cached or can be skipped, its descendants will
// be published to the URL Queue so they can be processed.
//
// Publish to Work Queue:
// Once a URL item is filtered, and not cached it will be sent
// to the Work Queue to be crawled.
//
func main() {
	// Configuration file containing all basic configuration for a server instance to run
	cfgFilename := flag.String("config", "config.json", "The foreman configuration file.")

	flag.Parse()
	cfg, err := LoadConfig(*cfgFilename)
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize the queue receiver to receive URLs that are being
	// queue to be crawled
	urlQueueRecv, err := queue.NewReceiver(cfg.URLQueueConfig)
	if err != nil {
		log.Fatalln("Queue Receiver initialization failed:", err)
	}
	defer urlQueueRecv.Close()

	// If queued items have already been crawled, will need to find descendants,
	// and enqueue them.
	urlQueuePub, err := queue.NewPublisher(cfg.URLQueueConfig)
	if err != nil {
		log.Fatalln("Queue Publisher initialization failed:", err)
	}
	defer urlQueuePub.Close()

	// Initialize the queue publisher to publish the filtered URLs
	// to the workers that will perform the crawling
	workQueuePub, err := queue.NewPublisher(cfg.WorkQueueConfig)
	if err != nil {
		log.Fatalln("Worker Queue Publisher initialization failed", err)
	}
	defer workQueuePub.Close()

	// Initialize the storage so the known and previously crawled state of URLs
	// can be determined.
	sc, err := storage.NewClient(cfg.StorageConfig)
	if err != nil {
		log.Fatalln("Storage NewClient failed:", err)
	}
	defer sc.Close()

	foreman := NewForeman(workQueuePub, urlQueuePub, sc, cfg.MaxLevel, cfg.CacheMaxAge)

	log.Println("Ready: Waiting for URL queue items...")
	for {
		item := <-urlQueueRecv.Receive()
		foreman.ProcessQueueItem(item)
	}
}

// Provides the Foreman's configuration information. For connecting to
// Queues, storage, and other runtime settings.
type Config struct {
	// Storage connection configuration
	StorageConfig storage.ClientConfig `json:"storage"`

	// Queue for receiving queue request from the web server, worker,
	// and from foreman if the refer URL had already been crawled.
	URLQueueConfig queue.QueueConfig `json:"urlQueue"`

	// Queue for sending URI items from  the foreman's to workers
	WorkQueueConfig queue.QueueConfig `json:"workQueue"`

	// the maximum level the crawling should be allowed to travel
	MaxLevel int `json:"maxLevel"`

	// Maximum age a URL can be cached for before it is allowed to
	// e.g: 1m23s for 1 minute and 23 seconds
	// See http://golang.org/pkg/time/#ParseDuration for formatting
	CacheMaxAgeStr string `json:"cacheMaxAge"`

	// The CacheMaxAgeStr will be parsed, and its value placed into the CacheMaxAge field.
	// Used to determine maximum age to cache a URL for before it is crawled again.
	CacheMaxAge time.Duration `json:"-"`
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

	if cfg.CacheMaxAgeStr != "" {
		cfg.CacheMaxAge, err = time.ParseDuration(cfg.CacheMaxAgeStr)
		if err != nil {
			return cfg, fmt.Errorf("%s, %s", err.Error(), cfg.CacheMaxAgeStr)
		} else if cfg.CacheMaxAge < 0 {
			return cfg, fmt.Errorf("Invalid work delay, must be positive", cfg.CacheMaxAgeStr)
		}
	}

	return cfg, nil
}
