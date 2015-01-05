package storage

import (
	"database/sql"
	"fmt"
	"github.com/jasdel/harvester/internal/common"
	"github.com/lib/pq"
	"time"
)

// Provides a name spaced collection of URL based storage operations. JURLClient
// does not hold non go-routine state, and is safe to share across multiples.
type URLClient struct {
	// Storage client already configured and connected to the storage provider
	client *Client
}

// Requests a URL record by Id.
// If no URL is found, nil will be returned for the URL
func (u *URLClient) GetURLById(id common.URLId) (*URL, error) {
	const queryURLById = `SELECT id,url,mime,crawled_on FROM url WHERE id = $1`
	return getURLFromRow(u.client.db.QueryRow(queryURLById, id))

}

// Requests a URL record for the URL by URL string value.
// If no URL is found, nil will be returned for the URL
func (u *URLClient) GetURLByURL(url string) (*URL, error) {
	const queryURLByName = `SELECT id,url,mime,crawled_on FROM url WHERE url = $1`
	return getURLFromRow(u.client.db.QueryRow(queryURLByName, url))
}

// Attempts to get a URL if it already exists. If the URL does not
// exist a new entry will be added, and that URL entry will be returned.
// The 'mime' value will only be used if the URL needs to be added.
func (u *URLClient) GetOrAddURLByURL(urlStr, mime string) (*URL, error) {
	url, err := u.GetURLByURL(urlStr)
	if err != nil {
		return nil, err
	}
	if url == nil {
		var err error
		if url, err = u.Add(urlStr, mime); err != nil {
			return nil, err
		}
	}

	return url, nil
}

