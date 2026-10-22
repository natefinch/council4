package cards_test

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
)

type productionWorkRecord struct {
	OracleID                                             string `json:"oracle_id"`
	Revision                                             string `json:"revision"`
	Name, Status, Owner, Family, Batch, Evidence, Reason string
	Blocker                                              int
}

type productionBatch struct {
	ID        string
	Family    string
	OracleIDs []string `json:"oracle_ids"`
}

func validateProductionClaims(records, baseline []productionWorkRecord, batches []productionBatch) error {
	membership := make(map[string]string)
	byBatch := make(map[string]productionBatch)
	for _, b := range batches {
		if _, exists := byBatch[b.ID]; exists || b.ID == "" || b.Family == "" || len(b.OracleIDs) < 3 || len(b.OracleIDs) > 5 {
			return fmt.Errorf("invalid saved batch %s", b.ID)
		}
		byBatch[b.ID] = b
		for _, id := range b.OracleIDs {
			if _, exists := membership[id]; exists || id == "" {
				return fmt.Errorf("duplicate/missing saved batch identity %s", id)
			}
			membership[id] = b.ID
		}
	}
	activeBatch, owner := "", ""
	byID := make(map[string]productionWorkRecord)
	for _, r := range records {
		if _, exists := byID[r.OracleID]; exists || r.OracleID == "" || r.Revision == "" || r.Name == "" {
			return fmt.Errorf("duplicate/missing pinned identity %s", r.OracleID)
		}
		byID[r.OracleID] = r
		if r.Batch != membership[r.OracleID] || r.Batch != "" && r.Family != byBatch[r.Batch].Family {
			return fmt.Errorf("dangling saved batch identity %s", r.OracleID)
		}
		if r.Status != "claimed" {
			if r.Owner != "" {
				return fmt.Errorf("dangling owner on %s", r.OracleID)
			}
			continue
		}
		if strings.TrimSpace(r.Owner) == "" || r.Batch == "" {
			return fmt.Errorf("untracked claim %s", r.OracleID)
		}
		if activeBatch != "" && (activeBatch != r.Batch || owner != r.Owner) {
			return fmt.Errorf("multiple claimed batches or owners")
		}
		activeBatch, owner = r.Batch, r.Owner
	}
	for id, batchID := range membership {
		if _, exists := byID[id]; !exists {
			return fmt.Errorf("missing record %s for saved batch %s", id, batchID)
		}
	}
	if activeBatch != "" {
		for _, id := range byBatch[activeBatch].OracleIDs {
			if r := byID[id]; r.Status != "claimed" || r.Owner != owner {
				return fmt.Errorf("partial saved claim %s", activeBatch)
			}
		}
	}
	for _, r := range baseline {
		if r.Owner != "" || r.Batch != "" || r.Status != "" {
			return fmt.Errorf("claim/status on baseline association %s", r.OracleID)
		}
	}
	return nil
}

func TestProductionManifestAdmissionsAndQueue(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../testdata/card-authoring/production.json")
	if err != nil {
		t.Fatal(err)
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
		Records       []productionWorkRecord
		Baseline      []productionWorkRecord `json:"baseline_production"`
		Batches       []productionBatch
		RulesEvidence rulesEvidence `json:"rules_evidence"`
	}
	if err := json.Unmarshal(data, &queue); err != nil {
		t.Fatal(err)
	}
	if queue.Corpus != "6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4" ||
		len(queue.Records)+len(queue.Baseline) != 38744 {
		t.Fatal("pinned population changed without explicit reconciliation")
	}
	if err := validateProductionClaims(queue.Records, queue.Baseline, queue.Batches); err != nil {
		t.Fatal(err)
	}
	compiledData, err := os.ReadFile("../../testdata/card-preservation/installed-compiler.json")
	if err != nil {
		t.Fatal(err)
	}
	var compiled struct {
		Cards []productionWorkRecord
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
		case "claimed", "unclaimed", "excluded", "":
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

func TestProductionManifestClaimOwnership(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{
		"coherent", "no active claim", "dangling owner", "missing owner", "blank owner", "partial", "multiple owners",
		"multiple batches", "unknown batch", "wrong member", "wrong family", "duplicate member",
		"missing record", "duplicate record", "baseline claim",
	} {
		t.Run(scenario, func(t *testing.T) {
			batches := []productionBatch{
				{ID: "first", Family: "family", OracleIDs: []string{"a", "b", "c"}},
				{ID: "second", Family: "other", OracleIDs: []string{"d", "e", "f"}},
			}
			var records []productionWorkRecord
			for _, b := range batches {
				for _, id := range b.OracleIDs {
					r := productionWorkRecord{OracleID: id, Revision: "printing-" + id, Name: id, Batch: b.ID, Family: b.Family, Status: "unclaimed"}
					if b.ID == "first" {
						r.Status, r.Owner = "claimed", "writer"
					}
					records = append(records, r)
				}
			}
			baseline := []productionWorkRecord{{OracleID: "legacy", Revision: "printing-legacy", Name: "Legacy"}}
			switch scenario {
			case "coherent":
			case "no active claim":
				for i := range records {
					records[i].Status, records[i].Owner = "unclaimed", ""
				}
			case "dangling owner":
				records[3].Owner = "writer"
			case "missing owner":
				records[0].Owner = ""
			case "blank owner":
				for i := 0; i < 3; i++ {
					records[i].Owner = " "
				}
			case "partial":
				records[1].Status, records[1].Owner = "unclaimed", ""
			case "multiple owners":
				records[1].Owner = "other-writer"
			case "multiple batches":
				for i := 3; i < len(records); i++ {
					records[i].Status, records[i].Owner = "claimed", "writer"
				}
			case "unknown batch":
				records[0].Batch = "unknown"
			case "wrong member":
				batches[0].OracleIDs[0] = "missing"
			case "wrong family":
				records[0].Family = "other"
			case "duplicate member":
				batches[0].OracleIDs[1] = "a"
			case "missing record":
				records = slices.Delete(records, 0, 1)
			case "duplicate record":
				records = append(records, records[0])
			case "baseline claim":
				baseline[0].Status, baseline[0].Owner = "claimed", "writer"
			}
			err := validateProductionClaims(records, baseline, batches)
			valid := scenario == "coherent" || scenario == "no active claim"
			if (err == nil) != valid {
				t.Fatalf("%s: %v", scenario, err)
			}
		})
	}
}
