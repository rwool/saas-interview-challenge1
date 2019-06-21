// Package service implements the business logic for the document parsing
// system.
package service

import (
	"bufio"
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/rwool/saas-interview-challenge1/pkg/service/keyvalue"

	"github.com/rwool/saas-interview-challenge1/pkg/service/queue"

	"github.com/pkg/errors"
)

// WorkerService wraps the set of methods for a document parser worker.
type WorkerService interface {
	ParseDocument(ctx context.Context, doc DocumentID) (DocumentFrequencyReport, error)
}

// DocumentRequest represents a document to parse.
type DocumentID struct {
	DocumentRequest
	ID string
}

// Frequency describes the frequency of a word.
type Frequency struct {
	Word      string
	Frequency int
}

type DocumentFrequencyReport struct {
	DocumentFrequenciesResponse
}

type WorkerServiceConfig struct {
	Queue   queue.Queue
	KeyVal  keyvalue.KeyValue
	Log     log.Logger
	Channel string
}

func NewWorkerService(conf WorkerServiceConfig) WorkerService {
	return newWorkerService(conf)
}

func newWorkerService(conf WorkerServiceConfig) *workerService {
	return &workerService{
		q:       conf.Queue,
		kv:      conf.KeyVal,
		log:     conf.Log,
		channel: conf.Channel,
	}
}

type workerService struct {
	log     log.Logger
	q       queue.Queue
	kv      keyvalue.KeyValue
	channel string
}

func topN(m map[string]int, n int) []Frequency {
	out := make([]Frequency, 0, len(m))
	for k, v := range m {
		out = append(out, Frequency{
			Word:      k,
			Frequency: v,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Frequency > out[j].Frequency
	})

	length := n
	if len(out) < n {
		length = len(out)
	}

	return out[:length]
}

// ParseDocument parses a document to find the frequencies of the different
// words in the document.
//
// The ID of the stored document and the top 10 frequencies are returned.
func (w *workerService) ParseDocument(ctx context.Context, doc DocumentID) (DocumentFrequencyReport, error) {
	wait := time.NewTimer(time.Duration(doc.DurationSeconds) * time.Second)
	defer wait.Stop() // Don't leak the timer.
	waitOrCancel := func() {
		select {
		case <-wait.C:
			return
		case <-ctx.Done():
			return
		}
	}
	defer waitOrCancel()

	// Simple word scanner. Does not respect punctuation or capitalization
	// differences.
	words := make(map[string]int)
	reader := strings.NewReader(doc.Document)
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := scanner.Text()
		if _, ok := words[word]; !ok {
			words[word] = 1
		} else {
			words[word]++
		}
	}
	if err := scanner.Err(); err != nil {
		return DocumentFrequencyReport{}, errors.Wrap(err, "error while scanning document")
	}

	id := doc.ID
	if id == "" {
		id = createStringSHA256(doc.Document)
	}

	dfr := DocumentFrequencyReport{
		DocumentFrequenciesResponse: DocumentFrequenciesResponse{
			DocumentID:  id,
			Frequencies: topN(words, 10),
		},
	}

	// Pretend this work is more intensive than it actually is.
	waitOrCancel()

	dfrBytes, err := json.Marshal(dfr)
	if err != nil {
		return DocumentFrequencyReport{}, errors.WithStack(err)
	}
	if err := w.kv.Store(ctx, id, dfrBytes, 30*time.Second); err != nil {
		// Suppress error because the normal way of handling errors is to log
		// errors, but that is not working.
		_ = w.log.Log("LEVEL", "ERROR", "MESSAGE", err.Error())
		// Failure in a non-essential operation. Continue.
	}
	return dfr, nil
}
