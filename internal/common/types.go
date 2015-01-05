package common

import (
	"fmt"
	"time"
)

// Job Id, Use for identifying and searching for job records.
type JobId int64

// satisfies the stringer interface
func (id JobId) String() string {
	return fmt.Sprintf("%d", id)
}

type URLId int64

// satisfies the stringer interface
func (id URLId) String() string {
	return fmt.Sprintf("%d", id)
}

const DefaultURLMime = ``

// Invalid job state.  Any job with an id of this should not be processed.
const InvalidId = -1

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

	// Mapping of individual URL status.  A true for a URL means that
	// it has been processed, and only the false, URLs are pending.
	URLs map[string]bool
}

// Result map for a Job.  The map contains a mapping between refer URL and a list
// of all direct descendant URL which are linked on the refer URL's page.
type JobResults map[string][]string

// URL task to be queued for processing. This item will be processed by the foreman
// and sent to workers to crawl.
type URLQueueItem struct {
	// Initial Job URL which spawned the recursive chain of URL items to be queued
	// Origin   string `json:"origin"`
	OriginId URLId `json:"originId"`

	// The URL which contained a link to this URL
	// Refer   string `json:"refer"`
	ReferId URLId `json:"referId"`

	// The URL to be processed
	// URL   string `json:"url"`
	URLId URLId `json:"urlId"`

	// The recursive distance this URL is from the Origin URL
	Level int `json:"level"`
}
