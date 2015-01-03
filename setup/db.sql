-- Job
CREATE TABLE IF NOT EXISTS job (
    id           serial                   PRIMARY KEY,
    created_on   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Origin URLs from a job
CREATE TABLE IF NOT EXISTS job_url (
    id           serial PRIMARY KEY,
    job_id       INT    NOT NULL,
    url          TEXT   NOT NULL,
    completed_on TIMESTAMP WITH TIME ZONE,
    FOREIGN key (job_id) REFERENCES job(id)
);

-- Results for each job
CREATE TABLE IF NOT EXISTS job_result (
    id      serial PRIMARY KEY,
    job_id  INT    NOT NULL,
    refer   TEXT,
    url     TEXT   NOT NULL,
    mime    TEXT 
);

-- Collection of URLs encountered
CREATE TABLE IF NOT EXISTS url (
    id         serial PRIMARY KEY,
    mime       TEXT, -- content type this URL references
    url        TEXT   NOT NULL,
    refer      TEXT   NOT NULL, --origin URL encountered for this URL record
    crawled    BOOL   NOT NULL DEFAULT FALSE,
    created_on TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX url_pair ON url (url, refer);

-- job URL still pending
CREATE TABLE IF NOT EXISTS url_pending (
	id     serial PRIMARY KEY,
	origin TEXT   NOT NULL,
	url    TEXT   NOT NULL
);
CREATE UNIQUE INDEX url_pending_pair ON url_pending (url, origin);