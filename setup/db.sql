-- Job
CREATE TABLE IF NOT EXISTS job (
    id           serial                   PRIMARY KEY,
    created_on   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Origin URLs from a job
CREATE TABLE IF NOT EXISTS job_url (
    job_id       INT    NOT NULL,          -- Job this URL belongs to
    url          TEXT   NOT NULL,          -- URL to be crawled for this job
    completed_on TIMESTAMP WITH TIME ZONE, -- The time stamp the crawl was completed
);

-- Results for each job
CREATE TABLE IF NOT EXISTS job_result (
    job_id  INT    NOT NULL,
    refer   TEXT   NOT NULL, -- URL which this job result was found from
    url     TEXT   NOT NULL, -- URL for this result
    mime    TEXT             -- content type this URL result references
);

-- Collection of URLs encountered
CREATE TABLE IF NOT EXISTS url (
    id         serial PRIMARY KEY,
    mime       TEXT,                          -- content type this URL references
    url        TEXT   NOT NULL,               -- URL of the content
    refer      TEXT   NOT NULL DEFAULT '',    -- Where the URL was encountered from
    crawled    BOOL   NOT NULL DEFAULT FALSE, -- If the URL has been crawled
    created_on TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- job URL still pending
CREATE TABLE IF NOT EXISTS url_pending (
	origin TEXT   NOT NULL, -- The Job URL that this URL is a descendant of 
	url    TEXT   NOT NULL  -- URL that is pending being crawled.
);
