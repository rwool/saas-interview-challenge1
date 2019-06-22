package service

import (
	"context"
	"net"
	gohttp "net/http"
	"os"
	"sync"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/rwool/saas-interview-challenge1/pkg/endpoint"
	"github.com/rwool/saas-interview-challenge1/pkg/http"
	"github.com/rwool/saas-interview-challenge1/pkg/queuesubscribe"
	"github.com/rwool/saas-interview-challenge1/pkg/service"
	"github.com/rwool/saas-interview-challenge1/pkg/service/keyvalue"
	"github.com/rwool/saas-interview-challenge1/pkg/service/queue"
)

const workerQueueName = "worker_document_parser"

func getRedisClient() (*redis.Client, error) {
	address, ok := os.LookupEnv("REDIS_ADDRESS")
	if !ok {
		return nil, errors.New("missing Redis address")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     "",
		DB:           0,
		MaxRetries:   10,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	if err := client.Ping().Err(); err != nil {
		return nil, errors.WithStack(err)
	}
	return client, nil
}

// Run runs the API and Worker services.
//
// Note that normally there would only be one service being initialized in a
// function like this, but due to the constraints of the of the problem, two
// services are being set up.
func Run() {
	setup()
}

// Config contains all of the configuration for running the services.
type Config struct {
	Log    log.Logger
	Queue  queue.Queue
	KeyVal keyvalue.KeyValue
}

func setup() {
	l := log.NewJSONLogger(os.Stderr)

	rc, err := getRedisClient()
	if err != nil {
		_ = l.Log("LEVEL", "ERROR", "MESSAGE", err)
		os.Exit(1)
	}
	q := queue.NewRedisAdapter(rc)
	kv := keyvalue.NewRedisAdapter(rc)

	// Business logic.
	apiService := service.NewAPIService(q, kv, workerQueueName, l)
	workerService := service.NewWorkerService(service.WorkerServiceConfig{
		Queue:   q,
		KeyVal:  kv,
		Log:     l,
		Channel: workerQueueName,
	})

	// Endpoints.
	apiEndpoint := endpoint.MakeAPIProcessDocumentEndpoint(apiService)
	workerEndpoint := endpoint.MakeWorkerParseDocumentEndpoint(workerService)

	// Transports.
	httpHandler := http.NewAPIHTTPHandler(apiEndpoint, nil)
	subscriber := queuesubscribe.MakeWorkerHandler(queuesubscribe.Config{
		Endpoint: workerEndpoint,
		Queue:    q,
		Log:      l,
		Channel:  workerQueueName,
	})

	server, err := serveHTTP(httpHandler)
	if err != nil {
		_ = l.Log("LEVEL", "ERROR", "MESSAGE", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Message loops.
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		server(ctx, l)
	}()
	go func() {
		defer wg.Done()
		subscriber(ctx)
	}()
	wg.Wait()
}

func serveHTTP(h gohttp.Handler) (func(context.Context, log.Logger), error) {
	// Separate listening and serving to capture listen errors.
	l, err := net.Listen("tcp", "0.0.0.0:8080")
	if err != nil {
		return nil, errors.Wrap(err, "unable to create TCP listener")
	}

	return func(ctx context.Context, logger log.Logger) {
		go func() {
			<-ctx.Done()
			if err := l.Close(); err != nil {
				_ = logger.Log("LEVEL", "WARN", "MESSAGE", err)
			}
		}()
		err := gohttp.Serve(l, h)
		_ = logger.Log("LEVEL", "ERROR", "MESSAGE", err)
	}, nil
}
