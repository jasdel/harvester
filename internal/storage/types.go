package storage

import (
	"github.com/jasdel/harvester/internal/common"
	"time"
)

// Definition of a 'url' table record.  A URL is defined as a URL + Refer URL
// that the URL was found on.
type URL struct {
	// ID (primary key) of this entry
	Id int64

	// This URL the record is for
	URL string

	// The URL which this URL record was found on
	Refer string

	// The Content type of the URL, e.g: text/html
	Mime string

	// If the URL has been crawled the Crawled flag
	// will be true, This includes if it was crawled
	// but found no descendants.
	Crawled bool

	// The time stamp the URL entry was created.
	CreatedOn time.Time
}

// Job Entry for the 'job' record. The Job also includes the
// URLs that were specified as tasks of a Job.
type Job struct {
	// ID (primary key) of the job
	Id common.JobId

	// List of URLs belonging to this job. Includes their
	// competition status.
	URLs []JobURL

	// The time stamp the Job was created on.
	CreatedOn time.Time
}

// Returns the status of the job.  The status includes the progress
// of completed vs pending, and total elapsed time.
func (j *Job) Status() *common.JobStatus {
	status := &common.JobStatus{Id: j.Id}
	status.URLs = make(map[string]bool)
	var compTime time.Time
	for _, u := range j.URLs {
		if u.Completed {
			status.Completed++
			if compTime.Before(u.CompletedOn) {
				compTime = u.CompletedOn
			}
		} else {
			status.Pending++
		}
		status.URLs[u.URL] = u.Completed
	}

	if status.Pending != 0 {
		compTime = time.Now().UTC()
	}
	status.Elapsed = compTime.Sub(j.CreatedOn)

	return status
}

// Job URL entry for the _'job_url' table. The CompletedOn value will only
// be valid if the 'Completed' flag is true.
type JobURL struct {
	// Job URL that was requested
	URL string

	// If this Job URL has been completely crawled
	Completed bool

	// The time stamp the URL was finished crawling. Only valid if 'Completed'
	// is also set.
	CompletedOn time.Time

	// The JobId this URL belongs to.
	JobId common.JobId
}
