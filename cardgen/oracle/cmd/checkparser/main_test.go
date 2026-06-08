package main

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/internal/corpuscheck"
)

func TestCheckParserReportsSyntaxDiagnostics(t *testing.T) {
	t.Parallel()
	input := `[
		{"id":"one","name":"Good","type_line":"Creature","oracle_text":"Flying"},
		{"id":"two","name":"Bad","type_line":"Creature","oracle_text":"Flying)"},
		{"id":"three","name":"Modal","type_line":"Sorcery","oracle_text":"Choose one —"}
	]`
	report, err := corpuscheck.Check(strings.NewReader(input), 3, checkParser)
	if err != nil {
		t.Fatal(err)
	}
	if report.CardCount != 3 || report.OracleTextCount != 3 || report.UnsupportedCount != 2 {
		t.Fatalf("report = %#v", report)
	}
	if got := report.Unsupported[0].Issues[0].Reason; got != "unmatched parenthesis" {
		t.Fatalf("first reason = %q", got)
	}
	if got := report.Unsupported[1].Issues[0].Reason; got != "modal ability has no options" {
		t.Fatalf("second reason = %q", got)
	}
}

func TestHasCardType(t *testing.T) {
	t.Parallel()
	if !hasCardType("Legendary Planeswalker — Jace", "Planeswalker") {
		t.Fatal("did not find planeswalker type")
	}
	if hasCardType("Creature — Shapeshifter", "Planeswalker") {
		t.Fatal("found absent planeswalker type")
	}
}
