package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerMassReanimationExchangeCreature(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Living Death",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Each player exiles all creature cards from their graveyard, then sacrifices all " +
			"creatures they control, then puts all cards they exiled this way onto the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	prim, ok := mode.Sequence[0].Primitive.(game.MassReanimationExchange)
	if !ok {
		t.Fatalf("primitive = %T, want game.MassReanimationExchange", mode.Sequence[0].Primitive)
	}
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("selection = %#v, want creature filter", prim.Selection)
	}
	if prim.Selection.Controller != game.ControllerAny {
		t.Fatalf("selection controller = %v, want ControllerAny", prim.Selection.Controller)
	}
}

func TestLowerMassReanimationExchangeArtifact(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Test Scrap Mastery",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Each player exiles all artifact cards from their graveyard, then sacrifices all " +
			"artifacts they control, then puts all cards they exiled this way onto the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	prim, ok := mode.Sequence[0].Primitive.(game.MassReanimationExchange)
	if !ok {
		t.Fatalf("primitive = %T, want game.MassReanimationExchange", mode.Sequence[0].Primitive)
	}
	if !slices.Equal(prim.Selection.RequiredTypes, []types.Card{types.Artifact}) {
		t.Fatalf("selection = %#v, want artifact filter", prim.Selection)
	}
}
