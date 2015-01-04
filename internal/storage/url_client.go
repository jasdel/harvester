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

	return &URL{
		Id:        id.Int64,
		URL:       url.String,
		Refer:     refer.String,
		Mime:      mime.String,
		Crawled:   crawled.Bool,
		CreatedOn: createdOn.Time,
	}, nil
}

// func (u *URLClient) Known() (bool, error) {
// 	const queryURLKnown = `SELECT exists(SELECT 1 FROM url WHERE url = $1)`

// 	var known sql.NullBool
// 	if err := u.client.db.QueryRow(queryURLKnown, u.url).Scan(&known); err != nil {
// 		return false, err
// 	}

// 	return known.Valid && known.Bool, nil
// }

// func (u *URLClient) KnownWithRefer(refer string) (bool, error) {
// 	const queryURLKnownWRefer = `SELECT exists(SELECT 1 FROM url WHERE url = $1 AND refer = $2)`
// 	var known sql.NullBool
// 	if err := u.client.db.QueryRow(queryURLKnownWRefer, u.url, refer).Scan(&known); err != nil {
// 		return false, err
// 	}

// 	return known.Valid && known.Bool, nil
// }

// func (u *URLClient) Crawled() (bool, error) {
// 	const queryURLCrawled = `SELECT exists(SELECT 1 FROM url WHERE url = $1 AND crawled = TRUE)`
// 	var crawled sql.NullBool
// 	if err := u.client.db.QueryRow(queryURLCrawled, u.url).Scan(&crawled); err != nil {
// 		return false, err
// 	}

// 	return crawled.Valid && crawled.Bool, nil
// }

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
