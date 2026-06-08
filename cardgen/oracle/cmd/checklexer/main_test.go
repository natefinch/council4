package main

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/internal/corpuscheck"
)

func TestCheckerReportsUnsupportedTextsInInputOrder(t *testing.T) {
	t.Parallel()
	input := `[
		{
			"id": "one",
			"oracle_id": "oracle-one",
			"name": "Supported",
			"set": "tst",
			"collector_number": "1",
			"oracle_text": "{T}: Add {G}."
		},
		{
			"id": "two",
			"name": "Broken Faces",
			"set": "tst",
			"collector_number": "2",
			"card_faces": [
				{"name": "Front", "oracle_id": "front", "oracle_text": "Bad {T"},
				{"name": "Back", "oracle_id": "back", "oracle_text": "NUL \u0000 here"}
			]
		},
		{
			"id": "three",
			"name": "Broken Root",
			"oracle_text": "Middle \uFEFF BOM"
		}
	]`
	report, err := corpuscheck.Check(strings.NewReader(input), 4, checkLexer)
	if err != nil {
		t.Fatal(err)
	}
	if report.CardCount != 3 || report.OracleTextCount != 4 || report.UnsupportedCount != 3 {
		t.Fatalf("report counts = %#v", report)
	}
	got := report.Unsupported
	if got[0].ID != "two" || got[0].FaceName != "Front" ||
		got[0].Issues[0].Reason != "unclosed braced symbol" {
		t.Fatalf("first unsupported = %#v", got[0])
	}
	if got[1].ID != "two" || got[1].FaceName != "Back" ||
		got[1].Issues[0].Reason != "NUL is not valid in Oracle text" {
		t.Fatalf("second unsupported = %#v", got[1])
	}
	if got[2].ID != "three" ||
		got[2].Issues[0].Reason != "a UTF-8 BOM is only valid at the start of Oracle text" {
		t.Fatalf("third unsupported = %#v", got[2])
	}
}

func TestCheckerRejectsInvalidInput(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"object":    `{}`,
		"truncated": `[{"name":"broken"}`,
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if _, err := corpuscheck.Check(strings.NewReader(input), 2, checkLexer); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
