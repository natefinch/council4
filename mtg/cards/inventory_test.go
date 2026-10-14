package cards_test

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
)

func TestProductionInventoryPreservesLegacyAndAuthoredEvidence(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	inventory := registry.Inventory()
	if len(inventory) != registry.Len() {
		t.Fatal("inventory lost production registrations")
	}
	for _, entry := range authored.Cards {
		records := registry.SupportFor(entry.Name)
		if len(records) != 1 || records[0].OracleID != entry.OracleID || !records[0].Loadable || !records[0].BehaviorVerified ||
			records[0].SourceRevision == "" || records[0].RulesRevision == "" {
			t.Fatalf("production evidence for %s: %+v", entry.Name, records)
		}
	}
	forest := registry.SupportFor("Forest")
	if len(forest) != 1 || forest[0].OracleID != "" || forest[0].BehaviorVerified || !forest[0].Loadable {
		t.Fatalf("legacy support promoted to exact/verified identity: %+v", forest)
	}
	entry := authored.Cards[0]
	entry.Programs = nil
	broken, err := cards.NewProgramRegistry([]cardset.Entry{entry})
	if err != nil {
		t.Fatal(err)
	}
	evidence := broken.SupportFor(entry.Name)
	if len(evidence) != 1 || evidence[0].Loadable || evidence[0].BehaviorVerified || evidence[0].Failure == "" {
		t.Fatal("deliberately missing handler produced successful-looking support")
	}
}
