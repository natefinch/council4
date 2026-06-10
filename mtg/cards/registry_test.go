package cards

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestRegistrySupportsDistinctDefinitionsWithSameName(t *testing.T) {
	first := &game.CardDef{CardFace: game.CardFace{Name: "Pirate"}}
	second := &game.CardDef{CardFace: game.CardFace{Name: "Pirate"}}

	registry := NewRegistry([]*game.CardDef{first, second})

	if got := registry.Lookup("Pirate"); got != first {
		t.Fatalf("Lookup returned %p, want first definition %p", got, first)
	}
	matches := registry.LookupAll("Pirate")
	if len(matches) != 2 || matches[0] != first || matches[1] != second {
		t.Fatalf("LookupAll returned %#v", matches)
	}
	if registry.Len() != 2 {
		t.Fatalf("Len = %d, want 2", registry.Len())
	}
}
