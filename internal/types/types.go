package types

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
}

type JobResults map[string][]JobResult

type JobResult struct {
	Mime string `json:"mime"`
	URL  string `json:"url"`
}

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
