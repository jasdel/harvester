package types

import (
	"time"
)

type JobId int64

const InvalidJobId = -1

type JobStatus struct {
	JobId     JobId
	Completed int
	Pending   int
	Elapsed   time.Duration
}

type JobResult struct {
	JobStatus
	Results map[string][]string
}

type URLQueueItem struct {
	Origin      string   `json:"origin"`
	Refer       string   `json:"refer"`
	URL         string   `json:"url"`
	Descendants []string `json:"descendants"`
	Level       int      `json:"level"`
}
