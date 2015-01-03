-- Job
CREATE TABLE IF NOT EXISTS job (
    id           serial                   PRIMARY KEY,
    created_on   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Origin URLs from a job
CREATE TABLE IF NOT EXISTS job_url (
    job_id       INT    NOT NULL,
    url          TEXT   NOT NULL,
    completed_on TIMESTAMP WITH TIME ZONE,
    FOREIGN key (job_id) REFERENCES job(id)
);

-- Results for each job
CREATE TABLE IF NOT EXISTS job_result (
    job_id  INT    NOT NULL,
    origin  TEXT   NOT NULL,
    url     TEXT   NOT NULL,
    mime    TEXT
);

-- Collection of URLs encountered
CREATE TABLE IF NOT EXISTS url (
    id         serial PRIMARY KEY,
    mime       TEXT,            -- content type this URL references
    url        TEXT   NOT NULL,
    refer      TEXT   NOT NULL, -- Where the URL was encountered from
    crawled    BOOL   NOT NULL DEFAULT FALSE,
    created_on TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- job URL still pending
CREATE TABLE IF NOT EXISTS url_pending (
	origin TEXT   NOT NULL,
	url    TEXT   NOT NULL
);
