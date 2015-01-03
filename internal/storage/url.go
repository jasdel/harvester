package storage

import (
	"database/sql"
	"fmt"
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
