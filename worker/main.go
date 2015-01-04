package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jasdel/harvester/internal/common"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"log"
	"os"
	"time"
)

func main() {
	cfgFilename := flag.String("config", "config.json", "The web server configuration file.")
	flag.Parse()

	cfg, err := LoadConfig(*cfgFilename)
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize the queue receiver of the filter URLs from the foreman.
	// URLs received from this queue will be crawled
	workQueueRecv, err := queue.NewReceiver(cfg.WorkQueue.ConnURL, cfg.WorkQueue.Topic)
	if err != nil {
		log.Fatalln("Worker Queue Receiver: initialization failed:", err)
	}
	defer workQueueRecv.Close()

	// Initialize the queue publisher for publishing descendants of
	// a previously queued URL to be queued for crawling
	urlQueuePub, err := queue.NewPublisher(cfg.URLQueue.ConnURL, cfg.URLQueue.Topic)
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

	crawler := NewCrawler(urlQueuePub, sc, cfg.MaxLevel)

	log.Println("Ready: Waiting for URL work items...")
	for {
		item := <-workQueueRecv.Receive()
		crawler.Crawl(item)

		<-time.After(cfg.WorkDelay)
	}
}

// TODO document these fields
type Config struct {
	StorageConfig storage.ClientConfig `json:"storage"`
	WorkQueue     common.QueueConfig   `json:"workQueue"`
	URLQueue      common.QueueConfig   `json:"urlQueue"`
	MaxLevel      int                  `json:"maxLevel"`
	WorkDelayStr  string               `json:"workDelay"`
	WorkDelay     time.Duration        `json:"-"`
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
