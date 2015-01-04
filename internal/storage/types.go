package storage

import (
	"github.com/jasdel/harvester/internal/types"
	"time"
)

type URL struct {
	Id        int64
	URL       string
	Refer     string
	Crawled   bool
	CreatedOn time.Time
}

type Job struct {
	Id        types.JobId
	URLs      []JobURL
	createdOn time.Time
}

type JobURL struct {
	URL         string
	CompletedOn time.Time
	JobId       types.JobId
}

type PendingURLCrawl struct {
	Origin string
	URL    string
}
