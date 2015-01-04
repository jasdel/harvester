package storage

import (
	"database/sql"
	"fmt"
	"github.com/jasdel/harvester/internal/types"
	"github.com/lib/pq"
	"time"
)

type URLClient struct {
	client *Client
}

const DefaultURLMime = ``

// Returns a URL record if it exists for the URL+refer pair
func (u *URLClient) GetURLWithRefer(url, refer string) (*URL, error) {
	const queryURLWithRefer = `SELECT id,url,refer,mime,crawled,created_on FROM url WHERE url = $1 AND refer = $2`

	return getURLFromRow(u.client.db.QueryRow(queryURLWithRefer, url, refer))
}

// Returns a list of direct descendants of the passed in URL.  The passed in URL
// will be the 'refer' value for each of the returned URLs, if there are any.
func (u *URLClient) GetAllURLsWithRefer(refer string) ([]*URL, error) {
	const queryAllURLsWithRefer = `SELECT id,url,refer,mime,crawled,created_on FROM url WHERE refer = $1`

	rows, err := u.client.db.Query(queryAllURLsWithRefer, refer)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := []*URL{}
	for rows.Next() {
		url, err := getURLFromRows(rows)
		if err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return urls, nil
}

// Adds a URL to the database for a specific URL/refer combination.
// mime is the content-type of the url
func (u *URLClient) Add(url, refer, mime string) error {
	const queryURLAdd = `INSERT INTO url (url, refer, mime) VALUES ($1, $2, $3)`

	if _, err := u.client.db.Exec(queryURLAdd, url, refer, mime); err != nil {
		return err
	}
	return nil
}

// Updates the mime content-type of a preexisting URL.
func (u *URLClient) Update(url, mime string, crawled bool) error {
	const queryURLUpdateMime = `UPDATE url SET mime = $2, crawled = $3 WHERE url = $1`

	if _, err := u.client.db.Exec(queryURLUpdateMime, url, mime, crawled); err != nil {
		return err
	}
	return nil
}

// Adds the URL as pending under a origin
func (u *URLClient) AddPending(url, origin string) error {
	const queryURLAddPending = `INSERT INTO url_pending (url,origin) VALUES ($1, $2)`

	if _, err := u.client.db.Exec(queryURLAddPending, url, origin); err != nil {
		return err
	}
	return nil
}

// Deletes an existing pending record for a URL that no longer will be crawled
func (u *URLClient) DeletePending(url, origin string) error {
	const queryURLDeletePending = `DELETE FROM url_pending WHERE url = $1 AND origin = $2`

	if _, err := u.client.db.Exec(queryURLDeletePending, url, origin); err != nil {
		return err
	}
	return nil
}

// Records a new crawled URL result for a job
func (u *URLClient) AddResult(jobId types.JobId, url, refer, mime string) error {
	const queryURLInsertResult = `INSERT INTO job_result (url, job_id, refer, mime) VALUES ($1, $2, $3, $4)`

	if _, err := u.client.db.Exec(queryURLInsertResult, url, jobId, refer, mime); err != nil {
		return err
	}
	return nil
}

// Adds a batch of URLs to the result
func (u *URLClient) AddURLsToResults(jobIds []types.JobId, refer string, urls []*URL) error {
	for _, jobId := range jobIds {
		for _, url := range urls {
			if err := u.AddResult(jobId, url.URL, refer, url.Mime); err != nil {
				return err
			}
		}
	}
	return nil
}

// Returns true if there are any pending entries for the origin URL provided. The origin
// field is used for this search.
func (u *URLClient) HasPending(url string) (bool, error) {
	const queryURLHasPending = `SELECT exists(SELECT 1 FROM url_pending WHERE origin = $1)`

	var pending sql.NullBool
	if err := u.client.db.QueryRow(queryURLHasPending, url).Scan(&pending); err != nil {
		return false, err
	}

	return pending.Valid && pending.Bool, nil
}

// Marks a pre-existing job's URL as completed. This means that all descendants have been
// crawled up to the max level.
func (u *URLClient) MarkJobURLComplete(url string) error {
	const queryURLJobURComplete = `UPDATE job_url SET completed_on = $1 WHERE url = $2 AND completed_on IS NULL`
	curTime := time.Now().UTC()
	if _, err := u.client.db.Exec(queryURLJobURComplete, curTime, url); err != nil {
		return err
	}
	return nil
}

// Searches for all JobIds associated with this origin URL.
func (u *URLClient) GetJobIdsForURL(url string) ([]types.JobId, error) {
	const queryURLJobURLOrigin = `SELECT job_id FROM job_url WHERE url = $1 AND completed_on IS NULL`
	rows, err := u.client.db.Query(queryURLJobURLOrigin, url)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobIds := []types.JobId{}
	for rows.Next() {
		var id sql.NullInt64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if !id.Valid {
			return nil, fmt.Errorf("No job id for url", url)
		}
		jobIds = append(jobIds, types.JobId(id.Int64))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobIds, nil
}

// Checks if a URL has any pending entries in the job URL pending table.
// If there are no longer any entries, All URLs associated with a jobs, not yet completed
// will be marked as completed.
func (u *URLClient) UpdateJobURLIfComplete(url string) (bool, error) {
	if pending, _ := u.HasPending(url); !pending {
		if err := u.MarkJobURLComplete(url); err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
	return false, nil
}

// Extracts the URL from a QueryRow row.
func getURLFromRow(row *sql.Row) (*URL, error) {
	var (
		id        sql.NullInt64
		url       sql.NullString
		refer     sql.NullString
		mime      sql.NullString
		crawled   sql.NullBool
		createdOn pq.NullTime
	)

	if err := row.Scan(&id, &url, &refer, &mime, &crawled, &createdOn); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if !id.Valid || !url.Valid {
		return nil, fmt.Errorf("Invalid URL result from QueryRow scan")
	}

	return &URL{
		Id:        id.Int64,
		URL:       url.String,
		Refer:     refer.String,
		Mime:      mime.String,
		Crawled:   crawled.Bool,
		CreatedOn: createdOn.Time,
	}, nil
}

// Extracts the URL fields from a Query rows.
func getURLFromRows(rows *sql.Rows) (*URL, error) {
	var (
		id        sql.NullInt64
		url       sql.NullString
		refer     sql.NullString
		mime      sql.NullString
		crawled   sql.NullBool
		createdOn pq.NullTime
	)

	if err := rows.Scan(&id, &url, &refer, &mime, &crawled, &createdOn); err != nil {
		return nil, err
	}

	if !id.Valid || !url.Valid {
		return nil, fmt.Errorf("Invalid URL result from QueryRow scan")
	}

	return &URL{
		Id:        id.Int64,
		URL:       url.String,
		Refer:     refer.String,
		Mime:      mime.String,
		Crawled:   crawled.Bool,
		CreatedOn: createdOn.Time,
	}, nil
}
