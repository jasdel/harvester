package main

import (
	"encoding/json"
	"flag"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"log"
	"net/http"
	"os"
	"path"
)

// Web server for exposing an interface for scheduling jobs, checking their status, and
// receiving their result.
//
// The endpoints exposed are:
// POST: /
//		- Schedule Job. Body is newline separated list of URls to scheduled to be crawled.
//
// GET: /status/:jobId
//		- Get the status of an already scheduled job.
//
// GET: /result/:jobId
//		- Get the result of an already scheduled job
//
// Queues Used:
// Publish to URL Queue:
// Scheduled Job URLs will be sent to the URL Queue to be filtered and later crawled.
//
func main() {
	cfgFilename := flag.String("config", "config.json", "The web server configuration file.")
	hostAddr := flag.String("addr", "", "Host address to override config file")
	flag.Parse()

	cfg, err := LoadConfig(*cfgFilename)
	if err != nil {
		log.Fatalln(err)
	}

	// Allow the host address to be overridden via command line, for multiple instances
	if *hostAddr != "" {
		cfg.HTTPAddr = *hostAddr
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
	defer sc.Close()

	// Create the HTTP handlers to be able to provide an interface for serving
	// job schedule, status, and result requests. The Trailing '/' have to be append
	// because path.Join will strip off the trailing '/'
	http.Handle(path.Join("/", cfg.HTTPRootPath), &JobScheduleHandler{urlQueuePub: urlQueuePub, sc: sc})
	http.Handle(path.Join("/", cfg.HTTPRootPath, "status")+"/", &JobStatusHandler{sc: sc})
	http.Handle(path.Join("/", cfg.HTTPRootPath, "result")+"/", &JobResultHandler{sc: sc})

	log.Println("Listening on", cfg.HTTPAddr)
	if err := http.ListenAndServe(cfg.HTTPAddr, nil); err != nil {
		log.Fatalln(err)
	}
}

// Provides the web server's configuration information. For connecting to
// Queues, storage, and other runtime settings.
type Config struct {
	// Storage connection configuration
	StorageConfig storage.ClientConfig `json:"storage"`

	// URL queue for publishing scheduled job URLs to the foreman
	URLQueueConfig queue.QueueConfig `json:"urlQueue"`

	// HTTP address to service content from
	HTTPAddr string `json:"httpAddr"`

	// Root path the HTTP routes should be based of of. Useful when
	// nesting the service behind a reverse proxy
	HTTPRootPath string `json:"httpRootPath"`
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
