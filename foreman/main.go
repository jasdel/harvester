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
	queueRecv, err := queue.NewReceiver(cfg.RecvQueue.ConnURL, cfg.RecvQueue.Topic)
	if err != nil {
		log.Fatalln("Queue Receiver initialization failed:", err)
	}
	defer queueRecv.Close()

	// Initialize the queue publisher to publish the filtered URLs
	// to the workers that will perform the crawling
	queuePub, err := queue.NewPublisher(cfg.PubQueue.ConnURL, cfg.PubQueue.Topic)
	if err != nil {
		log.Fatalln(err)
	}
	defer queuePub.Close()

	// Initialize the storage so the known and previously crawled state of URLs
	// can be determined.
	sc, err := storage.NewClient(cfg.StorageConfig)
	if err != nil {
		log.Fatalln("Storage NewClient failed:", err)
	}

	log.Println("Ready: Waiting for URL queue items...")
	for {
		item := <-queueRecv.Receive()
		log.Printf("Foreman: Queue URL: %s, from: %s, origin: %s, level: %d", item.URL, item.Refer, item.Origin, item.Level)

		su := sc.ForURL(item.URL)
		if crawled, _ := su.Crawled(); crawled {
			log.Println("Foreman: URL already known and crawled, skipping", item.URL, item.Origin, item.Level)
			su.DeletePending(item.Origin)
			// TODO need to check if this was the last job, and if so mark as complete
			// Get all URLs where this URL is the refer, and enqueue them
			continue
		} else if known, _ := su.Known(); !known {
			su.Add(item.Refer, storage.DefaultURLMime)
		}

		queuePub.Send(item)
	}
}

type Config struct {
	StorageConfig storage.ClientConfig `json:"storage"`
	RecvQueue     types.QueueConfig    `json:"recvQueue"`
	PubQueue      types.QueueConfig    `json:"pubQueue"`
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
