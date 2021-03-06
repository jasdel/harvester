package main

import (
	"fmt"
	"github.com/jasdel/harvester/internal/common"
	"github.com/jasdel/harvester/internal/storage"
	"log"
	"net/http"
	"path"
)

// Response to a successful request of a Job
type jobStatusMsg struct {
	// The Number of completely crawled Job URLs
	Completed int `json:"completed"`

	// The number of Job URLs pending completion.
	Pending int `json:"pending"`

	// The amount of time that the Job has been processing for.
	Elapsed string `json:"elapsed"`

	// Mapping of individual URL status.  A true for a URL means that
	// it has been processed, and only the false, URLs are pending.
	URLs map[string]bool `json:"urls"`
}

// Handles the request checking on the status of a previously scheduled job.
// Returns an error if the job isn't found, or invalid input. If the job
// exists its status will be returned. If the job does not exists a 404 status
// code and message will be returned.
//
// e.g:
// curl -X GET "http://localhost:8080/status/1234"
//
// Response:
//	- Success: {completed: 2, pending: 3, elapsed: 5m10s, urls: { <url>: <complete> } }
//	- Failure: {code: <code>, message: <message>}
type JobStatusHandler struct {
	sc *storage.Client
}

func (h *JobStatusHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.Header().Set("Allow", "GET")
		http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
		return
	}

	id, err := jobIdFromString(path.Base(r.URL.Path))
	if err != nil {
		log.Println("routeJobStatus status request failed.", err)
		writeJSONError(w, "BadRequest", err.Error(), http.StatusBadRequest)
		return
	}

	status, jobErr := h.jobStatus(id)
	if jobErr != nil {
		log.Println("routeJobStatus request job status failed.", jobErr)
		writeJSONError(w, "NotFound", jobErr.Short(), http.StatusNotFound)
		return
	}

	// Write job status out
	writeJSON(w, jobStatusMsg{
		Completed: status.Completed,
		Pending:   status.Pending,
		URLs:      status.URLs,
		Elapsed:   status.Elapsed.String(),
	}, http.StatusOK)
}

// Connects to the remote service hosting job information, and
// the job's current status information.
func (h *JobStatusHandler) jobStatus(id common.JobId) (*common.JobStatus, *ErroMsg) {
	job, err := h.sc.JobClient().GetJob(id)
	if err != nil || job == nil {
		return nil, &ErroMsg{
			Source: "jobStatus",
			Info:   fmt.Sprintf("Failed to get job %d status", id),
			Err:    err,
		}
	}

	return job.Status(), nil
}
