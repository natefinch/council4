package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerGraveyardReturnSubtypeQualifiedToHand proves a subtype adjective that
// qualifies a card-type noun ("Zombie creature card") round-trips through the
// parser's exact-effect reconstruction so the targeted graveyard return lowers.
// Before this support graveyardCardNoun rejected a noun that carried both a
// subtype and a card type, leaving the effect inexact and failing closed.
func TestLowerGraveyardReturnSubtypeQualifiedToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Subtype To Hand",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target Zombie creature card from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		!slices.Equal(target.Selection.Val.RequiredTypes, []types.Card{types.Creature}) ||
		!slices.Equal(target.Selection.Val.SubtypesAny, []types.Sub{types.Sub("Zombie")}) {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
		t.Fatalf("primitive = %#v, want MoveCard graveyard -> hand", mode.Sequence[0].Primitive)
	}
}

// TestLowerGraveyardReturnSubtypeQualifiedToBattlefieldTapped proves the
// subtype-qualified return to the battlefield with an entry-tapped rider lowers,
// the mode-2 wording on Deadly Plot ("Return target Zombie creature card from
// your graveyard to the battlefield tapped.").
func TestLowerGraveyardReturnSubtypeQualifiedToBattlefieldTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Subtype Tapped",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Return target Zombie creature card from your graveyard to the battlefield tapped.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if !slices.Equal(target.Selection.Val.SubtypesAny, []types.Sub{types.Sub("Zombie")}) {
		t.Fatalf("target = %#v, want Zombie subtype", target)
	}
	put, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield)
	if !ok || !put.EntryTapped {
		t.Fatalf("primitive = %#v, want PutOnBattlefield tapped", mode.Sequence[0].Primitive)
	}
}

// TestLowerGraveyardReturnSubtypeQualifiedMultiTarget proves the pluralized
// subtype-qualified multi-target return lowers one PutOnBattlefield per target,
// the wording on Wondrous Revival ("Return up to three target Hero creature cards
// from your graveyard to the battlefield.").
func TestLowerGraveyardReturnSubtypeQualifiedMultiTarget(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Subtype Multi",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return up to three target Hero creature cards from your graveyard to the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.MinTargets != 0 || target.MaxTargets != 3 ||
		!slices.Equal(target.Selection.Val.SubtypesAny, []types.Sub{types.Sub("Hero")}) {
		t.Fatalf("target = %#v, want up-to-three Hero", target)
	}
	if len(mode.Sequence) != 3 {
		t.Fatalf("sequence = %#v, want three instructions", mode.Sequence)
	}
	for i, instr := range mode.Sequence {
		if _, ok := instr.Primitive.(game.PutOnBattlefield); !ok {
			t.Fatalf("sequence[%d] = %#v, want PutOnBattlefield", i, instr.Primitive)
		}
	}
}

// TestLowerGraveyardReturnSupertypeExclusionStillFailsClosed guards the narrow
// scope of the subtype-qualified support: a supertype exclusion ("nonlegendary")
// is not captured by the card noun, so the effect stays inexact and the card
// fails closed rather than silently dropping the exclusion.
func TestLowerGraveyardReturnSupertypeExclusionStillFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Nonlegendary Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target nonlegendary creature card from your graveyard to the battlefield.",
	})
}
