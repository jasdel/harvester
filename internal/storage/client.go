package storage

import (
	"database/sql"
	"fmt"
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

// Return an Job which can be used to perform queries and manipulation
// of job data stored in storage.
func (c *Client) JobClient() *JobClient {
	return &JobClient{
		client: c,
	}
}

// Return an URL which can be used to perform queries and manipulation
// of URL data stored in storage.
func (c *Client) URLClient() *URLClient {
	return &URLClient{
		client: c,
	}
}
