package storage

import (
	"github.com/jasdel/harvester/internal/common"
	"time"
)

type URL struct {
	Id        int64
	URL       string
	Refer     string
	Mime      string
	Crawled   bool
	CreatedOn time.Time
}

type Job struct {
	Id        common.JobId
	URLs      []JobURL
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

type JobURL struct {
	URL         string
	Completed   bool
	CompletedOn time.Time
	JobId       common.JobId
}
