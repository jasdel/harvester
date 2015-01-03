package storage

import (
	"database/sql"
	"fmt"
	"github.com/jasdel/harvester/internal/types"
	"github.com/lib/pq"
	"time"
)

type Job struct {
	id     types.JobId
	client *Client
}

func (j *Job) Id() types.JobId {
	return j.id
}

const queryJob = `SELECT created_on FROM job WHERE id = $1`
const queryJobURLStatus = `SELECT completed_on FROM job_url WHERE job_id = $1`

func (j *Job) Status() (*types.JobStatus, error) {
	var jobStartedOn pq.NullTime
	if err := j.client.db.QueryRow(queryJob, j.id).Scan(&jobStartedOn); err != nil {
		return nil, err
	}
	if !jobStartedOn.Valid {
		return nil, fmt.Errorf("Job not found %d", j.id)
	}

	rows, err := j.client.db.Query(queryJobURLStatus, j.id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	status := &types.JobStatus{Id: j.id}
	var compTime time.Time
	for rows.Next() {
		var completedOn pq.NullTime
		if err := rows.Scan(&completedOn); err != nil {
			return nil, err
		}
		if completedOn.Valid {
			status.Completed++
			if compTime.Before(completedOn.Time) {
				compTime = completedOn.Time
			}
		} else {
			status.Pending++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if status.Pending != 0 {
		compTime = time.Now().UTC()
	}
	status.Elapsed = compTime.Sub(jobStartedOn.Time)

	return status, nil
}

const queryJobResult = `SELECT origin,url,mime FROM job_result WHERE job_id = $1 AND mime like $2`

// Queries the results images for each image URL
func (j *Job) Result(mimeFilter string) (types.JobResults, error) {
	rows, err := j.client.db.Query(queryJobResult, j.id, mimeFilter+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(types.JobResults)

	for rows.Next() {
		var origin sql.NullString
		var u sql.NullString
		var mime sql.NullString
		if err := rows.Scan(&origin, &u, &mime); err != nil {
			return nil, err
		}
		if !origin.Valid || !u.Valid {
			// Invalid mimes are ignored, because they might be null, if the URL
			// wasn't crawled deeper.
			return nil, fmt.Errorf("Invalid job result for job id %d", j.id)
		}

		if _, ok := result[origin.String]; !ok {
			result[origin.String] = []types.JobResult{}
		}
		result[origin.String] = append(result[origin.String], types.JobResult{Mime: mime.String, URL: u.String})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}
