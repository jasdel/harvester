package main

import (
	"bufio"
	"fmt"
	"github.com/apcera/nats"
	"github.com/jasdel/harvester/internal/queue"
	"github.com/jasdel/harvester/internal/storage"
	"github.com/jasdel/harvester/internal/types"
	"github.com/jasdel/harvester/internal/util"
	"github.com/zenazn/goji/web"
	"io"
	"log"
	"net/http"
	"net/url"
)

var queuePub queue.Publisher

func init() {
	var err error
	queuePub, err = queue.NewPublisher(nats.DefaultURL, "url_queue")
	if err != nil {
		panic(err)
	}
}

type jobScheduledMsg struct {
	JobId types.JobId `json:"jobId"`
}

// Handles the request to schedule a new job. Expects a new line separated
// list of URLs as input in the request's body. Will respond back with error
// message, or job id if the schedule was successful.
//
// e.g:
// curl -X POST --data-binary @- "http://localhost:8000" << EOF
// https://www.google.com
// http://example.com
// EOF
//
// Response:
//	- Success: {jobId: 1234}
//	- Failure: {code: <code>, message: <message>}
func routeScheduleJob(c web.C, w http.ResponseWriter, r *http.Request) {
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
	id, err := scheduleJob(urls)
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

	urls := []string{}
	for scanner.Scan() {
		if scanner.Text() == "" {
			continue
		}
		u, err := url.Parse(scanner.Text())
		if err != nil || u.Host == "" {
			return nil, &util.Error{
				Source: "getRequestedJobURLs",
				Info:   fmt.Sprintf("Invalid URL: %s", scanner.Text()),
				Err:    err,
			}
		}
		urls = append(urls, u.String())
	}
	if err := scanner.Err(); err != nil {
		return nil, &util.Error{
			Source: "getRequestedJobURLs",
			Info:   "Unexpected error in input",
			Err:    err,
		}
	}

	return urls, nil
}

// Requests that a job be created, and the parts of it be scheduled.
// a job id will be returned if the job was successfully created, and
// error if there was a failure.
func scheduleJob(urls []string) (types.JobId, *util.Error) {
	c := storage.NewClient()
	job, err := c.CreateJob(urls)
	if err != nil {
		return types.InvalidJobId, &util.Error{
			Source: "scheduleJob",
			Info:   fmt.Sprintf("Create Job Failed"),
			Err:    err,
		}
	}

	go func() {
		for _, u := range urls {
			queuePub.Send(&types.URLQueueItem{Origin: u, URL: u})
		}
	}()

	return job.Id(), nil
}
