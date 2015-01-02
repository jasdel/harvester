package storage

import (
	"database/sql"
	"fmt"
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

func (u *URL) AddPending(origin string) error {
	return nil
}

func (u *URL) DeletePending(origin string) error {
	return nil
}

const queryURLJobURLOrigin = `SELECT job_id FROM job_url WHERE url = $1 AND completed_on IS NULL`
const queryURLInsertResult = `INSERT INTO job_result (url, job_id, refer, mime) VALUES ($1, $2, $3, $4)`

func (u *URL) AddResult(origin, refer, mime string) error {
	var jobId sql.NullInt64

	if err := u.client.db.QueryRow(queryURLJobURLOrigin, origin).Scan(&jobId); err != nil {
		return err
	}
	if !jobId.Valid {
		return fmt.Errorf("URL origin not known as job, %s", origin)
	}

	if _, err := u.client.db.Exec(queryURLInsertResult, u.url, jobId.Int64, refer, mime); err != nil {
		return err
	}
	return nil
}
