/*
 Create Tables used by harvester
*/
CREATE TABLE IF NOT EXISTS job (
    id           serial                   PRIMARY KEY,
    created_on   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_on TIMESTAMP WITH TIME ZONE -- is this really needed since each sub URL will have an entry?
);

CREATE TABLE IF NOT EXISTS job_urls (
    id        serial PRIMARY KEY,
    job_id    INT    NOT NULL,
    url       text   NOT NULL,
    completed_on   TIMESTAMP WITH TIME ZONE,
    FOREIGN key (job_id) REFERENCES job(id)
);

CREATE TABLE IF NOT EXISTS url (
    id              serial PRIMARY KEY,
    mime            TEXT, -- content type this url links to
    url             TEXT   NOT NULL,
    referer         TEXT   NOT NULL, --origin URL encountered for this URL record
    has_descendants BOOL, -- null=not scanned, TRUE|FALSE scanned with or without descendants
    created_on      TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX url_pair ON url (url, referer);

CREATE TABLE IF NOT EXISTS url_pending (
	id serial PRIMARY KEY,
	origin TEXT NOT NULL,
	descendant TEXT NOT NULL,
)