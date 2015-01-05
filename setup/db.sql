
-- Collection of URLs encountered
CREATE TABLE IF NOT EXISTS url (
    id         serial PRIMARY KEY,
    mime       TEXT,                   -- content type this URL references
    url        TEXT   NOT NULL,        -- URL of the content
    crawled_on TIMESTAMP WITH TIME ZONE
);
CREATE UNIQUE INDEX url_unique ON url(url);

-- Scheduled Job
CREATE TABLE IF NOT EXISTS job (
    id           serial                   PRIMARY KEY,
    created_on   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Origin URLs from a job
CREATE TABLE IF NOT EXISTS job_url (
    job_id       INT    NOT NULL,          -- Job this URL belongs to
    url_id       INT    NOT NULL,          -- URL to be crawled for this job
    completed_on TIMESTAMP WITH TIME ZONE, -- The time stamp the crawl was completed

    FOREIGN KEY (url_id) REFERENCES url(id)
);

-- Results for each job.
CREATE TABLE IF NOT EXISTS job_result (
    job_id   INT  NOT NULL,
    refer_Id INT  NOT NULL, -- URL which this job URL result was found on
    url_id   INT  NOT NULL, -- URL for this result
    mime     TEXT DEFAULT '', -- content type this URL result references

    FOREIGN KEY (refer_id) REFERENCES url(id),
    FOREIGN KEY (url_id)   REFERENCES url(id)
);
CREATE UNIQUE INDEX job_result_pair ON job_result(job_id,refer_id,url_id);


-- Links a refer URL with a content URL
CREATE TABLE IF NOT EXISTS url_link (
    url_id   INT NOT NULL,
    refer_id INT NOT NULL
);
CREATE UNIQUE INDEX url_link_pair ON url_link (url_id, refer_id);

-- job URL still pending
CREATE TABLE IF NOT EXISTS url_pending (
	origin_id INT NOT NULL, -- The Job URL that this URL is a descendant of 
	url_Id    INT NOT NULL -- URL that is pending being crawled.
);
