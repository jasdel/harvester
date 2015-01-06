# Overview #
------------
Harvester is a distributed web crawler service. It runs on one or multiple hosts and will accept requests for URLs to be crawled. The URLs are scheduled into a queue, and distributed between workers to be crawled recursively up to a max depth.

# Usage #
---------
All responses are JSON formated.

**Scheduling a Job**:
To schedule a job a POST http request is made to the web_server with a body containing a list of new line separated URLs. The URLs are required to have a host and a scheme of http, https, or no protocol at all. If no protocol is provided http will be used. Any duplicate entries in the list will be removed.
```
curl -X POST --data-binary @- "http://localhost:8080" << EOF
https://www.google.com
example.com
EOF
> {jobId: <jobID>}
```
Note: URLs with different scheme/protocols will be crawled as different tasks of the Job, and will show up as different entries in the job result.

To force crawling a cached previously crawled URL add the 'forceCrawl' query parameter to the schedule job API call. If the 'forecCrawl' parameter is present the URL, and all of its descendants, will be crawled regardless of their cache status. A value for the query parameter is not required, and will be ignored if one is provided.

**Retrieve Job Status**:
The Job status can be requested any time after a job has been scheduled. The status call will contain the counts of completed vs pending, the total running time of the job, and a breakdown of the Job URL individual status.

Requesting a job id which does not exist will return a 404 error code with an error message stating the job id was not found.
```
curl -X GET "http://localhost:8080/status/<jobId>" 
> {completed: 0, pending: 2, elapsed: 1m23s, urls:{"https://www.google.com":false, "http://example.com":false})
```

**Retrieve Job Result**:
The job result can be requested at any time after a job has been scheduled, and will return partial results until the job is completed. The result will contain URLs grouped in a list under the URL that they were found on.

Requesting a job id which does not exist will return a 404 error code with an error message stating the job id was not found.
```
curl -X GET "http://localhost:8080/result/<jobId>"
> { "https://www.example.com": ["http://www.example.com/somePath", ...], ...} 
```

