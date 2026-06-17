package main

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// routeOf returns the queue a single outcome routes into, exercising buildReport
// on a one-card corpus. The outcome must already have its authoritative fields
// set (as applyAuthority would).
func routeOf(t *testing.T, card cardOutcome) string {
	t.Helper()
	output := buildReport([]cardOutcome{card})
	switch {
	case output.SupportedCards == 1:
		return "supported"
	case output.LoweringBacklog == 1:
		return "lowering"
	case output.ParserBacklog == 1:
		return "parser"
	default:
		t.Fatalf("card routed nowhere: %+v", output)
		return ""
	}
}

func TestRoutingPartition(t *testing.T) {
	cases := []struct {
		name string
		card cardOutcome
		want string
	}{
		{
			name: "generated card is supported",
			card: cardOutcome{eligible: true, generated: true, parserComplete: true},
			want: "supported",
		},
		{
			name: "generated but parser-incomplete stays supported",
			card: cardOutcome{eligible: true, generated: true, parserComplete: false},
			want: "supported",
		},
		{
			name: "parser-complete ungenerated routes to lowering",
			card: cardOutcome{
				eligible:          true,
				parserComplete:    true,
				loweringSummaries: []string{"unsupported damage spell"},
			},
			want: "lowering",
		},
		{
			name: "parser-incomplete ungenerated routes to parser",
			card: cardOutcome{
				eligible:  true,
				uncovered: []uncoveredItem{{component: "if you do", normalized: "if you do", blocker: parser.CoverageBlockerCondition}},
			},
			want: "parser",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := routeOf(t, tc.card); got != tc.want {
				t.Fatalf("routed to %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildReportCountsAndPartition(t *testing.T) {
	outcomes := []cardOutcome{
		{index: 0, name: "Excluded", eligible: false},
		{index: 1, name: "Generated", eligible: true, generated: true, parserComplete: true},
		{index: 2, name: "GenIncomplete", eligible: true, generated: true, parserComplete: false},
		{
			index: 3, name: "Lower", eligible: true, parserComplete: true,
			loweringSummaries: []string{"unsupported damage spell"},
		},
		{
			index: 4, name: "Parse", eligible: true,
			uncovered: []uncoveredItem{{component: "if you do", normalized: "if you do", blocker: parser.CoverageBlockerCondition}},
		},
	}
	output := buildReport(outcomes)

	if output.EligibleCards != 4 {
		t.Errorf("eligible = %d, want 4", output.EligibleCards)
	}
	if output.ExcludedCards != 1 {
		t.Errorf("excluded = %d, want 1", output.ExcludedCards)
	}
	if output.SupportedCards != 2 {
		t.Errorf("supported = %d, want 2", output.SupportedCards)
	}
	if output.ParserCompleteCards != 2 {
		t.Errorf("parser-complete = %d, want 2", output.ParserCompleteCards)
	}
	if output.LoweringBacklog != 1 {
		t.Errorf("lowering backlog = %d, want 1", output.LoweringBacklog)
	}
	if output.ParserBacklog != 1 {
		t.Errorf("parser backlog = %d, want 1", output.ParserBacklog)
	}
	if output.GeneratedIncomplete != 1 {
		t.Errorf("generated-incomplete = %d, want 1", output.GeneratedIncomplete)
	}
	if !output.PartitionOK {
		t.Errorf("partition not OK: %d+%d+%d vs %d",
			output.SupportedCards, output.LoweringBacklog, output.ParserBacklog, output.EligibleCards)
	}
	if got := output.GeneratedIncompleteNames; len(got) != 1 || got[0] != "GenIncomplete" {
		t.Errorf("generated-incomplete names = %v, want [GenIncomplete]", got)
	}
}

func TestLoweringQueueBucketing(t *testing.T) {
	outcomes := []cardOutcome{
		{index: 0, name: "Solo", eligible: true, parserComplete: true, loweringSummaries: []string{"A"}},
		{index: 1, name: "Pair", eligible: true, parserComplete: true, loweringSummaries: []string{"A", "B"}},
		{index: 2, name: "Gen", eligible: true, generated: true, loweringSummaries: nil},
		{
			index: 3, name: "ParserOnly", eligible: true, loweringSummaries: []string{"A"},
			uncovered: []uncoveredItem{{component: "x", normalized: "x", blocker: parser.CoverageBlockerEffect}},
		},
	}
	output := buildReport(outcomes)

	if len(output.LoweringQueue) != 2 {
		t.Fatalf("lowering queue rows = %d, want 2: %+v", len(output.LoweringQueue), output.LoweringQueue)
	}
	top := output.LoweringQueue[0]
	if top.Reason != "A" || top.AffectedCards != 2 || top.SoleBlockers != 1 {
		t.Errorf("top row = %+v, want reason=A affected=2 sole=1", top)
	}
	second := output.LoweringQueue[1]
	if second.Reason != "B" || second.AffectedCards != 1 || second.SoleBlockers != 0 {
		t.Errorf("second row = %+v, want reason=B affected=1 sole=0", second)
	}
}

// TestLoweringQueueCollisionDemotion exercises the edge the reconciliation guard
// exists for: a card the per-card recompile thinks is clean but compilecards
// demoted with no lowering diagnostic. It must route OUT of supported and into
// the lowering queue under the synthetic collision reason, never silently
// counted as supported.
func TestLoweringQueueCollisionDemotion(t *testing.T) {
	outcomes := []cardOutcome{
		{
			index: 0, name: "Demoted", eligible: true, parserComplete: true,
			perCardGenerated:  true,
			generated:         false,
			loweringSummaries: nil,
		},
	}
	output := buildReport(outcomes)

	if output.SupportedCards != 0 {
		t.Fatalf("collision-demoted card counted as supported: %+v", output)
	}
	if output.LoweringBacklog != 1 || len(output.LoweringQueue) != 1 {
		t.Fatalf("expected 1 lowering-queue row, got backlog=%d queue=%+v",
			output.LoweringBacklog, output.LoweringQueue)
	}
	if got := output.LoweringQueue[0].Reason; got != reasonCollisionOrParseRejection {
		t.Errorf("reason = %q, want %q", got, reasonCollisionOrParseRejection)
	}
}

func TestParserQueueBucketing(t *testing.T) {
	outcomes := []cardOutcome{
		{
			index: 0, name: "Alpha", eligible: true,
			uncovered: []uncoveredItem{
				{component: "if you do", normalized: "if you do", blocker: parser.CoverageBlockerCondition},
				{component: "deal 3 damage", normalized: "deal N damage", blocker: parser.CoverageBlockerEffect},
			},
		},
		{
			index: 1, name: "Beta", eligible: true,
			uncovered: []uncoveredItem{
				{component: "if you do", normalized: "if you do", blocker: parser.CoverageBlockerCondition},
			},
		},
		{
			index: 2, name: "Gen", eligible: true, generated: true,
			uncovered: []uncoveredItem{
				{component: "if you do", normalized: "if you do", blocker: parser.CoverageBlockerCondition},
			},
		},
	}
	output := buildReport(outcomes)

	if len(output.ParserQueue) != 2 {
		t.Fatalf("parser queue rows = %d, want 2: %+v", len(output.ParserQueue), output.ParserQueue)
	}
	top := output.ParserQueue[0]
	if top.Component != string(parser.CoverageBlockerCondition) || top.Cluster != "if you do" || top.Count != 2 {
		t.Errorf("top row = %+v, want condition/if you do/count=2", top)
	}
	if len(top.Examples) != 2 || top.Examples[0] != "Alpha" || top.Examples[1] != "Beta" {
		t.Errorf("top examples = %v, want [Alpha Beta]", top.Examples)
	}
}
