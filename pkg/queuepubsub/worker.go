// Package queuepubsub provides support for transport of publish/subscribe
// messages to and from a message queue..
package queuepubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/rwool/saas-interview-challenge1/pkg/service"

	"github.com/go-kit/kit/log"

	"github.com/go-kit/kit/endpoint"
	"github.com/rwool/saas-interview-challenge1/pkg/service/queue"

	"github.com/pkg/errors"
)

const QueueName = "worker_parse_document"

func MakeQueuePubSubHandler(endpoint endpoint.Endpoint, queue queue.Queue, logger log.Logger) func(context.Context) {
	dataC := make(chan []byte)
	subLoop := makeSubscribeLoop(queue, logger, dataC, QueueName)
	return func(ctx context.Context) {
		go subLoop(ctx)
		for {
			select {
			case data := <-dataC:
				// Process incoming data asynchronously to not block other
				// requests.
				go func() {
					pdr, err := decodeWorkerParseDocumentRequest(ctx, data)
					if err != nil {
						// TODO: Better way of handling errors here.
						//  Possibly a dead letter queue or way of notifying the
						//  creator of the request?
						_ = logger.Log("LEVEL", "ERROR", "MESSAGE", err.Error())
						return
					}
					_ = logger.Log("LEVEL", "DEBUG", "MESSAGE", fmt.Sprintf("Received document request %s", pdr.(service.DocumentID).ID))
					// TODO: Create sibling trace span here.
					dfrr, err := endpoint(ctx, pdr)
					if err != nil {
						_ = logger.Log("LEVEL", "ERROR", "MESSAGE", err.Error())
						return
					}
					err = encodeWorkerParseDocumentResponse(ctx, queue, QueueName, logger, dfrr)
					if err != nil {
						_ = logger.Log("LEVEL", "ERROR", "MESSAGE", err.Error())
						return
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}
}

// makeSubscribe loop returns a function for sending messages from a
// subscription over a channel.
func makeSubscribeLoop(q queue.Queue, l log.Logger, c chan []byte, channel string) func(context.Context) {
	return func(ctx context.Context) {
		_ = l.Log("LEVEL", "INFO", "MESSAGE", fmt.Sprintf("Beginning subscription for %s", channel))
		for {
			data, err := q.Pull(ctx, channel)
			// Check if the Pull was stopped from a context cancellation or
			// deadline.
			select {
			case <-ctx.Done():
				return
			default:
			}
			if err != nil {
				_ = l.Log("LEVEL", "ERROR", "MESSAGE", err.Error())
				continue
			}

			// Check to see if the context is done before sending the data to
			// minimize chance of randomly choosing to send the data multiple
			// times despite the context being done.
			// See: https://golang.org/ref/spec#Select_statements .
			select {
			case <-ctx.Done():
				return
			default:
			}
			select {
			case c <- data:
			case <-ctx.Done():
			}
		}
	}
}

func jsonDecode(b []byte, into interface{}) error {
	r := bytes.NewReader(b)
	decoder := json.NewDecoder(r)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(into)
	return errors.Wrap(err, "unable to read out JSON data")
}

func decodeWorkerParseDocumentRequest(_ context.Context, request []byte) (interface{}, error) {
	var doc service.DocumentID
	err := jsonDecode(request, &doc)
	// TODO: Handle child trace span creation here.
	return doc, err
}

func encodeWorkerParseDocumentResponse(ctx context.Context, q queue.Queue, ch string, l log.Logger, r interface{}) error {
	if v, ok := r.(endpoint.Failer); ok && v.Failed() != nil {
		// TODO: Better internal server error handling for pub/sub.
		_ = l.Log("LEVEL", "ERROR", "MESSAGE", v.Failed().Error())
		return nil
	}
	// TODO: Handle sibling trace span creation here.
	data, err := json.Marshal(r)
	if err != nil {
		return errors.WithStack(err)
	}
	err = q.Push(ctx, ch, [][]byte{data})
	return errors.WithStack(err)
}
