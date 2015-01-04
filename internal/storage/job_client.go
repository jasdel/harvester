package storage

import (
	"database/sql"
	"fmt"
	"github.com/jasdel/harvester/internal/common"
	"github.com/lib/pq"
)

// Provides a name spaced collection of Job based storage operations. JobClient
// does not hold non go-routine state, and is safe to share across multiples.
type JobClient struct {
	// Storage client already configured and connected to the storage provider
	client *Client
}

// Extracts a job from a QueryRow.  Nil for the job will be returned
// if the job does not exist.
func getJobFromRow(row *sql.Row) (*Job, error) {
	var (
		id        sql.NullInt64
		createdOn pq.NullTime
	)

	if err := row.Scan(&id, &createdOn); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if !id.Valid || !createdOn.Valid {
		return nil, fmt.Errorf("Invalid result for get job")
	}

	return &Job{
		Id:        common.JobId(id.Int64),
		CreatedOn: createdOn.Time,
	}, nil
}

// Extracts the Job URLs from a Query of rows.
func getJobURLFromRows(rows *sql.Rows) (jobURL JobURL, err error) {
	var (
		jobId       sql.NullInt64
		url         sql.NullString
		completedOn pq.NullTime
	)

	if err = rows.Scan(&jobId, &url, &completedOn); err != nil {
		return jobURL, err
	}

	if !jobId.Valid || !url.Valid {
		return jobURL, fmt.Errorf("Invalid result for job URLs")
	}

	jobURL = JobURL{
		JobId:       common.JobId(jobId.Int64),
		URL:         url.String,
		CompletedOn: completedOn.Time,
	}
	if completedOn.Valid {
		jobURL.Completed = true
	}

	return jobURL, nil
}

// Create a new job entry with its URLS, returning the job object.
func (j *JobClient) CreateJob(urls []string) (*Job, error) {
	const queryInsertJob = `INSERT INTO job DEFAULT VALUES RETURNING id,created_on`
	const queryInsertJobURLs = `INSERT INTO job_url (job_id, url) VALUES ($1, $2)`

	job, err := getJobFromRow(j.client.db.QueryRow(queryInsertJob))
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, fmt.Errorf("Failed to get created job")
	}

	job.URLs = make([]JobURL, 0, len(urls))
	for _, u := range urls {
		if _, err := j.client.db.Exec(queryInsertJobURLs, job.Id, u); err != nil {
			return nil, err
		}
		job.URLs = append(job.URLs, JobURL{JobId: job.Id, URL: u})
	}

	return job, nil
}

// Searches for a job, and returns it and its URLs if the job exist. Nil is return if
// the job does not exist
func (j *JobClient) GetJob(id common.JobId) (*Job, error) {
	const queryJob = `SELECT id,created_on FROM job WHERE id = $1`

	job, err := getJobFromRow(j.client.db.QueryRow(queryJob, id))
	if err != nil || job == nil {
		return nil, err
	}

	const queryJobURLs = `SELECT job_id,url,completed_on FROM job_url WHERE job_id = $1`
	rows, err := j.client.db.Query(queryJobURLs, job.Id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	job.URLs = []JobURL{}
	for rows.Next() {
		jobURL, err := getJobURLFromRows(rows)
		if err != nil {
			return nil, err
		}
		job.URLs = append(job.URLs, jobURL)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return job, err
}

// Queries the results images for each image URL
func (j *JobClient) Result(id common.JobId, mimeFilter string) (common.JobResults, error) {
	const queryJobResult = `SELECT refer,url,mime FROM job_result WHERE job_id = $1 AND mime like $2`

	rows, err := j.client.db.Query(queryJobResult, id, mimeFilter+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(common.JobResults)
	knownResults := make(map[string]map[string]struct{})
	for rows.Next() {
		var refer sql.NullString
		var u sql.NullString
		var mime sql.NullString
		if err := rows.Scan(&refer, &u, &mime); err != nil {
			return nil, err
		}
		if !refer.Valid || !u.Valid {
			// Invalid mimes are ignored, because they might be null, if the URL
			// wasn't crawled deeper.
			return nil, fmt.Errorf("Invalid job result for job id %d", id)
		}

		if _, ok := result[refer.String]; !ok {
			result[refer.String] = []string{}
			knownResults[refer.String] = make(map[string]struct{})
		} else {
			if _, ok := knownResults[refer.String][u.String]; ok {
				// Prevent duplicate entries
				continue
			}
		}
		knownResults[refer.String][u.String] = struct{}{}

		result[refer.String] = append(result[refer.String], u.String)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
