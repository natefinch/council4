package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerGrowthSpiralPutsLandFromHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Growth Spiral",
		Layout:     "normal",
		TypeLine:   "Instant",
		OracleText: "Draw a card. You may put a land card from your hand onto the battlefield.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("Growth Spiral did not lower to a spell ability")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then put-from-hand", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.Draw); !ok {
		t.Fatalf("first primitive = %T, want Draw", mode.Sequence[0].Primitive)
	}
	put, ok := mode.Sequence[1].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("second primitive = %T, want game.ChooseFromZone", mode.Sequence[1].Primitive)
	}
	if !mode.Sequence[1].Optional {
		t.Fatal("put-from-hand step is not optional (the \"you may\" wrapper was lost)")
	}
	if put.Player.Kind() != game.PlayerReferenceController || put.Quantity.Value() != 1 {
		t.Fatalf("put = %#v, want controller put one", put)
	}
	if len(put.Filter.RequiredTypes) != 1 || put.Filter.RequiredTypes[0] != types.Land {
		t.Fatalf("selection = %#v, want land card", put.Filter)
	}
}

func TestLowerMandatoryPutLandFromHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ramp",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a land card from your hand onto the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single put-from-hand", mode.Sequence)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone); !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if mode.Sequence[0].Optional {
		t.Fatal("mandatory put-from-hand was marked optional")
	}
}

func TestLowerPutFromHandTappedAndAttacking(t *testing.T) {
	t.Parallel()
	// Preeminent Captain shape: a self-attack trigger whose optional resolving
	// effect puts a creature card from hand onto the battlefield tapped and
	// attacking (CR 508.4).
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Captain",
		Layout:     "normal",
		TypeLine:   "Creature — Kithkin Soldier",
		OracleText: "Whenever this creature attacks, you may put a Soldier creature card from your hand onto the battlefield tapped and attacking.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	if !face.TriggeredAbilities[0].Optional {
		t.Fatal("triggered ability is not optional (the \"you may\" wrapper was lost)")
	}
	mode := face.TriggeredAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single put-from-hand", mode.Sequence)
	}
	put, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if !put.Riders.EntersTapped {
		t.Fatal("put-from-hand did not carry the \"tapped\" entry rider")
	}
	if !put.Riders.EntersAttacking {
		t.Fatal("put-from-hand did not carry the \"attacking\" entry rider")
	}
	if len(put.Filter.RequiredTypes) != 1 || put.Filter.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %#v, want creature card", put.Filter)
	}
}

func TestPutFromHandFailsClosedForLibrarySource(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Bad Ramp",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a land card from your library onto the battlefield.",
	})
}

// TestLowerPutAnyNumberCreaturesFromHand proves the "put any number of <filter>
// cards from your hand onto the battlefield" mass form (Ghalta, Stampede Tyrant)
// lowers to a ChooseFromZone whose Count is ChooseAnyNumber and whose filter,
// source, and destination match "creature cards from hand onto the battlefield".
func TestLowerPutAnyNumberCreaturesFromHand(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Ghalta",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put any number of creature cards from your hand onto the battlefield.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single put-from-hand", mode.Sequence)
	}
	if mode.Sequence[0].Optional {
		t.Fatal("any-number put-from-hand was marked optional (the empty choice is intrinsic)")
	}
	put, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if put.Count != game.ChooseAnyNumber {
		t.Fatalf("count = %v, want game.ChooseAnyNumber", put.Count)
	}
	if put.SourceZone != zone.Hand || put.Destination.Zone != zone.Battlefield {
		t.Fatalf("zones = %v -> %v, want hand -> battlefield", put.SourceZone, put.Destination.Zone)
	}
	if len(put.Filter.RequiredTypes) != 1 || put.Filter.RequiredTypes[0] != types.Creature {
		t.Fatalf("selection = %#v, want creature card", put.Filter)
	}
}

// TestPutFromHandFailsClosedForUpToRange proves the bounded "up to two" range
// form is not silently collapsed onto the any-number path: it fails closed
// because the put-from-hand lowerer models only exactly-one and any-number.
func TestPutFromHandFailsClosedForUpToRange(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Bad Ghalta",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put up to two creature cards from your hand onto the battlefield.",
	})
}

// TestLowerLastMarchOfTheEnts proves the full card composes: the "can't be
// countered" static survives, and the ordered spell body lowers to a draw whose
// amount is the greatest toughness among the controller's creatures followed by
// an any-number put-from-hand of creature cards.
func TestLowerLastMarchOfTheEnts(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Last March of the Ents",
		Layout:   "normal",
		ManaCost: "{6}{G}{G}",
		TypeLine: "Sorcery",
		OracleText: "This spell can't be countered.\n" +
			"Draw cards equal to the greatest toughness among creatures you control, then put any number of creature cards from your hand onto the battlefield.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1 (can't be countered)", len(face.StaticAbilities))
	}
	if !face.SpellAbility.Exists {
		t.Fatal("no spell ability lowered")
	}
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then put-any-number", mode.Sequence)
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("first primitive = %T, want game.Draw", mode.Sequence[0].Primitive)
	}
	dyn := draw.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountGreatestToughnessInGroup {
		t.Fatalf("draw amount = %#v, want greatest toughness in group", draw.Amount)
	}
	put, ok := mode.Sequence[1].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("second primitive = %T, want game.ChooseFromZone", mode.Sequence[1].Primitive)
	}
	if put.Count != game.ChooseAnyNumber {
		t.Fatalf("count = %v, want game.ChooseAnyNumber", put.Count)
	}
	if put.SourceZone != zone.Hand || put.Destination.Zone != zone.Battlefield {
		t.Fatalf("zones = %v -> %v, want hand -> battlefield", put.SourceZone, put.Destination.Zone)
	}
	if len(put.Filter.RequiredTypes) != 1 || put.Filter.RequiredTypes[0] != types.Creature {
		t.Fatalf("put selection = %#v, want creature card", put.Filter)
	}
}

func TestLowerMandatoryPutLandFromHandTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Tapped Ramp",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a land card from your hand onto the battlefield tapped.",
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single put-from-hand", mode.Sequence)
	}
	put, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if !put.Riders.EntersTapped {
		t.Fatal("put-from-hand did not carry the \"tapped\" entry rider")
	}
	if mode.Sequence[0].Optional {
		t.Fatal("mandatory put-from-hand was marked optional")
	}
	if len(put.Filter.RequiredTypes) != 1 || put.Filter.RequiredTypes[0] != types.Land {
		t.Fatalf("selection = %#v, want land card", put.Filter)
	}
}

func TestLowerOptionalPutLandFromHandTapped(t *testing.T) {
	t.Parallel()
	// Horizon of Progress: an activated ability whose sole resolving effect is an
	// optional put-from-hand onto the battlefield tapped.
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Horizon",
		Layout:     "normal",
		TypeLine:   "Land",
		OracleText: "{3}, {T}: You may put a land card from your hand onto the battlefield tapped.",
	})
	if len(face.ActivatedAbilities) != 1 {
		t.Fatalf("activated abilities = %d, want 1", len(face.ActivatedAbilities))
	}
	mode := face.ActivatedAbilities[0].Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want single put-from-hand", mode.Sequence)
	}
	put, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if !put.Riders.EntersTapped {
		t.Fatal("put-from-hand did not carry the \"tapped\" entry rider")
	}
	if !mode.Sequence[0].Optional {
		t.Fatal("put-from-hand step is not optional (the \"you may\" wrapper was lost)")
	}
}
