# ForgeRock SaaS Software Engineer Coding Challenge

### What the Service Does
The service takes in "documents" and performs a simple word count on it.

#### How it Works
An HTTP service accepts POST requests that contain the "document" to parse.
The API service will check if the result of the parse is cached in Redis, and if it is, return it.
If there is a cache miss, the data is sent to a queue in Redis for one of the
worker nodes to process.

For simplicity, all of the containers of this program run both the HTTP API
service and the worker service.
The HTTP API can only be accessed from port 8080 with the "api" container.
All of the worker nodes can pull off of the request queue to process messages.
The different worker nodes processing the data can be seen by the different log 
lines with docker-compose.

Once the messages have been processed by a worker, they are written to Redis
with the key being a Base64 encoded SHA256 hash of the document contents. The
API service polls this key until it gets the parsing results or times out.

The structure of the requests is based on the following Go struct:
```Go
// DocumentRequest is a request for a document to be processed.
type DocumentRequest struct {
	Document        string `json:"document"`
	DurationSeconds int    `json:"duration_seconds"`
}
```

If the DurationSeconds is set, the worker will wait that many seconds before
writing its result to Redis.
If the API is called quickly enough with the same document, the duration may be
skipped due to a cache hit skipping the worker.

To run/build: `docker-compose up`

To upload a document to parse: `curl -X POST http://localhost:8080/document -H 'Host: 127.0.0.1' -d '{"document": "This is a a test document"}'`

To run all unit tests: `go test ./...`

To run all integration tests:
1. Have a Redis instance running (that does not require credentials to use).
2. `REDIS_ADDRESS=<REDIS ADDRESS HERE> go test -tags integration ./...`

### Shortcomings
This code leaves multiple things unimplemented that would be expected in a microservice.
These include, but are not limited to tracing, metrics, and better logging.

The service is built on top of Go kit, and because of the separation of the
layers, it is easy to add middleware other improvements.

The HTTP response status codes are not completely correct currently as they now
just indicate failure or not failure.

### Scaling
To scale this project, a better queue implementation would be needed to avoid 
data loss in the event that a worker node fails. This is because the current
implementation pulls messages off of a queue in Redis to process them. Once the message is read, it can no longer be read from Redis, in the event of a node failure.

Load balacing the HTTP API and have more nodes/sharding for Redis could improve
the scaling of the system currently.

### Parallel vs. Sequential
To implement sequential tasks, a queue per sequence of tasks could be used to
impose ordering.
Coordination of the workers would be required so that only one message at a time
could be processed from these queues.

The requests are already processed in parallel with the current limitation, with
the exception of Redis which serializes the message queue.
Less contention could be had by having more queues spread between more Redis
nodes.

