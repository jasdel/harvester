# Overview #
------------
Harvester is a distributed web crawler service. It runs on one or multiple hosts and will
accept requests for URLs to be crawled. The URLs are scheduled into a queue, and processed
in the order they are received.

There are three main parts that make up the service
- Web Service (web_server): Receives requests for jobs, and schedules them with the queue service
- Queue service (foreman): Receives URLs to be crawled and schedules them after filtering against the services URL cache
- Worker service (worker): Receives URLs and crawls them. All URLs encountered are sent back to the queue service to be queued up.

Each can be layer can be scaled vertically independently of each other. A Postgres database provides the persistent storage for the service, and will be the bottle neck for raw through put. Gnatsd NATS service provides the message queue functionality between the services. With Harvester's architecture the three layers could be split into clusters with multiple gnatsd service instances feeding the layers.

# Dependencies #
----------------
- testify: Simple assert/require test syntax sugar
- postgresql w/ go bindings: persistent storage.
- gnatsd w/ go bindings: message queue between web server => queue server <=> worker


# Usage #
---------
	# Schedule a job
	$ curl -X POST --data-binary @- "http://localhost:8080" << EOF
	https://www.google.com
	http://example.com
	EOF
	> {jobId:1}

	# Retrieve the status of a job
	$ curl -X GET "http://localhost:8080/status/1" 
	> {completed: 0, pending: 2}

	# Retrieve the results of a job
	$ curl -X GET "http://localhost:8080/result/1"
	> { "<refer>": [<all url>, ...], ...} 

	# Filter results for a specific mime type, e.g. image/*
	$ curl -X GET "http://localhost:8080/result/1?mime=image"
	> { "<refer>": [<image only url>, ...], ...} 

# Setup #
---------
- gnatsd for message queues

	$ go get github.com/apcera/gnatsd
	$ go get github.com/apcera/nats
	$ gnatsd

- Get Docker

	$ curl -sSL https://get.docker.com/ubuntu/ | sudo sh

- Create postgresql container from docker file

	$ sudo docker build -t eg_postgresql ./setup

- Setup and start postgresql

	$ sudo docker run --rm -p 24001:5432 --name pg_test eg_postgresql
	$ psql -h localhost -p 24001 -d docker -U docker --password < setup/db.sql


# Design #
----------
![Alt text](/images/HarvesterHighLevel.svg "High level architecture")
![Alt text](/images/HarvesterDB.svg "Database table architecture")

TODO: replace image here:
- Reverse Proxy (e.g: nginx) provides the multiplexing between Web server instances
- Web server serves incoming requests to schedule jobs, check status, or receive results
  - If the request is for status or results the web server only needs to talk to the database
  - If a job schedule request is received, the url will be packaged and sent to the queue service
- NATS URL Queue
  - gnatsd provides the message queue transport between web server and queue service
  - NATS queue Receivers are configured so only a single service instance will pull an item off of a queue.
-  Queue service pulls a URL off of the queue, check if it is already cached, and if not send the item to the worker queue. If the item is cache the worker won't need to process it, and its direct descendants can be queued up instead.
- NATS Worker queue
- Worker instances pull items from the worker queue, crawls them, and adds their results to the Job result list. Nested URLs are then queued up to be crawled recursively if the max depth from the origin hasn't been reached yet.

# Design Decisions #
--------------------

# Short Comings #
-----------------
- Web Server does not cache results of completed jobs. Each request for results requires the results to be extracted out of the db.  An improvement would be to cache the result to file, and have the web server's reverse proxy (nginx) service the static content instead.An alternative to storing the result to disk would be to keep the previous X results in memory.

# Possible Improvements #
-------------------------
- Update logging statement to dynamically get file, function and line number for error messages
- Workers use per domain throttling
- Workers parse, and respect servers robots.txt file
- Workers could re-queue URLs which fail with 50x status or connection errors, and try again later.
- Workers could support gzip so that the request payloads are smaller
- Workers could use headless browser for crawling so dynamic JS pages could be crawled.
