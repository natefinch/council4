package cards_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
)

func TestProductionManifestAdmissionsAndQueue(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../testdata/card-authoring/production.json")
	if err != nil {
		t.Fatal(err)
	}
	type workRecord struct {
		OracleID                                             string `json:"oracle_id"`
		Revision                                             string `json:"revision"`
		Name, Status, Owner, Family, Batch, Evidence, Reason string
		Blocker                                              int
	}
	type oracleEvidence struct {
		OracleID string `json:"oracle_id"`
		TextHash string `json:"oracle_text_sha256"`
	}
	type rulesEvidence struct {
		Cards []oracleEvidence
	}
	var queue struct {
		Corpus        string `json:"corpus_sha256"`
		Records       []workRecord
		Baseline      []workRecord  `json:"baseline_production"`
		RulesEvidence rulesEvidence `json:"rules_evidence"`
	}
	if err := json.Unmarshal(data, &queue); err != nil {
		t.Fatal(err)
	}
	if queue.Corpus != "6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4" ||
		len(queue.Records)+len(queue.Baseline) != 38744 {
		t.Fatal("pinned population changed without explicit reconciliation")
	}
	compiledData, err := os.ReadFile("../../testdata/card-preservation/installed-compiler.json")
	if err != nil {
		t.Fatal(err)
	}
	var compiled struct {
		Cards []workRecord
	}
	if err := json.Unmarshal(compiledData, &compiled); err != nil {
		t.Fatal(err)
	}
	compiledIDs := make(map[string]bool)
	for _, card := range compiled.Cards {
		compiledIDs[card.OracleID] = true
	}
	hashes := make(map[string]string)
	for _, evidence := range queue.RulesEvidence.Cards {
		hashes[evidence.OracleID] = evidence.TextHash
	}
	registry := cards.NewDefaultRegistry()
	seen := make(map[string]bool)
	firstBatches := make(map[string]int)
	for _, r := range append(queue.Records, queue.Baseline...) {
		if r.OracleID == "" || r.Revision == "" || seen[r.OracleID] {
			t.Fatalf("missing/duplicate pinned identity %+v", r)
		}
		seen[r.OracleID] = true
		if r.Owner != "" {
			t.Fatalf("committed manifest has a dangling claim: %s", r.OracleID)
		}
		switch r.Status {
		case "implemented-verified":
			def, err := registry.LookupOracleID(r.OracleID)
			if err != nil || def.Name != r.Name || !strings.Contains(r.Evidence, "mtg/rules/") {
				t.Fatalf("unverified production admission %s: %v", r.Name, err)
			}
			if hash := fmt.Sprintf("%x", sha256.Sum256([]byte(def.OracleText))); hashes[r.OracleID] != hash {
				t.Fatalf("registered card changed pinned Oracle text: %s", r.Name)
			}
			if err := registry.Programs().ValidateCard(def); err != nil {
				t.Fatal(err)
			}
			firstBatches[r.Batch]++
		case "blocked":
			if r.Blocker <= 0 || r.Reason == "" {
				t.Fatalf("imprecise blocked card %s", r.Name)
			}
			if _, err := registry.LookupOracleID(r.OracleID); err == nil {
				t.Fatalf("whole blocked card was partially registered: %s", r.Name)
			}
		case "installed-compiled":
			if !compiledIDs[r.OracleID] || r.Evidence != "testdata/card-preservation/installed-compiler.json" {
				t.Fatalf("compiler installation lacks checked-in source evidence: %s", r.Name)
			}
		case "unclaimed", "excluded", "":
		default:
			t.Fatalf("untracked claim/status %s: %s", r.Name, r.Status)
		}
	}
	for _, b := range []string{"01-destruction", "02-stack-interaction", "03-library-search"} {
		if firstBatches[b] != 5 {
			t.Fatalf("%s delivered %d, want five new cards", b, firstBatches[b])
		}
	}
}
