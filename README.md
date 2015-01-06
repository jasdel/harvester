# Overview #
------------
Harvester is a distributed web crawler service. It runs on one or multiple hosts and will
accept requests for URLs to be crawled. The URLs are scheduled into a queue, and processed
in the order they are received.

There are three main parts that make up the service
- Web Service (web_server): Receives requests for jobs, and schedules them with the queue service
- Queue service (foreman): Receives URLs to be crawled and schedules them after filtering against the services URL cache.
- Worker service (worker): Receives URLs and crawls them. All URLs encountered are sent back to the queue service to be queued up.

Each layer can be scaled independently of the others. gnatsd NATS service provides the message queue functionality between the service parts. With Harvester's architecture the three layers could be split into clusters with multiple gnatsd service instances feeding the layers. A Postgreql database provides the persistent storage and state for the service. The database will be the bottle neck for raw throughput.

# Usage #
---------
All responses are JSON formated.

**Scheduling a Job**:
To schedule a job a POST http request is made to the web_server with a body containing a list of new line separated URLs. The URLs are required to have a scheme of http, https, or no protocol at all, and a host. If no protocol is provided http will be substituted in. Any duplicate entries in the list will be removed also.
```
curl -X POST --data-binary @- "http://localhost:8080" << EOF
https://www.google.com
example.com
EOF
> {jobId: <jobID>}
```
Note: URLs with different scheme/protocols will be crawled as different tasks of the Job, and will show up as different entries in the job result.

**Retrieve Job Status**:
The Job status can be requested any time after a job has been scheduled. Requesting a job id which does not exist will return a 404 error code with an error message stating the job id was not found.
```
curl -X GET "http://localhost:8080/status/<jobId>" 
> {completed: 0, pending: 2, elapsed: 1m23s, urls:{"https://www.google.com":false, "http://example.com":false})
```

**Retrieve Job Result**:
The job result can be requested at any time after a job has been scheduled. Requesting a job id which does not exist will return a 404 error code with an error message stating the job id was not found.
```
curl -X GET "http://localhost:8080/result/<jobId>"
> { "https://www.example.com": ["http://www.example.com/somePath", ...], ...} 
```

**Filter Results**:
Filter results for a specific mime type, e.g. all images (image/*). Any content crawled or discovered which has an image mime type, or image extension (jpeg, jpg, png, gif) will be available under the image filter.
```
curl -X GET "http://localhost:8080/result/1?mime=image"
> { "https://www.example.com": ["https://www.example.com/someImage.png", ...], ...} 
```
The mime filter is not limited to images, and can be used with any mime type. For example to find all javascript files discovered while crawling a Job use the mime filter of "?mime=text/javascript". 

# Setup #
---------
**Harvester**:
```
go get github.com/jasdel/harvester/web_server
go get github.com/jasdel/harvester/foreman
go get github.com/jasdel/harvester/worker
```
**gnatsd**:
```
go get github.com/apcera/gnatsd
gnatsd
```
**Docker & Postgreql**
```
curl -sSL https://get.docker.com/ubuntu/ | sudo sh
sudo docker build -t eg_postgresql ./setup
```
**Start Postgresql and Inject Tables**
```
$ sudo docker run --rm -p 24001:5432 --name pg_test eg_postgresql
$ psql -h localhost -p 24001 -d docker -U docker --password < setup/db.sql
```

# Design #
----------
![Alt text](https://rawgit.com/jasdel/harvester/master/images/HarvesterHighLevel.svg "High level architecture")
![Alt text](https://rawgit.com/jasdel/harvester/master/images/HarvesterDB.svg "Database table architecture")

# Design Decisions #
--------------------

**Dependences**:
- Postgresql: 
- Docker Container for Postgresql:
- gnatsd Message Queue: 

# Short Comings / Improvements #
-------------------------------
- Web Server does not cache results of completed jobs. Each request for results requires the results to be extracted out of the db. An improvement would be to cache the result to file, and have the web server's reverse proxy (nginx) service the static content instead. An alternative to storing the result to disk would be to keep the previous X results in memory.
- Web Server does not limit the number of URLs, or size of content that it processes during a job schedule request. This will allow very large crawl request to have a significant negative impact on the service. A possible solution would be to limit the number of URLs which will be parsed, and only processes up to X bytes from the request body
- The logic used by the foreman when processing cache URL's and the worker's processing of a crawled URL are very similar. It should be possible to refactor the two so that they share more of the same code base reducing the chance logic bugs producing different results based if a URL is cached or not.
- The way the service parts are configured are via a JSON file. It is simple to use for single instances, but can create complications for multiple instances. A more robust configuration system that pulls in configuration from environment or command line would provide a more easier to configure multiple instances.
- DB Queries are only tested at runtime by manual testing at the moment. 

- Workers should have some kind of per domain throttling
- Workers should parse, and respect servers robots.txt file
- Workers should re-queue URLs which fail with 50x status or connection errors, and re-queue to try again later.
- Workers could support gzip so that the request payloads are smaller
- Workers could use headless browser for more robust crawling of a page so dynamic JS pages could be crawled.
