# Goals #
---------
- Create a HTTP server
	- the server will publish endpoints
	- / : Receive list of URLs in the body to enqueue, and returns a job id for checking the status
	- /status/:jobId : to receive the status of a job
	- /result/:jobId : to receive the result of a job
	* /list : returns a list of pending jobs, status, and time elapsed
	* /cancel/:jobId : cancels a pending job id
- Push URLs to inspect to a queue that will be pulled off of by workers
- Workers will parse the document for URLs and add those new URLs to a queue to be checked.
	- If the URL is for an image (png, jpg, gif) add the URL to the result list for the jobId
	- Only add the first layer of URLs to the queue, secondary URLs should just be scrapped
- The result list of URLs should be persisted. 
- A Queue of jobs maintained.

Parts:
- Entry(s). Provides public API for creating, and viewing jobs.
	- When a job is complete Entry will query the storage directly for results of job
- Foreman. Provides synchronization of jobs queued and status. Makes Queued URLs available to workers.
- Worker(s). Request queued URLs from Foreman

- document parser. searches in a document for things that might be URLs.
	- <scheme>://<path>
	- pushes to a channel the list of URLs that were found
- Queued
	- check then extension of the URI path to guess if it is a image
	- if not try a head request to get the content type.
		- if HTML pull down the content and send to document parser


# Dependencies #
----------------
- testify: Simple assert/require test syntax sugar
- goji: web server router with url path params, and graceful shutdown
- gnatsd w/ go bindings: message queue between webserver => queue server <=> worker


# Possible Improvements #
-------------------------
- Update logging statement to dynamically get file, function and line number for error messages
- Workers use per domain rate limiting
- Workers parse, and respect the robots.txt file
- Workers could re-queue URLs which fail with 50x status of connection errors, and try again later.
- Workers could support gzip so that the request payloads are smaller
- Workers should use headless browser for crawling so that accurate link following, and JS functionality is supported.
- Usage of database optimized so fewer write queries are made

# Usage #
---------
curl -X POST --data-binary @- "http://localhost:8080" << EOF
https://www.google.com
http://example.com
EOF

# Setup #
---------
gnatsd for message queues
$ go get github.com/apcera/gnatsd
$ go get github.com/apcera/nats

Create containers from docker file
$ sudo docker build -t eg_postgresql ./setup

Setup and start postgresql
$ sudo docker run --rm -p 24001:5432 --name pg_test eg_postgresql
$ psql -h localhost -p 24001 -d docker -U docker --password < setup/db.sql
$ gnatsd -p 4442

# Notes #
---------
- http://postgres.cz/wiki/PostgreSQL_SQL_Tricks#Taking_first_unlocked_row_from_table
	- Use postgresql as a queue
- https://github.com/lib/pq/blob/master/listen_example/doc.go
	- Using pq to be notified when work is ready
- http://stackoverflow.com/questions/6151084/which-timestamp-type-to-choose-in-a-postgresql-database
	- Using timestamps with postgresql
- Creating notify functions
	- http://bjorngylling.com/2011-04-13/postgres-listen-notify-with-node-js.html
- http://www.postgresql.org/docs/9.2/static/plpgsql-trigger.html
	- triggers (39-4)