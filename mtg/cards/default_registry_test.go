package cards_test

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
)

func TestDefaultRegistryChecksToothCollectorOracleIdentity(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("b7c165a4-53a2-4117-a452-b9a8fe309c09")
	if err != nil {
		t.Fatal(err)
	}
	if def.Name != "Tooth Collector" || len(def.TriggeredAbilities) != 2 {
		t.Fatalf("registered Tooth Collector = %q with %d triggers, want exact identity with both abilities", def.Name, len(def.TriggeredAbilities))
	}
}

func TestNewDefaultRegistryResolvesKnownCard(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	if registry.Len() == 0 {
		t.Fatal("NewDefaultRegistry returned an empty registry")
	}
	if registry.Lookup("Anger") == nil {
		t.Error(`Lookup("Anger") = nil, want a known committed card`)
	}
	for _, name := range registry.All() {
		if registry.Lookup(name) == nil {
			t.Errorf("Lookup(%q) = nil for a registered card name", name)
		}
	}
}

func TestDefaultCardSetsCountMatchesRegistry(t *testing.T) {
	total := 0
	for _, set := range cards.DefaultCardSets() {
		for _, entry := range set {
			if !entry.ReplaceLegacy {
				total++
			}
		}
	}
	if got := cards.NewDefaultRegistry().Len(); got != total {
		t.Errorf("registry Len() = %d, want %d (sum of DefaultCardSets)", got, total)
	}
}
