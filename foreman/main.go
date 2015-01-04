package main

import (
	"encoding/json"
	"flag"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"log"
	"os"
)

// Foreman is an intermediate Queue filter which filters out URLs which have previously
// been crawled.
func main() {
	cfgFilename := flag.String("config", "config.json", "The foreman configuration file.")
	flag.Parse()

	cfg, err := LoadConfig(*cfgFilename)
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize the queue receiver to receive URLs that are being
	// queue to be crawled
	urlQueueRecv, err := queue.NewReceiver(cfg.URLQueue.ConnURL, cfg.URLQueue.Topic)
	if err != nil {
		log.Fatalln("Queue Receiver initialization failed:", err)
	}
	defer urlQueueRecv.Close()

	// If queued items have already been crawled, will need to find descendants,
	// and enqueue them.
	urlQueuePub, err := queue.NewPublisher(cfg.URLQueue.ConnURL, cfg.URLQueue.Topic)
	if err != nil {
		log.Fatalln("Queue Publisher initialization failed:", err)
	}
	defer urlQueuePub.Close()

	// Initialize the queue publisher to publish the filtered URLs
	// to the workers that will perform the crawling
	workQueuePub, err := queue.NewPublisher(cfg.WorkQueue.ConnURL, cfg.WorkQueue.Topic)
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

	foreman := NewForeman(workQueuePub, urlQueuePub, sc, cfg.MaxLevel)

	log.Println("Ready: Waiting for URL queue items...")
	for {
		item := <-urlQueueRecv.Receive()
		foreman.ProcessQueueItem(item)
	}
}

type Config struct {
	// Storage connection configuration
	StorageConfig storage.ClientConfig `json:"storage"`

	// Queue for receiving queue request from the web server, worker,
	// and from foreman if the refer URL had already been crawled.
	URLQueue types.QueueConfig `json:"urlQueue"`

	// Queue for sending URI items from  the foremans to workers
	WorkQueue types.QueueConfig `json:"workQueue"`

	// the maximum level the crawling should be allowed to travel
	MaxLevel int `json:"maxLevel"`
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

	return cfg, nil
}
