package main

import (
	"fmt"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"github.com/jasdel/harvester/internal/util"
	"log"
	"net/http"
	"path"
)

type jobStatusMsg struct {
	Completed int             `json:"completed"`
	Pending   int             `json:"pending"`
	Elapsed   string          `json:"elapsed"`
	URLs      map[string]bool `json:"urls"`
}

// Handles the request checking on the status of a previously scheduled job.
// Returns an error if the job isn't found, or invalid input. If the job
// exists its status will be returned
//
// e.g:
// curl -X GET "http://localhost:8080/status/1234"
//
// Response:
//	- Success: {completed: 2, pending: 3, elapsed: 5m10s}
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
		Elapsed:   status.Elapsed.String(),
		URLs:      status.URLs,
	}, http.StatusOK)
}

// Connects to the remote service hosting job information, and
// the job's current status information.
func (h *JobStatusHandler) jobStatus(id types.JobId) (*types.JobStatus, *util.Error) {
	job, err := h.sc.JobClient().GetJob(id)
	if err != nil || job == nil {
		return nil, &util.Error{
			Source: "jobStatus",
			Info:   fmt.Sprintf("Failed to get job %d status", id),
			Err:    err,
		}
	}

	return job.Status(), nil
}