// Returns a list of direct descendants of the passed in URL.  The passed in URL
// will be the 'refer' value for each of the returned URLs, if there are any.
func (u *URLClient) GetAllURLsWithReferById(referId common.URLId) ([]*URL, error) {
	const queryAllURLsWithRefer = `
SELECT url.id, url.url, url.mime, url.crawled_on
FROM url_link
LEFT JOIN url on url_link.url_id = url.id
WHERE url_link.refer_id = $1`

	rows, err := u.client.db.Query(queryAllURLsWithRefer, referId)
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

// Adds a new URL to the database returning a URL object for it.
// If no mime is known us common.DefaultMime in its place.
func (u *URLClient) Add(url, mime string) (*URL, error) {
	// magic
	const queryURLAdd = `
WITH s AS (
    SELECT id, url
    FROM url
    WHERE url = $1
), i as (
    INSERT INTO url (url, mime)
    SELECT $1, $2
    WHERE NOT EXISTS (SELECT 1 FROM s)
    RETURNING id
)
SELECT id
from i
union all
select id
from s`

	var id sql.NullInt64
	if err := u.client.db.QueryRow(queryURLAdd, url, mime).Scan(&id); err != nil {
		return nil, err
	}
	if !id.Valid {
		return nil, fmt.Errorf("Insert failed no URL id created")
	}

	return &URL{
		Id:   common.URLId(id.Int64),
		Mime: mime,
	}, nil
}

// Attempts to insert a link between a refer and URL into the storage. If the
// link already exists, the insert statement will be ignored.
func (u *URLClient) AddLink(urlId, referId common.URLId) error {
	const queryURLInsertLink = `
INSERT INTO url_link (url_id, refer_id)
	SELECT $1, $2
	WHERE NOT EXISTS (SELECT 1 FROM url_link WHERE url_id = $1 AND refer_id = $2)`

	if _, err := u.client.db.Exec(queryURLInsertLink, urlId, referId); err != nil {
		return err
	}
	return nil
}

// Updates the mime content-type of a preexisting URL.
func (u *URLClient) MarkCrawled(urlId common.URLId, mime string) error {
	const queryURLUpdateMime = `UPDATE url SET mime = $2, crawled_on = $3 WHERE id = $1`

	crawledOn := time.Now().UTC()
	if _, err := u.client.db.Exec(queryURLUpdateMime, urlId, mime, crawledOn); err != nil {
		return err
	}
	return nil
}

// Adds the URL as pending under a origin URL. If the record already exists the
// insert statement will be ignored.
func (u *URLClient) AddPending(urlId, originId common.URLId) error {
	const queryURLAddPending = `
INSERT INTO url_pending (url_id,origin_id)
	SELECT $1, $2
	WHERE NOT EXISTS (SELECT 1 FROM url_pending WHERE url_id = $1 AND origin_id = $2)`

	if _, err := u.client.db.Exec(queryURLAddPending, urlId, originId); err != nil {
		return err
	}
	return nil
}

// Deletes a pending record for a URL that no longer needs be crawled. The pending
// record is a combination of url + origin, where origin is the origin URL the Job was
// created with.
func (u *URLClient) DeletePending(urlId, originId common.URLId) error {
	const queryURLDeletePending = `DELETE FROM url_pending WHERE url_id = $1 AND origin_id = $2`

	if _, err := u.client.db.Exec(queryURLDeletePending, urlId, originId); err != nil {
		return err
	}
	return nil
}

// Returns true if there are any pending entries for the origin URL provided. The origin
// field is used for this search.
func (u *URLClient) HasPending(originId common.URLId) (bool, error) {
	const queryURLHasPending = `SELECT exists(SELECT 1 FROM url_pending WHERE origin_id = $1)`

	var pending sql.NullBool
	if err := u.client.db.QueryRow(queryURLHasPending, originId).Scan(&pending); err != nil {
		return false, err
	}

	return pending.Valid && pending.Bool, nil
}

// Records a new crawled URL into the job results, for a specific jobId. If the result record
// already exists, the insert statement will be ignored.
func (u *URLClient) AddResult(jobId common.JobId, urlId, referId common.URLId, mime string) error {
	const queryURLInsertResult = `
INSERT INTO job_result (job_id, url_id, refer_id, mime)
	SELECT $1, $2, $3, $4
	WHERE NOT EXISTS (SELECT 1 FROM job_result WHERE job_id = $1 AND url_id = $2 AND refer_id = $3)`

	if _, err := u.client.db.Exec(queryURLInsertResult, jobId, urlId, referId, mime); err != nil {
		return err
	}
	return nil
}

// Adds a batch of URLs to the job results. Will update the job result for each job Id provided
func (u *URLClient) AddURLsToResults(jobIds []common.JobId, referId common.URLId, urls []*URL) error {
	for _, jobId := range jobIds {
		for _, url := range urls {
			if err := u.AddResult(jobId, url.Id, referId, url.Mime); err != nil {
				return err
			}
		}
	}
	return nil
}

// Marks a pre-existing job's URL as completed. This means that all descendants have been
// crawled up to the max level.
func (u *URLClient) MarkJobURLComplete(urlId common.URLId) error {
	const queryURLJobURComplete = `UPDATE job_url SET completed_on = $1 WHERE url_id = $2 AND completed_on IS NULL`
	curTime := time.Now().UTC()
	if _, err := u.client.db.Exec(queryURLJobURComplete, curTime, urlId); err != nil {
		return err
	}
	return nil
}

// Searches for all JobIds associated with this origin URL.
func (u *URLClient) GetJobIdsForURLById(urlId common.URLId) ([]common.JobId, error) {
	const queryURLJobURLOrigin = `SELECT job_id FROM job_url WHERE url_id = $1 AND completed_on IS NULL`
	rows, err := u.client.db.Query(queryURLJobURLOrigin, urlId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	jobIds := []common.JobId{}
	for rows.Next() {
		var id sql.NullInt64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if !id.Valid {
			return nil, fmt.Errorf("No job id for URL", urlId)
		}
		jobIds = append(jobIds, common.JobId(id.Int64))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return jobIds, nil
}

// Checks if a URL has any pending entries in the job URL pending table.
// If there are no longer any entries, All URLs associated with a jobs, not yet completed
// will be marked as completed.
func (u *URLClient) UpdateJobURLIfComplete(urlId common.URLId) (bool, error) {
	if pending, _ := u.HasPending(urlId); !pending {
		if err := u.MarkJobURLComplete(urlId); err != nil {
			return false, err
		} else {
			return true, nil
		}
	}
	return false, nil
}

// Extracts the URL from a QueryRow row. If no URL is found, nil will be returned for the URL
// Expects the query columns to be the following order:
//		id, url, mime, crawled_on
func getURLFromRow(row *sql.Row) (*URL, error) {
	var (
		id        sql.NullInt64
		url       sql.NullString
		mime      sql.NullString
		crawledOn pq.NullTime
	)

	if err := row.Scan(&id, &url, &mime, &crawledOn); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if !id.Valid || !url.Valid {
		return nil, fmt.Errorf("Invalid URL result from QueryRow scan")
	}

	return &URL{
		Id:        common.URLId(id.Int64),
		URL:       url.String,
		Mime:      mime.String,
		Crawled:   crawledOn.Valid,
		CrawledOn: crawledOn.Time,
	}, nil
}

// Extracts the URL fields from a Query rows. If no URL is found, nil will be returned for the URL
// Expects the query columns to be the following order:
//		id, url, mime, crawled_on
func getURLFromRows(rows *sql.Rows) (*URL, error) {
	var (
		id        sql.NullInt64
		url       sql.NullString
		mime      sql.NullString
		crawledOn pq.NullTime
	)

	if err := rows.Scan(&id, &url, &mime, &crawledOn); err != nil {
		return nil, err
	}

	if !id.Valid || !url.Valid {
		return nil, fmt.Errorf("Invalid URL result from QueryRow scan")
	}

	return &URL{
		Id:        common.URLId(id.Int64),
		URL:       url.String,
		Mime:      mime.String,
		Crawled:   crawledOn.Valid,
		CrawledOn: crawledOn.Time,
	}, nil
}
