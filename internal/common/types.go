package common

import (
	"time"
)

type JobId int64

const InvalidJobId = -1

type JobStatus struct {
	Id        JobId
	Completed int
	Pending   int
	Elapsed   time.Duration
	URLs      map[string]bool
}

type JobResults map[string][]string

type URLQueueItem struct {
	Origin string `json:"origin"`
	Refer  string `json:"refer"`
	URL    string `json:"url"`
	Level  int    `json:"level"`
}

type URLItem struct {
	URL  string
	Mime string
}

type QueueConfig struct {
	Topic   string `json:"topic"`
	ConnURL string `json:"connURL"`
}
