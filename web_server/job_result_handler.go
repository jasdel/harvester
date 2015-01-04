package main

import (
	"fmt"
	"github.com/jasdel/harvester/internal/common"
	"github.com/jasdel/harvester/internal/storage"
	"log"
	"net/http"
	"path"
)

// Handles the request checking on the status of a previously scheduled job.
// Returns an error if the job isn't found, or invalid input. If the job
// exists its status will be returned. A result mime content type filter can
// also be provided as the 'mime' query parameter. The parameter acts as a prefix
// filter when returning results of a job
//
// e.g:
// curl -X GET "http://localhost:8080/results/1234"
//
// Response:
//	- Success: {<domain>: [ {Mime: <mime>, URL: <urls>} ]}
//	- Failure: {code: <code>, message: <message>}
type JobResultHandler struct {
	sc *storage.Client
}

func (h *JobResultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	mimeFilter := r.URL.Query().Get("mime")

	result, jobErr := h.jobResult(id, mimeFilter)
	if jobErr != nil {
		log.Println("routeJobResult request job result failed.", jobErr)
		writeJSONError(w, "NotFound", jobErr.Short(), http.StatusNotFound)
		return
	}

	// Write job status out
	writeJSON(w, result, http.StatusOK)
}

// Connects to the remote service hosting job information, and
// the job's current result information. Filter selects specific
// mime types of job results. A filter of "" will return all results.
// the filter acts as the prefix to a mime content type patter.
//
// e.g: mimeFilter := "image" // returns all image URLs
func (h *JobResultHandler) jobResult(id common.JobId, mimeFilter string) (common.JobResults, *ErroMsg) {
	result, err := h.sc.JobClient().Result(id, mimeFilter)
	if err != nil {
		return nil, &ErroMsg{
			Source: "jobResult",
			Info:   fmt.Sprintf("Failed to get job %d result", id),
			Err:    err,
		}
	}

	return result, nil
}
