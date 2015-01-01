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
