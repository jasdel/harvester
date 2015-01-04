package main

import (
	"encoding/json"
	"flag"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"log"
	"net/http"
	"os"
)

// Web server for exposing an interface for scheduling jobs, checking their status, and
// receiving their result. The endpoints exposed are:
//
// POST: /
//		- Schedule Job. Body is newline separated list of URls to scheduled to be crawled.
//
// GET: /status/:jobId
//		- Get the status of a previously scheduled job.
//
// GET: /result/:jobId
//		- Get the result of a previously scheduled job
//
func main() {
	cfgFilename := flag.String("config", "config.json", "The web server configuration file.")
	flag.Parse()

	cfg, err := LoadConfig(*cfgFilename)
	if err != nil {
		log.Fatalln(err)
	}

	// Initialize the queue for publishing scheduled Job URLs
	urlQueuePub, err := queue.NewPublisher(cfg.URLQueueConfig)
	if err != nil {
		log.Fatalln("Queue Publisher initialization failed:", err)
	}
	defer urlQueuePub.Close()

	// Initialize the storage for checking the status and results of jobs
	sc, err := storage.NewClient(cfg.StorageConfig)
	if err != nil {
		log.Fatalln("Storage NewClient failed:", err)
	}

	// Create the HTTP handlers to be able to provide an interface for serving
	// job schedule, status, and result requests.
	http.Handle("/", &JobScheduleHandler{urlQueuePub: urlQueuePub, sc: sc})
	http.Handle("/status/", &JobStatusHandler{sc: sc})
	http.Handle("/result/", &JobResultHandler{sc: sc})

	log.Println("Listening on", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, nil); err != nil {
		log.Fatalln(err)
	}
}

type Config struct {
	StorageConfig  storage.ClientConfig `json:"storage"`
	URLQueueConfig queue.QueueConfig    `json:"urlQueue"`
	HTTPAddr       string               `json:"httpAddr"`
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
