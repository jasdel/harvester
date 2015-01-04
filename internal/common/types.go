package common

import (
	"time"
)

// Job Id, Use for identifying and searching for job records.
type JobId int64

// Invalid job state.  Any job with an id of this should not be processed.
const InvalidJobId = -1

// Status of a job. Provides information on the job's pending vs complete tasks
// and the duration that the job has been running.  If the job is completed
// the elapsed time will be the duration the job ran.
type JobStatus struct {
	// Job unique identifier.
	Id JobId

	// Number of completed tasks belonging to this job.
	Completed int

	// Number of tasks still outstanding.  Once Pending reaches 0 there
	// are no longer any more tasks to be performed for a job.
	Pending int

	// Duration the job has been running.  If all tasks are completed
	// this will reflect the duration the job ran for.
	Elapsed time.Duration

	// Map or Job URL to completion status.  True if the task has been completed
	// and false if it is still pending.
	// e.g. "http://example.com": true
	URLs map[string]bool
}

// Result map for a Job.  The map contains a mapping between refer URL and a list
// of all direct descendant URL which are linked on the refer URL's page.
type JobResults map[string][]string

// URL task to be queued for processing. This item will be processed by the foreman
// and sent to workers to crawl.
type URLQueueItem struct {
	// Initial Job URL which spawned the recursive chain of URL items to be queued
	Origin string `json:"origin"`

	// The URL which contained a link to this URL
	Refer string `json:"refer"`

	// The URL to be processed
	URL string `json:"url"`

	// The recursive distance this URL is from the Origin URL
	Level int `json:"level"`
}
