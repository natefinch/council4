package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerMemoryPlunderTargetedGraveyardFreeCast proves the single-sentence
// targeted free cast from an opponent's graveyard (Memory Plunder) lowers to a
// CastForFree that casts the lone targeted graveyard card, with no exile rider.
func TestLowerMemoryPlunderTargetedGraveyardFreeCast(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Memory Plunder",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "You may cast target instant or sorcery card from an opponent's graveyard without paying its mana cost.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Memory Plunder did not lower to a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].TargetZone != zone.Graveyard {
		t.Fatalf("targets = %#v, want one graveyard-card target", mode.Targets)
	}
	if got := mode.Targets[0].Selection.Val.Controller; got != game.ControllerOpponent {
		t.Fatalf("target controller = %v, want opponent's graveyard", got)
	}
	if len(mode.Sequence) != 1 || !mode.Sequence[0].Optional {
		t.Fatalf("sequence = %#v, want a single optional instruction", mode.Sequence)
	}
	cast, ok := mode.Sequence[0].Primitive.(game.CastForFree)
	if !ok {
		t.Fatalf("primitive = %T, want CastForFree", mode.Sequence[0].Primitive)
	}
	if cast.Card.Kind != game.CardReferenceTarget || cast.Zone != zone.Graveyard {
		t.Fatalf("cast = %#v, want a graveyard target cast", cast)
	}
	if cast.ExileOnResolution {
		t.Fatal("Memory Plunder has no exile-instead rider but ExileOnResolution is set")
	}
}

// TestLowerTorrentialGearhulkRider proves the ETB-triggered free cast with the
// "exile it instead" rider (Torrential Gearhulk) lowers to a CastForFree that
// casts the targeted graveyard card with ExileOnResolution set.
func TestLowerTorrentialGearhulkRider(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Torrential Gearhulk",
		Layout:   "normal",
		TypeLine: "Artifact Creature — Construct",
		OracleText: "Flash\n" +
			"When this creature enters, you may cast target instant card from your graveyard without paying its mana cost. If that spell would be put into your graveyard, exile it instead.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Targets) != 1 || mode.Targets[0].TargetZone != zone.Graveyard {
		t.Fatalf("targets = %#v, want one graveyard-card target", mode.Targets)
	}
	if got := mode.Targets[0].Selection.Val.RequiredTypesAny; len(got) != 1 || got[0] != types.Instant {
		t.Fatalf("target types = %#v, want instant", got)
	}
	if len(mode.Sequence) != 1 || !mode.Sequence[0].Optional {
		t.Fatalf("sequence = %#v, want a single optional instruction", mode.Sequence)
	}
	cast, ok := mode.Sequence[0].Primitive.(game.CastForFree)
	if !ok {
		t.Fatalf("primitive = %T, want CastForFree", mode.Sequence[0].Primitive)
	}
	if cast.Card.Kind != game.CardReferenceTarget || cast.Zone != zone.Graveyard {
		t.Fatalf("cast = %#v, want a graveyard target cast", cast)
	}
	if !cast.ExileOnResolution {
		t.Fatal("Torrential Gearhulk's exile-instead rider did not set ExileOnResolution")
	}
}

// TestLowerDreadhordeArcanistDynamicBoundUnsupported proves the family stays
// fail-closed for a dynamic ("mana value less than or equal to its power") bound
// this backend cannot express on a graveyard-card selection.
func TestLowerDreadhordeArcanistDynamicBoundUnsupported(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Dreadhorde Arcanist",
		Layout:   "normal",
		TypeLine: "Creature — Zombie Wizard",
		OracleText: "Trample\n" +
			"Whenever Dreadhorde Arcanist attacks, you may cast target instant or sorcery card with mana value less than or equal to Dreadhorde Arcanist's power from your graveyard without paying its mana cost. If that spell would be put into your graveyard this turn, exile it instead.",
	})
}
