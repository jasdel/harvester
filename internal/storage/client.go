package storage

import (
	"database/sql"
	"fmt"
	"github.com/jasdel/harvester/internal/types"
	_ "github.com/lib/pq"
)

type Client struct {
	db *sql.DB
}

func NewClient() *Client {
	connInfo := `user=docker password=docker dbname=docker host=localhost port=24001 sslmode=disable`

	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}
	return &Client{
		db: db,
	}
}

const queryInsertJob = `INSERT INTO job DEFAULT VALUES RETURNING id`
const queryInsertJobURLs = `INSERT INTO job_url (job_id, url) VALUES ($1, $2)`
const queryDeleteJob = `DELETE FROM job WHERE id = $1`

// Create a new job entry with its URLS, returning the job object
func (c *Client) CreateJob(urls []string) (*Job, error) {
	var id sql.NullInt64
	if err := c.db.QueryRow(queryInsertJob).Scan(&id); err != nil {
		return nil, err
	}
	if !id.Valid {
		return nil, fmt.Errorf("No jobId created")
	}

	// TODO do this in a single statement
	for _, u := range urls {
		if _, err := c.db.Exec(queryInsertJobURLs, id.Int64, u); err != nil {
			return nil, err
		}
	}

	return &Job{
		id:     types.JobId(id.Int64),
		client: c,
	}, nil
}

func (c *Client) ForJob(id types.JobId) *Job {
	return &Job{
		id:     id,
		client: c,
	}
}

func (c *Client) ForURL(url string) *URL {
	return &URL{
		url:    url,
		client: c,
	}
}
