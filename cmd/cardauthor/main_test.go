package main

import (
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

func TestOrdinaryDraftAndRealScenario(t *testing.T) {
	card := oracleCard{
		ID: "printing-revision", OracleID: "oracle-id", Name: "Example",
		Layout: "normal", ManaCost: "{2}{W}", TypeLine: "Instant",
		OracleText: "Destroy target permanent.", Colors: []string{"W"}, ColorIdentity: []string{"W"},
	}
	body, scenario, err := render(card, "Example")
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range [][]byte{body, scenario} {
		if _, err := parser.ParseFile(token.NewFileSet(), "draft.go", source, parser.AllErrors); err != nil {
			t.Fatal(err)
		}
	}
	for _, text := range []string{"printing-revision", `OracleID: "oracle-id"`, "cost.O(2), cost.W", "ExampleProgram", "unsupported draft"} {
		if !strings.Contains(string(body), text) {
			t.Fatalf("missing %q", text)
		}
	}
	for _, text := range []string{"NewEngineWithPrograms", "addImplementationSpellToStack", "resolveTopOfStack", "t.Fatal"} {
		if !strings.Contains(string(scenario), text) {
			t.Fatalf("scenario missing %q", text)
		}
	}
	if strings.Contains(string(body), "Cards =") || strings.Contains(string(scenario), "t.Skip") {
		t.Fatal("draft must not promote or silently pass")
	}
}

func TestScaffoldFailsClosed(t *testing.T) {
	for _, card := range []oracleCard{
		{Layout: "transform"},
		{Layout: "normal", ManaCost: "{bad}"},
		{Layout: "normal", Power: "*"},
		{Layout: "normal", Colors: []string{"bad"}},
	} {
		if _, _, err := render(card, "Example"); err == nil {
			t.Fatalf("accepted unsupported metadata: %+v", card)
		}
	}
	for _, input := range []struct{ oracleID, symbol, out string }{
		{"", "Example", ".cardwork/scaffold/example"},
		{"uuid", "lowercase", ".cardwork/scaffold/example"},
		{"uuid", "Example", "mtg/cards/authored"},
		{"uuid", "Example", "../outside"},
	} {
		if err := run("does-not-exist", input.oracleID, input.symbol, input.out); err == nil ||
			strings.Contains(err.Error(), "does-not-exist") {
			t.Fatalf("bad input reached filesystem: %v", err)
		}
	}
}
