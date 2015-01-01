package storage

import (
	"github.com/jasdel/harvester/internal/types"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) CreateJob(urls []string) (Job, error) {
	return &job{
		id:     0,
		client: c,
	}, nil
}

func (c *Client) DeleteJob(id types.JobId) error {
	return nil
}

func (c *Client) GetJob(id types.JobId) (Job, error) {
	return &job{
		id:     id,
		client: c,
	}, nil
}
