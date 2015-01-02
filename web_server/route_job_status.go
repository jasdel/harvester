package main

import (
	"fmt"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"github.com/jasdel/harvester/internal/util"
	"github.com/zenazn/goji/web"
	"log"
	"net/http"
)

type jobStatusMsg struct {
	Completed int    `json:"completed"`
	Pending   int    `json:"pending"`
	Elapsed   string `json:"elapsed"`
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
func routeJobStatus(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := jobIdFromString(c.URLParams["jobId"])
	if err != nil {
		log.Println("routeJobStatus status request failed.", err)
		writeJSONError(w, "BadRequest", err.Error(), http.StatusBadRequest)
		return
	}

	status, jobErr := jobStatus(id)
	if jobErr != nil {
		log.Println("routeJobStatus request job status failed.", jobErr)
		writeJSONError(w, "DependancyFailure", jobErr.Short(), http.StatusInternalServerError)
		return
	}

	// Write job status out
	writeJSON(w, jobStatusMsg{
		Completed: status.Completed,
		Pending:   status.Pending,
		Elapsed:   status.Elapsed.String(),
	}, http.StatusOK)
}

// Connects to the remote service hosting job information, and
// the job's current status information.
func jobStatus(id types.JobId) (*types.JobStatus, *util.Error) {
	c := storage.NewClient()
	job := c.ForJob(id)

	status, err := job.Status()
	if err != nil {
		return nil, &util.Error{
			Source: "jobStatus",
			Info:   fmt.Sprintf("Failed to get job %d status", id),
			Err:    err,
		}
	}

	return status, nil
}
