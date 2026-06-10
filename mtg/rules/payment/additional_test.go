package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestAdditionalCostSourceZone(t *testing.T) {
	tests := []struct {
		name   string
		source zone.Type
		want   zone.Type
	}{
		{name: "default is graveyard", source: zone.None, want: zone.Graveyard},
		{name: "explicit graveyard", source: zone.Graveyard, want: zone.Graveyard},
		{name: "hand", source: zone.Hand, want: zone.Hand},
		{name: "library", source: zone.Library, want: zone.Library},
		{name: "exile", source: zone.Exile, want: zone.Exile},
		{name: "command", source: zone.Command, want: zone.Command},
		{name: "unknown is unchanged", source: zone.Type(99), want: zone.Type(99)},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := additionalCostSourceZone(test.source); got != test.want {
				t.Fatalf("additionalCostSourceZone(%d) = %v, want %v", test.source, got, test.want)
			}
		})
	}
}

func TestAdditionalCostMatchesAnyCardSubtype(t *testing.T) {
	additional := cost.Additional{
		Kind:        cost.AdditionalReveal,
		SubtypesAny: []types.Sub{types.Forest, types.Mountain},
	}
	forest := &game.CardDef{CardFace: game.CardFace{
		Types:    []types.Card{types.Land},
		Subtypes: []types.Sub{types.Forest},
	}}
	if !additionalCostMatchesCard(forest, additional) {
		t.Fatal("Forest did not match Forest-or-Mountain reveal cost")
	}
	creature := &game.CardDef{CardFace: game.CardFace{
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}}
	if additionalCostMatchesCard(creature, additional) {
		t.Fatal("Elf matched Forest-or-Mountain reveal cost")
	}
}
