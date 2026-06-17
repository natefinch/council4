package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

type report struct {
	EligibleCards     int               `json:"eligible_cards"`
	ParserComplete    int               `json:"parser_complete_cards"`
	CompletePercent   float64           `json:"parser_complete_percent"`
	CardExact         int               `json:"card_exact_cards"`
	CardExactPercent  float64           `json:"card_exact_percent"`
	ResolvingEffects  int               `json:"resolving_effects"`
	ExactEffects      int               `json:"exact_effects"`
	ExactPercent      float64           `json:"effect_exact_percent"`
	UncoveredClusters []cluster         `json:"uncovered_clusters"`
	BlockerSummary    []blockerCount    `json:"blocker_summary"`
	Validation        *validationResult `json:"generated_subset_validation,omitempty"`
}

type cluster struct {
	Text     string                 `json:"text"`
	Count    int                    `json:"count"`
	Blocker  parser.CoverageBlocker `json:"blocker"`
	Examples []string               `json:"examples"`
}

type blockerCount struct {
	Blocker parser.CoverageBlocker `json:"blocker"`
	Count   int                    `json:"count"`
}

type validationResult struct {
	GeneratedCards int      `json:"generated_cards"`
	Violations     int      `json:"violations"`
	ViolationNames []string `json:"violation_names,omitempty"`
}

const clusterExampleLimit = 5

func buildReport(cards []cardResult) report {
	var output report
	clusters := map[string]*cluster{}
	blockerTotals := map[parser.CoverageBlocker]int{}
	for i := range cards {
		card := cards[i]
		if !card.eligible {
			continue
		}
		output.EligibleCards++
		if card.complete {
			output.ParserComplete++
			if card.exact == card.resolving {
				output.CardExact++
			}
		}
		output.ResolvingEffects += card.resolving
		output.ExactEffects += card.exact
		for _, blocker := range card.blockers {
			blockerTotals[blocker]++
		}
		accumulateClusters(clusters, card)
	}

	output.UncoveredClusters = rankClusters(clusters)
	output.BlockerSummary = rankBlockers(blockerTotals)
	output.CompletePercent = percent(output.ParserComplete, output.EligibleCards)
	output.CardExactPercent = percent(output.CardExact, output.EligibleCards)
	output.ExactPercent = percent(output.ExactEffects, output.ResolvingEffects)
	return output
}

func accumulateClusters(clusters map[string]*cluster, card cardResult) {
	seenExample := map[string]bool{}
	for _, item := range card.uncovered {
		entry, ok := clusters[item.normalized]
		if !ok {
			entry = &cluster{Text: item.normalized, Blocker: item.blocker}
			clusters[item.normalized] = entry
		}
		entry.Count++
		if !seenExample[item.normalized] && len(entry.Examples) < clusterExampleLimit {
			entry.Examples = append(entry.Examples, card.name)
			seenExample[item.normalized] = true
		}
	}
}

func rankClusters(clusters map[string]*cluster) []cluster {
	ranked := make([]cluster, 0, len(clusters))
	for _, entry := range clusters {
		ranked = append(ranked, *entry)
	}
	slices.SortFunc(ranked, func(a, b cluster) int {
		if a.Count != b.Count {
			return cmp.Compare(b.Count, a.Count)
		}
		return cmp.Compare(a.Text, b.Text)
	})
	return ranked
}

func rankBlockers(totals map[parser.CoverageBlocker]int) []blockerCount {
	ranked := make([]blockerCount, 0, len(totals))
	for blocker, count := range totals {
		ranked = append(ranked, blockerCount{Blocker: blocker, Count: count})
	}
	slices.SortFunc(ranked, func(a, b blockerCount) int {
		if a.Count != b.Count {
			return cmp.Compare(b.Count, a.Count)
		}
		return cmp.Compare(a.Blocker, b.Blocker)
	})
	return ranked
}

func percent(part, whole int) float64 {
	if whole == 0 {
		return 0
	}
	return float64(part) / float64(whole) * 100
}

func writeReport(path string, output report) error {
	encoded, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding report: %w", err)
	}
	encoded = append(encoded, '\n')
	if path == "-" {
		_, err := os.Stdout.Write(encoded)
		return err
	}
	if err := os.WriteFile(path, encoded, 0o600); err != nil {
		return fmt.Errorf("writing report: %w", err)
	}
	return nil
}
