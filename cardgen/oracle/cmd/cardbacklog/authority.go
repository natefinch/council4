package main

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
)

// compileAuthority is the canonical generated/unsupported/excluded partition
// produced by compilecards, keyed by Scryfall record id. compilecards is the
// single source of truth for whether a card lowers: it applies corpus-wide
// collision and parse-rejection passes that a per-card recompile cannot see, so
// cardbacklog routes on this report rather than on its own recompile.
type compileAuthority struct {
	// unsupported maps a record id to its distinct blocking diagnostic summaries.
	unsupported map[string][]string
	// excluded is the set of record ids compilecards excluded from eligibility.
	excluded map[string]bool
}

// compileReport is the subset of compilecards' JSON report cardbacklog consumes.
type compileReport struct {
	EligibleCount  int                  `json:"eligible_count"`
	GeneratedCount int                  `json:"generated_count"`
	ExcludedCount  int                  `json:"excluded_count"`
	Unsupported    []compileUnsupported `json:"unsupported"`
	Excluded       []compileExcluded    `json:"excluded"`
}

// compileUnsupported is one unsupported record in compilecards' report.
type compileUnsupported struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Diagnostics []compileDiagnostic `json:"diagnostics"`
}

// compileDiagnostic is one diagnostic attached to an unsupported record; only its
// summary is used to bucket the lowering queue.
type compileDiagnostic struct {
	Summary string `json:"summary"`
}

// compileExcluded is one excluded record in compilecards' report.
type compileExcluded struct {
	ID string `json:"id"`
}

func loadCompileAuthority(path string) (compileAuthority, compileReport, error) {
	file, err := os.Open(path)
	if err != nil {
		return compileAuthority{}, compileReport{}, fmt.Errorf("opening compile report: %w", err)
	}
	defer func() { _ = file.Close() }()
	var raw compileReport
	if err := json.NewDecoder(file).Decode(&raw); err != nil {
		return compileAuthority{}, compileReport{}, fmt.Errorf("decoding compile report: %w", err)
	}
	authority := compileAuthority{
		unsupported: make(map[string][]string, len(raw.Unsupported)),
		excluded:    make(map[string]bool, len(raw.Excluded)),
	}
	for i := range raw.Unsupported {
		entry := &raw.Unsupported[i]
		if entry.ID == "" {
			return compileAuthority{}, compileReport{},
				fmt.Errorf("compile report unsupported card %q has no id", entry.Name)
		}
		authority.unsupported[entry.ID] = distinctReportSummaries(entry.Diagnostics)
	}
	for i := range raw.Excluded {
		if raw.Excluded[i].ID == "" {
			return compileAuthority{}, compileReport{},
				fmt.Errorf("compile report excluded card %d has no id", i)
		}
		authority.excluded[raw.Excluded[i].ID] = true
	}
	return authority, raw, nil
}

func distinctReportSummaries(diagnostics []compileDiagnostic) []string {
	seen := make(map[string]bool, len(diagnostics))
	summaries := make([]string, 0, len(diagnostics))
	for i := range diagnostics {
		summary := diagnostics[i].Summary
		if summary != "" && !seen[summary] {
			seen[summary] = true
			summaries = append(summaries, summary)
		}
	}
	slices.Sort(summaries)
	return summaries
}

// reconciliation records where cardbacklog's independent per-card recompile
// disagrees with compilecards' authoritative report. A non-empty reconciliation
// means the two pipelines have diverged (for example a collision pass demoted a
// card cardbacklog's recompile still considers clean); it must fail the run
// loudly so the divergence is investigated, never silently mis-routed.
type reconciliation struct {
	// generatedDivergences lists ids where perCardGenerated disagrees with the
	// authoritative generated decision.
	generatedDivergences []divergence
	// exclusionConflicts lists ids our policy treated as eligible but the report
	// marked excluded.
	exclusionConflicts []divergence
	// missingFromCorpus lists report ids not seen in the streamed corpus.
	missingFromCorpus []string
}

type divergence struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	PerCard       string `json:"per_card"`
	Authoritative string `json:"authoritative"`
}

func (r reconciliation) ok() bool {
	return len(r.generatedDivergences) == 0 &&
		len(r.exclusionConflicts) == 0 &&
		len(r.missingFromCorpus) == 0
}

// applyAuthority sets each eligible card's authoritative generated flag and
// lowering summaries from compilecards' report, and reconciles that authority
// against the independent per-card recompile. It is a pure function over the
// streamed outcomes and the loaded report so the routing authority and the
// divergence guard can be unit-tested without a corpus.
func applyAuthority(outcomes []cardOutcome, authority compileAuthority) reconciliation {
	var rec reconciliation
	seen := make(map[string]bool, len(outcomes))
	for i := range outcomes {
		card := &outcomes[i]
		if !card.eligible {
			continue
		}
		seen[card.id] = true
		switch {
		case authority.excluded[card.id]:
			rec.exclusionConflicts = append(rec.exclusionConflicts, divergence{
				ID:            card.id,
				Name:          card.name,
				PerCard:       "eligible",
				Authoritative: "excluded",
			})
			card.generated = false
			card.loweringSummaries = []string{reasonReportExcluded}
		default:
			if summaries, unsupported := authority.unsupported[card.id]; unsupported {
				card.generated = false
				card.loweringSummaries = summaries
			} else {
				card.generated = true
			}
		}
		if card.generated != card.perCardGenerated {
			rec.generatedDivergences = append(rec.generatedDivergences, divergence{
				ID:            card.id,
				Name:          card.name,
				PerCard:       generatedLabel(card.perCardGenerated),
				Authoritative: generatedLabel(card.generated),
			})
		}
	}
	for id := range authority.unsupported {
		if !seen[id] {
			rec.missingFromCorpus = append(rec.missingFromCorpus, id)
		}
	}
	slices.Sort(rec.missingFromCorpus)
	return rec
}

func generatedLabel(generated bool) string {
	if generated {
		return "generated"
	}
	return "ungenerated"
}
