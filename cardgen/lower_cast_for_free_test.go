package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerRishkarsExpertiseCastsForFree(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Rishkar's Expertise",
		Layout:   "normal",
		TypeLine: "Sorcery",
		OracleText: "Draw cards equal to the greatest power among creatures you control.\n" +
			"You may cast a spell with mana value 5 or less from your hand without paying its mana cost.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Rishkar's Expertise did not lower to a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then cast-for-free", mode.Sequence)
	}
	cast, ok := mode.Sequence[1].Primitive.(game.CastForFree)
	if !ok {
		t.Fatalf("second primitive = %T, want CastForFree", mode.Sequence[1].Primitive)
	}
	if !mode.Sequence[1].Optional {
		t.Fatal("cast-for-free step is not optional (the \"you may\" wrapper was lost)")
	}
	if cast.Player.Kind() != game.PlayerReferenceController || cast.Zone != zone.Hand {
		t.Fatalf("cast = %#v, want controller casting from hand", cast)
	}
	if len(cast.Selection.ExcludedTypes) != 1 || cast.Selection.ExcludedTypes[0] != types.Land {
		t.Fatalf("selection = %#v, want nonland spell", cast.Selection)
	}
	if !cast.Selection.ManaValue.Exists {
		t.Fatalf("selection = %#v, want mana-value restriction", cast.Selection)
	}
}

func TestLowerMandatoryCastForFree(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Free Cast Test",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Cast a spell from your hand without paying its mana cost.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single cast-for-free", mode.Sequence)
	}
	cast, ok := mode.Sequence[0].Primitive.(game.CastForFree)
	if !ok {
		t.Fatalf("primitive = %T, want CastForFree", mode.Sequence[0].Primitive)
	}
	if mode.Sequence[0].Optional {
		t.Fatal("mandatory cast-for-free was marked optional")
	}
	if cast.Selection.ManaValue.Exists {
		t.Fatalf("selection = %#v, want no mana-value restriction", cast.Selection)
	}
}

func TestCastForFreeFailsClosedForPaidCast(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Bad Free Cast",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Cast a spell from your hand.",
	})
}
