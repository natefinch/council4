package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"sync"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// cardResult is the parser-coverage outcome for one corpus record.
type cardResult struct {
	index     int
	name      string
	eligible  bool
	complete  bool
	resolving int
	exact     int
	uncovered []uncoveredItem
	blockers  []parser.CoverageBlocker
}

// uncoveredItem is one contiguous uncovered run with its normalized cluster key.
type uncoveredItem struct {
	text       string
	normalized string
	blocker    parser.CoverageBlocker
}

func parseCorpus(input io.Reader, workers int) ([]cardResult, error) {
	if workers < 1 {
		workers = runtime.NumCPU()
	}
	decoder := json.NewDecoder(input)
	token, err := decoder.Token()
	if err != nil {
		return nil, fmt.Errorf("reading bulk-data array: %w", err)
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '[' {
		return nil, errors.New("bulk data from Scryfall must be a top-level JSON array")
	}

	jobs := make(chan job)
	results := make(chan cardResult)
	var workersDone sync.WaitGroup
	workersDone.Add(workers)
	for range workers {
		go func() {
			defer workersDone.Done()
			for item := range jobs {
				results <- coverCard(item)
			}
		}()
	}
	go func() {
		workersDone.Wait()
		close(results)
	}()

	decodeError := make(chan error, 1)
	go func() {
		defer close(jobs)
		sent := 0
		for decoder.More() {
			var card cardgen.ScryfallCard
			if err := decoder.Decode(&card); err != nil {
				decodeError <- fmt.Errorf("decoding card %d: %w", sent, err)
				return
			}
			jobs <- job{index: sent, card: card}
			sent++
		}
		if _, err := decoder.Token(); err != nil {
			decodeError <- fmt.Errorf("closing bulk-data array: %w", err)
			return
		}
		decodeError <- nil
	}()

	var all []cardResult
	for result := range results {
		all = append(all, result)
	}
	if err := <-decodeError; err != nil {
		return nil, err
	}
	return all, nil
}

type job struct {
	index int
	card  cardgen.ScryfallCard
}

func coverCard(item job) cardResult {
	card := item.card
	result := cardResult{index: item.index, name: card.Name}
	if _, excluded := (cardgen.CorpusPolicy{}).Exclusion(card); excluded {
		return result
	}
	result.eligible = true
	result.complete = true
	for _, face := range cardgen.ParseCardFaces(&card) {
		coverage := parser.DocumentCoverage(face.Document)
		result.resolving += coverage.ResolvingEffects
		result.exact += coverage.ExactEffects
		if !coverage.Complete {
			result.complete = false
		}
		for i := range coverage.Components {
			component := coverage.Components[i]
			result.blockers = append(result.blockers, component.Blocker)
			result.uncovered = append(result.uncovered, uncoveredItem{
				text:       component.Text,
				normalized: normalizeCluster(component.Text),
				blocker:    component.Blocker,
			})
		}
	}
	return result
}
