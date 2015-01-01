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

type jobResultMsg struct {
	jobStatusMsg
	Results map[string][]string `json:"results"`
}

// Handles the request checking on the status of a previously scheduled job.
// Returns an error if the job isn't found, or invalid input. If the job
// exists its status will be returned
//
// e.g:
// curl -X GET "http://localhost:8080/results/1234"
//
// Response:
//	- Success: {completed: 2, pending: 3, elapsed: 5m10s, results: {<domain>: [<urls>]}}
//	- Failure: {code: <code>, message: <message>}
func routeJobResult(c web.C, w http.ResponseWriter, r *http.Request) {
	id, err := jobIdFromString(c.URLParams["jobId"])
	if err != nil {
		log.Println("routeJobStatus status request failed.", err)
		writeJSONError(w, "BadRequest", err.Error(), http.StatusBadRequest)
		return
	}

	result, jobErr := jobResult(id)
	if jobErr != nil {
		log.Println("routeJobResult request job result failed.", jobErr)
		writeJSONError(w, "DependancyFailure", jobErr.Short(), http.StatusInternalServerError)
		return
	}

	// Write job status out
	writeJSON(w, result, http.StatusOK)
}

// Connects to the remote service hosting job information, and
// the job's current result information.
func jobResult(id types.JobId) (*types.JobResult, *util.Error) {
	c := storage.NewClient()
	job, err := c.GetJob(id)
	if err != nil {
		return nil, &util.Error{
			Source: "jobResult",
			Info:   fmt.Sprintf("Failed to get job %d", id),
			Err:    err,
		}
	}

	result, err := job.Result()
	if err != nil {
		return nil, &util.Error{
			Source: "jobResult",
			Info:   fmt.Sprintf("Failed to get job %d result", id),
			Err:    err,
		}
	}

	return result, nil
}
