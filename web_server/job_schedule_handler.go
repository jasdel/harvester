package main

import (
	"bufio"
	"fmt"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"github.com/jasdel/harvester/internal/util"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

type jobScheduledMsg struct {
	JobId types.JobId `json:"jobId"`
}

// Handles the request to schedule a new job. Expects a new line separated
// list of URLs as input in the request's body. Will respond back with error
// message, or job id if the schedule was successful.
//
// e.g:
// curl -X POST --data-binary @- "http://localhost:8080" << EOF
// https://www.google.com
// http://example.com
// EOF
//
// Response:
//	- Success: {jobId: 1234}
//	- Failure: {code: <code>, message: <message>}
type JobScheduleHandler struct {
	queuePub queue.Publisher
	sc       *storage.Client
}

func (h *JobScheduleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.Header().Set("Allow", "POST")
		http.Error(w, "MethodNotAllowed", http.StatusMethodNotAllowed)
		return
	}

	urls, err := getRequestedJobURLs(r.Body)
	if err != nil {
		log.Println("routeScheduleJob request parse failed", err)
		writeJSONError(w, "BadRequest", err.Short(), http.StatusBadRequest)
		return
	}

	if len(urls) == 0 {
		// Nothing can be done if there are no URLs to schedule
		log.Println("routeScheduleJob request has no URLs")
		writeJSONError(w, "BadRequest", "No URLs provided", http.StatusBadRequest)
		return
	}

	// Create job by sending the URLs to scheduler
	id, err := h.scheduleJob(urls)
	if err != nil {
		log.Println("routeScheduleJob request job schedule failed.", err)
		writeJSONError(w, "DependancyFailure", err.Short(), http.StatusInternalServerError)
		return
	}

	// Write job status out
	writeJSON(w, jobScheduledMsg{JobId: id}, http.StatusOK)
}

// Reads the input scanning for URLs. It expects a single URL per
// line. If there is a failure reading from the input, or a invalid
// URL is encountered an error will be returned.
func getRequestedJobURLs(in io.Reader) ([]string, *util.Error) {
	scanner := bufio.NewScanner(in)

	urlMap := make(map[string]struct{})
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}

		u, err := validateJobURL(scanner.Text())
		if err != nil {
			return nil, &util.Error{
				Source: "getRequestedJobURLs",
				Info:   fmt.Sprintf("Invalid URL: %s", scanner.Text()),
				Err:    err,
			}
		}
		urlMap[u] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, &util.Error{
			Source: "getRequestedJobURLs",
			Info:   "Unexpected error in input",
			Err:    err,
		}
	}

	return util.ArrayifyMap(urlMap), nil
}

// Validates the job URL contains at least a host and scheme. The scheme is also validated
// as being http or https. If no scheme is provided http will be used as the default.
func validateJobURL(jobURL string) (string, error) {
	if strings.HasPrefix(jobURL, "/") {
		return "", fmt.Errorf("Invalid URL, does not have host")
	}

	u, err := url.Parse(jobURL)
	if err != nil {
		return "", err
	}
	if u.Scheme != "" && u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("Invalid URL scheme")
	}
	if u.Scheme == "" {
		// set default scheme if non are provided, so the input could be www.example.com
		u.Scheme = "http"
	}
	return u.String(), nil
}

// Requests that a job be created, and the parts of it be scheduled.
// a job id will be returned if the job was successfully created, and
// error if there was a failure.
func (h *JobScheduleHandler) scheduleJob(urls []string) (types.JobId, *util.Error) {
	job, err := h.sc.JobClient().CreateJob(urls)
	if err != nil {
		return types.InvalidJobId, &util.Error{
			Source: "JobScheduleHandler.scheduleJob",
			Info:   fmt.Sprintf("Create Job Failed"),
			Err:    err,
		}
	}

	go func() {
		for _, u := range job.URLs {
			if err := h.sc.URLClient().AddPending(u.URL, u.URL); err != nil {
				log.Println("JobScheduleHandler.scheduleJob: failed to add job URL to pending list")
			}
			h.queuePub.Send(&types.URLQueueItem{Origin: u.URL, URL: u.URL})
		}
	}()

	return job.Id, nil
}
