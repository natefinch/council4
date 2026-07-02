package cardgen

import (
	"slices"
	"strings"
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

// TestLowerGraveyardReturnSupertypeExclusion covers a supertype exclusion
// ("nonlegendary creature card") on a graveyard-return target: the negated
// supertype now renders in the card noun and lowers to the runtime
// Selection.ExcludedSupertype, so the card generates rather than failing closed.
func TestLowerGraveyardReturnSupertypeExclusion(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Test Nonlegendary Return",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{B}",
		OracleText: "Return target nonlegendary creature card from your graveyard to the battlefield.",
	}, "t")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if !strings.Contains(source, "ExcludedSupertype: types.Legendary") {
		t.Fatalf("source missing ExcludedSupertype: types.Legendary:\n%s", source)
	}
}

// TestLowerGraveyardReturnSubtypeDisjunctionCardToHand proves a subtype
// disjunction noun with no card type ("Aura or Equipment card") round-trips
// through the parser's exact-effect reconstruction so the targeted graveyard
// return lowers, capturing both subtypes as an OR set (Ironclad Slayer).
func TestLowerGraveyardReturnSubtypeDisjunctionCardToHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Subtype Disjunction To Hand",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target Aura or Equipment card from your graveyard to your hand.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if target.Allow != game.TargetAllowCard || target.TargetZone != zone.Graveyard ||
		len(target.Selection.Val.RequiredTypes) != 0 ||
		!slices.Equal(target.Selection.Val.SubtypesAny, []types.Sub{types.Sub("Aura"), types.Sub("Equipment")}) {
		t.Fatalf("target = %#v", target)
	}
	move, ok := mode.Sequence[0].Primitive.(game.MoveCard)
	if !ok || move.FromZone != zone.Graveyard || move.Destination != zone.Hand {
		t.Fatalf("primitive = %#v, want MoveCard graveyard -> hand", mode.Sequence[0].Primitive)
	}
}

// TestLowerGraveyardReturnSubtypeDisjunctionSerialList proves a longer serial
// subtype disjunction ("Bat, Lizard, Rat, or Squirrel card") round-trips, the
// wording on Mudflat Village.
func TestLowerGraveyardReturnSubtypeDisjunctionSerialList(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Subtype Serial List",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target Bat, Lizard, Rat, or Squirrel card from your graveyard to your hand.",
	})
	target := face.SpellAbility.Val.Modes[0].Targets[0]
	if !slices.Equal(target.Selection.Val.SubtypesAny,
		[]types.Sub{types.Sub("Bat"), types.Sub("Lizard"), types.Sub("Rat"), types.Sub("Squirrel")}) {
		t.Fatalf("target = %#v, want Bat/Lizard/Rat/Squirrel OR set", target)
	}
}

// TestLowerGraveyardReturnSubtypeDisjunctionQualifiedToBattlefield proves a
// subtype disjunction that qualifies a card type ("Angel or Human creature
// card") lowers to the battlefield with RequiredTypes carrying the type (AND)
// and SubtypesAny carrying the subtypes (OR), the wording on Bruna, the Fading
// Light.
func TestLowerGraveyardReturnSubtypeDisjunctionQualifiedToBattlefield(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Subtype Disjunction Qualified",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Return target Angel or Human creature card from your graveyard to the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	target := mode.Targets[0]
	if !slices.Equal(target.Selection.Val.RequiredTypes, []types.Card{types.Creature}) ||
		!slices.Equal(target.Selection.Val.SubtypesAny, []types.Sub{types.Sub("Angel"), types.Sub("Human")}) {
		t.Fatalf("target = %#v, want creature type with Angel/Human OR set", target)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.PutOnBattlefield); !ok {
		t.Fatalf("primitive = %#v, want PutOnBattlefield", mode.Sequence[0].Primitive)
	}
}
