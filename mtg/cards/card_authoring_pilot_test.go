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
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
)

type frozenPilotCard struct {
	Name, Revision, Family, Baseline string
	OracleID                         string `json:"oracle_id"`
	TextHash                         string `json:"oracle_text_sha256"`
}

type pilotOutcome struct {
	Status, Evidence, Reason string
	Blocker                  int
}

type pilotBatchEvidence struct {
	Operations []string `json:"reused_operations"`
	Files      []string `json:"changed_files"`
	Scenarios  []string
}

type cardAuthoringPilot struct {
	Corpus   string            `json:"corpus_sha256"`
	Frozen   []frozenPilotCard `json:"frozen_cards"`
	Outcomes map[string]pilotOutcome
	Batches  map[string]pilotBatchEvidence
}

type laterProductionAdmission struct {
	OracleID         string `json:"oracle_id"`
	Status, Evidence string
}

func TestCardAuthoringFrozenPilotAdmissionsAndBlockers(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../testdata/card-authoring/pilot.json")
	if err != nil {
		t.Fatal(err)
	}
	var pilot cardAuthoringPilot
	if err := json.Unmarshal(data, &pilot); err != nil {
		t.Fatal(err)
	}
	if pilot.Corpus != "6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4" ||
		len(pilot.Frozen) != 20 || len(pilot.Outcomes) != 20 {
		t.Fatal("pilot pin or frozen population changed")
	}
	productionData, err := os.ReadFile("../../testdata/card-authoring/production.json")
	if err != nil {
		t.Fatal(err)
	}
	var production struct {
		Records []laterProductionAdmission
	}
	if err := json.Unmarshal(productionData, &production); err != nil {
		t.Fatal(err)
	}
	laterAdmissions := make(map[string]bool)
	for _, record := range production.Records {
		laterAdmissions[record.OracleID] = record.Status == "implemented-verified" && strings.Contains(record.Evidence, "mtg/rules/")
	}
	challenges := []string{"Necromancy", "Chains of Mephistopheles", "Humility", "Opalescence", "Possibility Storm", "Doomsday", "Panglacial Wurm", "Warp World"}
	registry := cards.NewDefaultRegistry()
	seen := make(map[string]bool)
	families := make(map[string]bool)
	implementedFamilies := make(map[string]bool)
	implemented := 0
	challengeCount := 0
	for _, card := range pilot.Frozen {
		if seen[card.OracleID] || card.OracleID == "" || card.Revision == "" || len(card.TextHash) != 64 ||
			card.Baseline != "unsupported" {
			t.Fatalf("invalid frozen identity: %+v", card)
		}
		seen[card.OracleID] = true
		if card.Family == "challenge" {
			if !slices.Contains(challenges, card.Name) {
				t.Fatalf("held-out challenge substituted: %s", card.Name)
			}
			challengeCount++
		} else {
			families[card.Family] = true
		}
		outcome, exists := pilot.Outcomes[card.Name]
		if !exists {
			t.Fatalf("missing outcome: %s", card.Name)
		}
		def, lookupErr := registry.LookupOracleID(card.OracleID)
		switch outcome.Status {
		case "implemented":
			implemented++
			implementedFamilies[card.Family] = true
			if lookupErr != nil || def == nil || def.Name != card.Name {
				t.Fatalf("%s is not exact-identity production support: %v", card.Name, lookupErr)
			}
			if err := registry.Programs().ValidateCard(def); err != nil {
				t.Fatal(err)
			}
			index := slices.IndexFunc(authored.Cards, func(entry cardset.Entry) bool { return entry.OracleID == card.OracleID })
			if index < 0 || !strings.Contains(authored.Cards[index].Verification, "mtg/rules/") {
				t.Fatalf("missing runtime evidence: %s", card.Name)
			}
			batch, exists := pilot.Batches[outcome.Evidence]
			if !exists || len(batch.Operations) == 0 || len(batch.Files) == 0 || len(batch.Scenarios) == 0 {
				t.Fatalf("incomplete batch evidence: %s", card.Name)
			}
		case "blocked":
			if outcome.Blocker == 0 || outcome.Reason == "" {
				t.Fatalf("partial promotion or imprecise blocker: %s", card.Name)
			}
			// Historical pilot outcomes remain frozen; subsequent exact admissions
			// require current production evidence rather than rewriting that pilot.
			if lookupErr == nil {
				if !laterAdmissions[card.OracleID] || def == nil || def.Name != card.Name ||
					fmt.Sprintf("%x", sha256.Sum256([]byte(def.OracleText))) != card.TextHash {
					t.Fatalf("unverified post-pilot admission: %s", card.Name)
				}
				if err := registry.Programs().ValidateCard(def); err != nil {
					t.Fatal(err)
				}
			} else if len(registry.LookupAll(card.Name)) != 0 {
				t.Fatalf("partial promotion of historical blocker: %s", card.Name)
			}
		default:
			t.Fatalf("unclassified card %s: %q", card.Name, outcome.Status)
		}
	}
	if challengeCount != 8 || len(families) < 3 || implemented < 10 || len(implementedFamilies) < 3 {
		t.Fatalf("pilot threshold not satisfied: challenges=%d neighbor families=%d implemented=%d", challengeCount, len(families), implemented)
	}
}
