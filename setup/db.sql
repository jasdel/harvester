/*
 Create Tables used by harvester
*/
CREATE TABLE IF NOT EXISTS job (
    id           serial                   PRIMARY KEY,
    created_on   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_on TIMESTAMP WITH TIME ZONE
);

CREATE TABLE IF NOT EXISTS job_urls (
    id        serial PRIMARY KEY,
    job_id    INT    NOT NULL,
    url       text   NOT NULL,
    completed BOOL   NOT NULL DEFAULT FALSE,
    FOREIGN key (job_id) REFERENCES job(id)
);

CREATE TABLE IF NOT EXISTS url_queue (
    id        serial PRIMARY KEY,
    processed BOOL   NOT NULL DEFAULT FALSE,
    url       TEXT   NOT NULL
);

CREATE TABLE IF NOT EXISTS url (
    id         serial PRIMARY KEY,
    url        TEXT   NOT NULL,
    referer    TEXT   NOT NULL, --origin URL encountered for this URL record
    created_on TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX url_pair ON url (url, referer);

/* Function for notifying on url_queue inserts */
-- cleanup: DROP FUNCTION url_queue_notify_trigger() CASCADE;
CREATE OR REPLACE FUNCTION url_queue_notify_trigger()
	RETURNS trigger AS $$
	DECLARE
	BEGIN
		PERFORM pg_notify('url_queue_watchers', null);
		RETURN new;
	END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER watched_url_queue_trigger AFTER INSERT ON url_queue
	FOR EACH ROW EXECUTE PROCEDURE url_queue_notify_trigger();