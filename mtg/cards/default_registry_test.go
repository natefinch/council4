package cards_test

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
)

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
		total += len(set)
	}
	if got := cards.NewDefaultRegistry().Len(); got != total {
		t.Errorf("registry Len() = %d, want %d (sum of DefaultCardSets)", got, total)
	}
}
