package main

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
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

func TestDiacriticNameDraftUsesASCIISourceNames(t *testing.T) {
	stem, symbol, err := draftNames("Lim-Dûl's Vault", "")
	if err != nil || stem != "lim_dul_s_vault" || symbol != "LimDulSVault" {
		t.Fatalf("draftNames = %q, %q, %v", stem, symbol, err)
	}
	if _, chosen, err := draftNames("Lim-Dûl's Vault", "Vault"); err != nil || chosen != "Vault" {
		t.Fatalf("explicit symbol = %q, %v", chosen, err)
	}
	card := oracleCard{
		ID: "printing-revision", OracleID: "oracle-id", Name: "Lim-Dûl's Vault",
		Layout: "normal", ManaCost: "{U}{B}", TypeLine: "Instant", Colors: []string{"U", "B"},
		ColorIdentity: []string{"U", "B"},
	}
	body, scenario, err := render(card, symbol)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `Name: "Lim-Dûl's Vault"`) || !strings.Contains(string(scenario), "TestAuthoredLimDulSVaultScenario") {
		t.Fatal("draft must keep the exact printed name and use the ASCII symbol")
	}
}

// TestExplicitSymbolsKeepWorkingForAnyPrintedName drafts real corpus records:
// an explicit symbol is any exported Go identifier, including a Unicode one,
// and names the draft files when the printed name has no ASCII spelling.
func TestExplicitSymbolsKeepWorkingForAnyPrintedName(t *testing.T) {
	dir := t.TempDir()
	corpus := filepath.Join(dir, "oracle.json")
	records := []oracleCard{}
	for i, name := range []string{"Séance", "_____", "Ψυχή"} {
		records = append(records, oracleCard{
			ID: fmt.Sprint("printing-", i), OracleID: fmt.Sprint("oracle-", i), Name: name,
			Layout: "normal", ManaCost: "{W}", TypeLine: "Instant", Colors: []string{"W"},
			ColorIdentity: []string{"W"},
		})
	}
	data, err := json.Marshal(records)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corpus, data, 0o600); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
	for _, tc := range []struct{ oracleID, symbol, name, stem, entry string }{
		{"oracle-0", "Séance", "Séance", "seance", "Séance"},
		{"oracle-0", "", "Séance", "seance", "Seance"},
		{"oracle-1", "BlankCreature", "_____", "blank_creature", "BlankCreature"},
		{"oracle-2", "Ψυχή", "Ψυχή", "card", "Ψυχή"},
	} {
		out := filepath.Join(".cardwork", "scaffold", tc.oracleID+tc.symbol)
		if err := run(corpus, tc.oracleID, tc.symbol, out); err != nil {
			t.Fatalf("%q with symbol %q: %v", tc.name, tc.symbol, err)
		}
		body, err := os.ReadFile(filepath.Join(out, tc.stem+".go"))
		if err != nil {
			t.Fatal(err)
		}
		scenario, err := os.ReadFile(filepath.Join(out, tc.stem+"_test.go"))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), fmt.Sprintf("Name: %q", tc.name)) ||
			!strings.Contains(string(body), "var "+tc.entry+" = cardset.Entry") ||
			!strings.Contains(string(scenario), "TestAuthored"+tc.entry+"Scenario") {
			t.Fatalf("%q draft must keep the exact name and symbol %q", tc.name, tc.entry)
		}
	}
	for _, oracleID := range []string{"oracle-1", "oracle-2"} {
		if err := run(corpus, oracleID, "", filepath.Join(".cardwork", "scaffold", "default-"+oracleID)); err == nil {
			t.Fatalf("%s without an ASCII name derived a default symbol", oracleID)
		}
	}
}
