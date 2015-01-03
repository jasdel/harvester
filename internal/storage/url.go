package storage

import (
	"database/sql"
	"fmt"
	"github.com/jasdel/harvester/internal/types"
	"time"
)

type URL struct {
	url    string
	client *Client
}

const DefaultURLMime = ``

const queryURLKnown = `SELECT exists(SELECT 1 FROM url WHERE url = $1)`

func (u *URL) Known() (bool, error) {
	var known sql.NullBool
	if err := u.client.db.QueryRow(queryURLKnown, u.url).Scan(&known); err != nil {
		return false, err
	}

	return known.Valid && known.Bool, nil
}

const queryURLKnownWRefer = `SELECT exists(SELECT 1 FROM url WHERE url = $1 AND refer = $2)`

func (u *URL) KnownWithRefer(refer string) (bool, error) {
	var known sql.NullBool
	if err := u.client.db.QueryRow(queryURLKnownWRefer, u.url, refer).Scan(&known); err != nil {
		return false, err
	}

	return known.Valid && known.Bool, nil
}

const queryURLCrawled = `SELECT exists(SELECT 1 FROM url WHERE url = $1 AND crawled = TRUE)`

func (u *URL) Crawled() (bool, error) {
	var crawled sql.NullBool
	if err := u.client.db.QueryRow(queryURLCrawled, u.url).Scan(&crawled); err != nil {
		return false, err
	}

	return crawled.Valid && crawled.Bool, nil
}

const queryURLAdd = `INSERT INTO url (url, refer, mime) VALUES ($1, $2, $3)`

// Adds a URL to the database for a specific URL/refer combination.
// mime is the content-type of the url
func (u *URL) Add(refer, mime string) error {
	if _, err := u.client.db.Exec(queryURLAdd, u.url, refer, mime); err != nil {
		return err
	}
	return nil
}

const queryURLUpdateMime = `UPDATE url SET mime = $2, crawled = $3 WHERE url = $1`

// Updates the mime content-type of a preexisting URL.
func (u *URL) Update(mime string, crawled bool) error {
	if _, err := u.client.db.Exec(queryURLUpdateMime, u.url, mime, crawled); err != nil {
		return err
	}
	return nil
}

const queryURLAddPending = `INSERT INTO url_pending (url,origin) VALUES ($1, $2)`

func (u *URL) AddPending(origin string) error {
	fmt.Println("Inserting into pending", u.url, origin)
	if _, err := u.client.db.Exec(queryURLAddPending, u.url, origin); err != nil {
		return err
	}
	return nil
}

const queryURLDeletePending = `DELETE FROM url_pending WHERE url = $1 AND origin = $2`

func (u *URL) DeletePending(origin string) error {
	if _, err := u.client.db.Exec(queryURLDeletePending, u.url, origin); err != nil {
		return err
	}
	return nil
}

const queryURLInsertResult = `INSERT INTO job_result (url, job_id, origin, refer, mime) VALUES ($1, $2, $3, $4, $5)`

func (u *URL) AddResult(jobId types.JobId, origin, refer, mime string) error {
	if _, err := u.client.db.Exec(queryURLInsertResult, u.url, jobId, origin, refer, mime); err != nil {
		return err
	}
	return nil
}

const queryURLHasPending = `SELECT exists(SELECT 1 FROM url_pending WHERE origin = $1)`

func (u *URL) HasPending() (bool, error) {
	var pending sql.NullBool
	if err := u.client.db.QueryRow(queryURLHasPending, u.url).Scan(&pending); err != nil {
		return false, err
	}

	return pending.Valid && pending.Bool, nil
}

const queryURLJobURComplete = `UPDATE job_url SET completed_on = $1 WHERE url = $2 AND completed_on IS NULL`

func (u *URL) MarkJobURLComplete() error {
	curTime := time.Now().UTC()
	if _, err := u.client.db.Exec(queryURLJobURComplete, curTime, u.url); err != nil {
		return err
	}
	return nil
}

const queryURLJobURLOrigin = `SELECT job_id FROM job_url WHERE url = $1 AND completed_on IS NULL`

func (u *URL) GetJobIds() ([]types.JobId, error) {
	rows, err := u.client.db.Query(queryURLJobURLOrigin, u.url)
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
			return nil, fmt.Errorf("No job id for url", u.url)
		}
		jobIds = append(jobIds, types.JobId(id.Int64))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobIds, nil
}
