package storage

import (
	"database/sql"
	"fmt"
	"github.com/jasdel/harvester/internal/types"
	_ "github.com/lib/pq"
)

// Configuration for the database storage.
type ClientConfig struct {
	User    string `json:"user"`
	Pass    string `json:"pass"`
	DBName  string `json:"dbname"`
	Host    string `json:"host"`
	Port    int    `json:"port"`
	SSLMode bool   `json:"sslmode"`
}

// Converts the configuration into a string for the sql.Open's connInfo parameter
func (c ClientConfig) String() string {
	sslMode := "disable"
	if c.SSLMode {
		sslMode = "enable"
	}

	return fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s",
		c.User, c.Pass, c.DBName, c.Host, c.Port, sslMode)
}

// Client for communicating with the storage service. Provides a way to
// Create jobs, update jobs, and manipulate URL entries
type Client struct {
	db *sql.DB
}

// Creates a new instance of the storage client. returning a client instance
// to perform operations with. The client is safe across multiple go routines.
func NewClient(cfg ClientConfig) (*Client, error) {
	db, err := sql.Open("postgres", cfg.String())
	if err != nil {
		return nil, err
	}
	return &Client{
		db: db,
	}, nil
}

const queryInsertJob = `INSERT INTO job DEFAULT VALUES RETURNING id`
const queryInsertJobURLs = `INSERT INTO job_url (job_id, url) VALUES ($1, $2)`

// Create a new job entry with its URLS, returning the job object
func (c *Client) CreateJob(urls []string) (*JobClient, error) {
	var id sql.NullInt64
	if err := c.db.QueryRow(queryInsertJob).Scan(&id); err != nil {
		return nil, err
	}
	if !id.Valid {
		return nil, fmt.Errorf("No jobId created")
	}

	// TODO This should be able to be done in a single statement
	for _, u := range urls {
		if _, err := c.db.Exec(queryInsertJobURLs, id.Int64, u); err != nil {
			return nil, err
		}
	}

	return c.ForJob(types.JobId(id.Int64)), nil
}

// Return an Job which can be used to perform queries and manipulation
// of job data stored in storage.
func (c *Client) ForJob(id types.JobId) *JobClient {
	return &JobClient{
		id:     id,
		client: c,
	}
}

// Return an URL which can be used to perform queries and manipulation
// of URL data stored in storage.
func (c *Client) ForURL(url string) *URLClient {
	return &URLClient{
		url:    url,
		client: c,
	}
}

func (c *Client) GetURL(url string) *URL {

}
