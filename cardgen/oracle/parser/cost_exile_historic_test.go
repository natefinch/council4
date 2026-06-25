package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/zone"
)

// TestParseExileAnyNumberHistoricThresholdCost asserts the variable-cardinality
// exile cost "Exile any number of historic cards from your graveyard with total
// mana value 30 or greater" decomposes into typed fields: any-number cardinality,
// a historic object filter, the graveyard source zone, and the aggregate
// mana-value threshold.
func TestParseExileAnyNumberHistoricThresholdCost(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t,
		"Exile any number of historic cards from your graveyard with total mana value 30 or greater: Draw a card.")
	if component.Kind != CostComponentExile {
		t.Fatalf("kind = %v, want exile", component.Kind)
	}
	if !component.AnyNumber {
		t.Fatal("AnyNumber = false, want true")
	}
	if !component.ObjectHistoric {
		t.Fatal("ObjectHistoric = false, want true")
	}
	if component.SourceZone != zone.Graveyard {
		t.Fatalf("SourceZone = %v, want graveyard", component.SourceZone)
	}
	if component.TotalManaValueAtLeast != 30 {
		t.Fatalf("TotalManaValueAtLeast = %d, want 30", component.TotalManaValueAtLeast)
	}
}

// TestParseGetEmblemEffect asserts "You get an emblem with \"...\"" reclassifies
// to an EffectCreateEmblem carrying each quoted ability parsed through the same
// pipeline, rather than being mis-modeled as another effect kind.
func TestParseGetEmblemEffect(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(
		"You get an emblem with \"Creatures you control have base power and toughness 9/9.\"",
		Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := document.Abilities[0]
	if len(ability.Sentences) != 1 || len(ability.Sentences[0].Effects) != 1 {
		t.Fatalf("sentences = %#v, want one effect", ability.Sentences)
	}
	effect := ability.Sentences[0].Effects[0]
	if effect.Kind != EffectCreateEmblem {
		t.Fatalf("kind = %v, want EffectCreateEmblem", effect.Kind)
	}
	if !effect.Exact {
		t.Fatal("Exact = false, want true")
	}
	if len(effect.EmblemAbilities) != 1 {
		t.Fatalf("EmblemAbilities = %d, want 1", len(effect.EmblemAbilities))
	}
}
