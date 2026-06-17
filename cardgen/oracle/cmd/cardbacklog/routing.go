package main

import (
	"cmp"
	"slices"
)

const (
	exampleLimit             = 5
	generatedIncompleteLimit = 50

	// reasonReportExcluded labels a card our policy treated as eligible but
	// compilecards' report marked excluded (an exclusion conflict).
	reasonReportExcluded = "excluded by compile report"
	// reasonCollisionOrParseRejection labels an authoritatively-ungenerated card
	// that carries no diagnostic summary, which happens when compilecards demoted
	// it through a corpus-wide collision or parse-rejection pass rather than a
	// lowering diagnostic. It keeps such cards in the lowering queue rather than
	// dropping them silently.
	reasonCollisionOrParseRejection = "generated collision or parse rejection"
)

// buildReport routes every eligible card to exactly one bucket and builds the two
// ranked queues. The routing is a strict partition of the eligible (non-excluded)
// cards: a card is generated (supported), or it is parser-complete-but-ungenerated
// (lowering backlog), or it is parser-incomplete-and-ungenerated (parser backlog).
// Generated cards that are not parser-complete are a small residue that stays in
// the supported bucket and is reported separately.
func buildReport(outcomes []cardOutcome) report {
	output := report{TotalCards: len(outcomes)}

	loweringBucket := newLoweringBuckets()
	parserBucket := newParserBuckets()

	for i := range outcomes {
		card := &outcomes[i]
		if !card.eligible {
			output.ExcludedCards++
			continue
		}
		output.EligibleCards++
		if card.parserComplete {
			output.ParserCompleteCards++
		}
		switch {
		case card.generated:
			output.SupportedCards++
			if !card.parserComplete {
				output.GeneratedIncomplete++
				if len(output.GeneratedIncompleteNames) < generatedIncompleteLimit {
					output.GeneratedIncompleteNames = append(output.GeneratedIncompleteNames, card.name)
				}
			}
		case card.parserComplete:
			output.LoweringBacklog++
			loweringBucket.add(card)
		default:
			output.ParserBacklog++
			parserBucket.add(card)
		}
	}

	output.LoweringQueue = loweringBucket.rank()
	output.ParserQueue = parserBucket.rank()
	output.PartitionOK =
		output.SupportedCards+output.LoweringBacklog+output.ParserBacklog == output.EligibleCards
	return output
}

// loweringBuckets tallies, over the lowering-backlog subset, how many
// parser-complete-but-ungenerated cards each distinct lowering diagnostic summary
// blocks, and how often it is a card's sole blocker. This is unsupported-reasons
// restricted to the ready-to-lower cards.
type loweringBuckets struct {
	reasons map[string]*loweringTally
}

type loweringTally struct {
	affected     int
	soleBlockers int
	examples     []string
}

func newLoweringBuckets() loweringBuckets {
	return loweringBuckets{reasons: make(map[string]*loweringTally)}
}

func (b loweringBuckets) add(card *cardOutcome) {
	summaries := card.loweringSummaries
	if len(summaries) == 0 {
		summaries = []string{reasonCollisionOrParseRejection}
	}
	sole := len(summaries) == 1
	for _, summary := range summaries {
		tally := b.reasons[summary]
		if tally == nil {
			tally = &loweringTally{}
			b.reasons[summary] = tally
		}
		tally.affected++
		if sole {
			tally.soleBlockers++
		}
		if len(tally.examples) < exampleLimit {
			tally.examples = append(tally.examples, card.name)
		}
	}
}

func (b loweringBuckets) rank() []loweringReason {
	rows := make([]loweringReason, 0, len(b.reasons))
	for summary, tally := range b.reasons {
		rows = append(rows, loweringReason{
			Reason:        summary,
			AffectedCards: tally.affected,
			SoleBlockers:  tally.soleBlockers,
			Examples:      tally.examples,
		})
	}
	slices.SortFunc(rows, func(a, b loweringReason) int {
		if a.AffectedCards != b.AffectedCards {
			return cmp.Compare(b.AffectedCards, a.AffectedCards)
		}
		if a.SoleBlockers != b.SoleBlockers {
			return cmp.Compare(b.SoleBlockers, a.SoleBlockers)
		}
		return cmp.Compare(a.Reason, b.Reason)
	})
	return rows
}

// parserBuckets tallies, over the parser-backlog subset, occurrences of each
// (owning component family, normalized uncovered-span cluster) pair so the
// grammar-recognition backlog can be ranked by occurrence, mirroring the
// parser-coverage work queue.
type parserBuckets struct {
	clusters map[parserKey]*parserTally
}

type parserKey struct {
	blocker    string
	normalized string
}

type parserTally struct {
	count       int
	examples    []string
	seenExample map[string]bool
}

func newParserBuckets() parserBuckets {
	return parserBuckets{clusters: make(map[parserKey]*parserTally)}
}

func (b parserBuckets) add(card *cardOutcome) {
	for i := range card.uncovered {
		item := card.uncovered[i]
		key := parserKey{blocker: string(item.blocker), normalized: item.normalized}
		tally := b.clusters[key]
		if tally == nil {
			tally = &parserTally{seenExample: make(map[string]bool)}
			b.clusters[key] = tally
		}
		tally.count++
		if !tally.seenExample[card.name] && len(tally.examples) < exampleLimit {
			tally.examples = append(tally.examples, card.name)
			tally.seenExample[card.name] = true
		}
	}
}

func (b parserBuckets) rank() []parserClusterRow {
	rows := make([]parserClusterRow, 0, len(b.clusters))
	for key, tally := range b.clusters {
		rows = append(rows, parserClusterRow{
			Component: key.blocker,
			Cluster:   key.normalized,
			Count:     tally.count,
			Examples:  tally.examples,
		})
	}
	slices.SortFunc(rows, func(a, b parserClusterRow) int {
		if a.Count != b.Count {
			return cmp.Compare(b.Count, a.Count)
		}
		if a.Component != b.Component {
			return cmp.Compare(a.Component, b.Component)
		}
		return cmp.Compare(a.Cluster, b.Cluster)
	})
	return rows
}
