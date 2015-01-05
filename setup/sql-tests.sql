--SQL
-- Select All job result with URL strings instead of id
SELECT refer.url as refer, url.url as url, url.mime as mime
FROM job_result
LEFT JOIN url AS url on job_result.url_id = url.id
LEFT join url as refer on job_result.refer_id = refer.id
WHERE job_result.job_id = 1 and url.mime LIKE 'image%'
;

-- GetAll URLs With Refer ById
SELECT url.id, url.url, url.mime, url.crawled_on
FROM url_link
LEFT JOIN url on url_link.url_id = url.id
WHERE url_link.refer_id = 1
;

-- Get JobURLs with URL from job_id
SELECT job_url.job_id, job_url.url_id, url.url, job_url.completed_on
FROM job_url
LEFT JOIN url AS url on job_url.url_id = url.id
WHERE job_url.job_id = 1
;

-- Insert with out result
INSERT INTO url_link (url_id, refer_id)
	SELECT 3, 4
	WHERE NOT EXISTS (SELECT 1 FROM url_link WHERE url_id = 3 AND refer_id = 4)
;


-- Insert with result
WITH s AS (
    SELECT id, url
    FROM url
    WHERE url = 'http://www.google.com'
), i as (
    INSERT INTO url (url, mime)
    SELECT 'http://www.google.com', 'text/css'
    WHERE NOT EXISTS (SELECT 1 FROM s)
    RETURNING id
)
SELECT id
from i
union all
select id
from s
;