**Filter Results**:
Filter results for a specific mime type, e.g. all images (image/*). Any content crawled URL which has an image mime type, or extension (jpeg, jpg, png, gif) will be available under the image filter.
```
curl -X GET "http://localhost:8080/result/1?mime=image"
> { "https://www.example.com": ["https://www.example.com/someImage.png", ...], ...} 
```
The mime filter is not limited to just images, and can be used with any mime type. For example to find all javascript files discovered while crawling a Job use the mime filter of "?mime=text/javascript". 

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
cd <harvester path>
sudo docker build -t eg_postgresql ./setup
```
**Start Postgresql and Inject Tables**
```
sudo docker run --rm -p 24001:5432 --name pg_test eg_postgresql
cd <harvester path>
psql -h localhost -p 24001 -d docker -U docker --password < setup/db.sql
```

# Configuration #
-----------------
Each part of the harvester service has its own configuration file, and is specified via the "-config <filename>" command line argument parameter.

web_server also takes and additional parameter, "-addr <bind addr>". If set, this parameter will override the web_server's configuration file's "httpAddr". This simplifies the process of running multiple instances of the web server without needing multiple configuration files.

The service will crawl URLs recursively up to a max depth from the original job URL. The max depth is a configuration setting in the foreman and worker's config.json files.

The service will cache crawled URLs and not crawl them again until the cache max age duration has expired. The foreman's configuration file specifies the duration of the cache max age as 'cacheMaxAge'. Syntax of this field is specified at "http://golang.org/pkg/time/#ParseDuration".

# Design & Architecture #
-------------------------
There are three main parts that make up the harvester service.
- Web Service (web_server): Receives requests for jobs, and schedules them with the queue service. If a Job URL isn't already known in the url table it will be inserted. Job URLs are published to the URL Queue
- Queue service (foreman): Receives URLs from the URL queue to be crawled. If a URL has already been crawled the foreman will query for all of its descendants and enqueue them into the URL Queue. If the URL hasn't yet been crawled it will be published to the Work queue. All from cache job results are added to the job_result table by the foreman.
- Worker service (worker): Receives URLs from the Work queue and crawls them. All URLs encountered are sent back to the URL queue. All crawled job results are added to the job_result table by the worker.

![Alt text](https://rawgit.com/jasdel/harvester/master/images/HarvesterHighLevel.svg "High level architecture")

Each layer can be scaled independently of the others. gnatsd NATS service provides the message queue functionality between the service parts. With Harvester's architecture, the three layers could be split into clusters with multiple gnatsd service instances feeding the layers. A Postgreql database provides the persistent storage and state for the service. The database will be the bottle neck for raw throughput.

![Alt text](https://rawgit.com/jasdel/harvester/master/images/HarvesterDB.svg "Database table architecture")

The database tables are split into two main groups. URL and Job.

The URL group contains the actual URL value via the url table. All links between URLs with the url_link table. Both of these tables enforce unique indexes to prevent duplicate entries. Initially duplicate entries was a hurdle I was having difficulty working around, until I learned more robust SQL queries for inserting into the database which were fault tolerant and ignored the insert if the unique index already existed.

The Job group contains all information pertaining to a scheduled jobs, and their results. The job table represents just the scheduled job and when it was created. This provided a simple and reliable way to ensure job ids were unique.

The URLs for the job are inserted into the the job_url table. The job_url table would contain a reference to the job it was created for, and the URL id for a reference to the url table. The job_url table also contains when the job was completed. When determining the status of a job the Job URL's completed_on field is used to determine if a job has been completed, and if so, the Job's running time.

The job_result table contains all results for all jobs. The records are grouped under the job_id, refer_id, and url_id.  These three values make a unique entry. When the job result is written to a client, the results will be grouped under the URL they were directly crawled from (refer).

The job_pending table is used to temporarily keep track of a job's crawling status. It does this by storing entries for each recursivily crawled URL. The pending entry is then removed once the URL has been crawled. Once a Job URL (job_id, origin_id) no longer has any pending entries the Job URL is marked as complete.


# Dependences #
---------------
- Postgresql: Postgresql was chosen, because it was very simple to setup within a docker container. I also already had a little experience with the database in the past and felt I could iterate with it quickly. The github.com/lib/pq driver was also very easy to use. I ended up learning a lot about SQL statements using this database.
- Docker Container for Postgresql: A Docker container for Postgresql simplified starting and stopping the server without polluting my development system with Postgresql's footprint. Using a container also simplified deploying the database, pre-configured to any host.
- gnatsd Message Queue: gnatsd was chosen because it was dead simple to install, setup, and run. The go bindings were also very simple to understand and use. I briefly looked at zeromq, but zeromq was significantly more complex to use, and required me to either build my own intermediate layer to connect processes together, or have the service processes know about, and be directly connected to, each other.

# Short Comings / Improvements #
-------------------------------
- Web Server does not cache results of completed jobs. Each request for results requires the results to be queried from the database. An improvement would be to cache the result to file, and have the web server's reverse proxy (nginx) service the static content instead.
- Web Server does not limit the number of URLs, or size of content that it processes during a job schedule request. This will allow very large crawl request to have a significant negative impact on the service. A possible solution would be to limit the number of URLs which will be parsed, and only processes up to X bytes from the request body before bailing.
- The logic used by the foreman when processing cached URLs and the worker's processing of a crawled URLs are very similar. It should be possible to refactor the two so they share the same code. This would reduce the chance of logic bugs producing different results if a URL was cached or not.
- The way the service parts are configured via a JSON file could be improved and made more flexible. It is simple to use the service as single instances, but it becomes more complicated for multiple instances. A more robust configuration system that pulls in configuration from environment or command line would provide a easier to configuration process.
- DB Queries are only tested at runtime by manual testing. This allows logic and SQL bugs to go hidden until they are discovered at runtime. Testing these queries through unit tests, and integration tests should be implemented to improve the confidence in the code.
- Workers should have some kind of per domain throttling.
- Workers should parse, and respect servers robots.txt file.
- Workers could re-queue URLs which fail with 50x status or connection errors, and re-queue to try again later.
- Workers could support gzip so that the request payloads are smaller.
- Workers could use headless browser for more robust crawling of a pages so dynamic JS pages could be crawled.
