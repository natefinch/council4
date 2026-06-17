package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"runtime"
	"slices"
	"sync"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle/cmd/internal/cluster"
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// cardOutcome captures both signals for one corpus card. The parser-only
// coverage signal (is the card parser-complete, and which uncovered components
// block it) comes from this stream; the authoritative lowering signal (generated
// or not, and the blocking diagnostic summaries) is filled in later from
// compilecards' canonical report by applyAuthority. perCardGenerated is an
// independent in-process recompute kept only as a reconciliation guard against
// that authoritative set.
type cardOutcome struct {
	index            int
	id               string
	name             string
	eligible         bool
	parserComplete   bool
	uncovered        []uncoveredItem
	perCardGenerated bool

	// Authoritative fields populated by applyAuthority.
	generated         bool
	loweringSummaries []string
}

// uncoveredItem is one uncovered grammatical component with its owning blocker
// family and normalized cluster key, used to bucket the parser queue.
type uncoveredItem struct {
	component  string
	normalized string
	blocker    parser.CoverageBlocker
}

type job struct {
	index int
	card  cardgen.ScryfallCard
}

func parseCorpus(input io.Reader, workers int) ([]cardOutcome, error) {
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
	results := make(chan cardOutcome)
	var workersDone sync.WaitGroup
	workersDone.Add(workers)
	for range workers {
		go func() {
			defer workersDone.Done()
			for item := range jobs {
				results <- evaluateCard(item)
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

	var all []cardOutcome
	for result := range results {
		all = append(all, result)
	}
	if err := <-decodeError; err != nil {
		return nil, err
	}
	slices.SortFunc(all, func(a, b cardOutcome) int {
		return a.index - b.index
	})
	return all, nil
}

// evaluateCard computes the parser signal and the per-card guard signal for one
// card. Excluded cards return an ineligible outcome so they are dropped before
// routing, exactly as compilecards and parsercoverage drop them. The
// authoritative generated/lowering signal is applied separately from
// compilecards' report; this function never decides generated membership.
func evaluateCard(item job) cardOutcome {
	card := item.card
	outcome := cardOutcome{index: item.index, id: card.ID, name: card.Name}
	if _, excluded := (cardgen.CorpusPolicy{}).Exclusion(card); excluded {
		return outcome
	}
	outcome.eligible = true
	outcome.parserComplete = parserSignal(&card, &outcome)
	outcome.perCardGenerated = perCardGenerated(&card)
	return outcome
}

// parserSignal records the parser-only coverage: the card is parser-complete
// when every executable face is parser-complete, and every uncovered component
// is collected so an incomplete card can be bucketed in the parser queue.
func parserSignal(card *cardgen.ScryfallCard, outcome *cardOutcome) bool {
	complete := true
	for _, face := range cardgen.ParseCardFaces(card) {
		coverage := parser.DocumentCoverage(face.Document)
		if !coverage.Complete {
			complete = false
		}
		for i := range coverage.Components {
			component := coverage.Components[i]
			outcome.uncovered = append(outcome.uncovered, uncoveredItem{
				component:  component.Text,
				normalized: cluster.Normalize(component.Text),
				blocker:    component.Blocker,
			})
		}
	}
	return complete
}

// perCardGenerated recompiles the card in-process to decide, per card alone,
// whether it would generate. It deliberately omits the corpus-wide collision and
// parse-rejection passes compilecards runs (rejectPathCollisions,
// rejectIdentifierCollisions, disambiguateCollisions): those passes demote some
// cards a single-card view considers clean, so this value can disagree with
// compilecards' authoritative report. That disagreement is exactly what the
// reconciliation guard surfaces; this is a guard signal, never the routing
// authority.
func perCardGenerated(card *cardgen.ScryfallCard) bool {
	identity, err := cardgen.GeneratedIdentity(card, false)
	if err != nil {
		return false
	}
	letter := identity.PackageName
	if len(letter) != 1 || letter[0] < 'a' || letter[0] > 'z' {
		return false
	}
	source, diagnostics, genErr :=
		(cardgen.ExecutableGenerator{IdentifierSuffix: identity.IdentifierSuffix}).
			GenerateCardSource(card, identity.PackageName)
	return genErr == nil && len(diagnostics) == 0 && source != ""
}
