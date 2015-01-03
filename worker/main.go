package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
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
	queueRecv, err := queue.NewReceiver(cfg.RecvQueue.ConnURL, cfg.RecvQueue.Topic)
	if err != nil {
		log.Fatalln("Worker Queue Receiver: initialization failed:", err)
	}
	defer queueRecv.Close()

	// Initialize the queue publisher for publishing descendants of
	// a previously queued URL to be queued for crawling
	queuePub, err := queue.NewPublisher(cfg.PubQueue.ConnURL, cfg.PubQueue.Topic)
	if err != nil {
		log.Fatalln("Worker Queue Publisher: initialization failed:", err)
	}
	defer queuePub.Close()

	// Initialize the storage for determining the status of a URL,
	// updating URL values, and Job completeness status
	sc, err := storage.NewClient(cfg.StorageConfig)
	if err != nil {
		log.Fatalln("Worker Storage Client: initialization failed:", err)
	}

	crawler := Crawler{queuePub: queuePub, sc: sc, maxLevel: cfg.MaxLevel}

	log.Println("Ready: Waiting for URL work items...")
	for {
		item := <-queueRecv.Receive()
		crawler.crawl(item)

		<-time.After(cfg.WorkDelay)
	}
}

// TODO document these fields
type Config struct {
	StorageConfig storage.ClientConfig `json:"storage"`
	RecvQueue     types.QueueConfig    `json:"recvQueue"`
	PubQueue      types.QueueConfig    `json:"pubQueue"`
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
