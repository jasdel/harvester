package queue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path"
)

const FirstLevel = 0

type ClientConfig struct {
	WebClient *http.Client
	Endpoint  string
}

type Client struct {
	endpoint  string
	webClient *http.Client
}

// Creates a new instance of the queue client. With a passed in
// configuration.  Only a single client is needed per process,
// and is save across go routines.
func NewClient(cfg *ClientConfig) *Client {
	// Verify th queue host is a valid URL now so it doesn't need to be
	// checked later.
	if _, err := url.Parse(cfg.Endpoint); err != nil {
		panic(fmt.Sprintf("queue.NewClient, invalid Endpoint URL", err))
	}

	return &Client{
		webClient: cfg.WebClient,
		endpoint:  cfg.Endpoint,
	}
}

// Enqueues a set of URLs with the queue service.
func (c *Client) Enqueue(urls []string, level int) error {
	if len(urls) == 0 {
		return fmt.Errorf("No URLs to enqueue")
	}

	respBody, err := c.post("enqueue", urls)
	if err != nil {
		return err
	}

	log.Println("Queue Client Enqueue: DEBUG:", string(respBody))

	return nil
}

// Submits a post HTTP request to the configured endpoint including the operation
// and content payload. The response body is returned in bytes.
func (c *Client) post(operation string, content interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	e := json.NewEncoder(buf)
	if err := e.Encode(content); err != nil {
		return nil, err
	}

	u, _ := url.Parse(c.endpoint)
	u.Path = path.Join(u.Path, operation)

	req, err := http.NewRequest("POST", u.String(), buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.webClient.Do(req)
	if err != nil {
		return nil, err
	}

	buf.Reset()
	if _, err := buf.ReadFrom(resp.Body); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
