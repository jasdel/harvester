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

// URL Id, used for identifying and searching for URL records
type URLId int64

// satisfies the stringer interface
func (id URLId) String() string {
	return fmt.Sprintf("%d", id)
}

// Default mime type URL mimes are initialized to.
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
	// Id of job this URL item originated from
	JobId JobId `json:"jobId"`

	// Initial Job URL which spawned the recursive chain of URL items to be queued
	OriginId URLId `json:"originId"`

	// The URL which contained a link to this URL
	ReferId URLId `json:"referId"`

	// The URL to be processed
	URLId URLId `json:"urlId"`

	// The recursive distance this URL is from the Origin URL
	Level int `json:"level"`

	// Flag instructing the processors craw the URL regardless
	// if it has already been crawled. The force crawl flag should
	// be passed down to descendants to ensure they are also crawled.
	// Note: Does not apply to skipped mime types.
	ForceCrawl bool `json:"forceCrawl"`
}
