package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// chosenPutPrimitive lowers a single-face non-target reanimation spell and
// returns its sole ChooseFromZone primitive, asserting the lowering produced one
// mode with one instruction.
func chosenPutPrimitive(t *testing.T, oracle string) game.ChooseFromZone {
	t.Helper()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Chosen Put",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracle,
	})
	mode := face.SpellAbility.Val.Modes[0]
	if len(mode.Targets) != 0 {
		t.Fatalf("targets = %#v, want none", mode.Targets)
	}
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", mode.Sequence)
	}
	choose, ok := mode.Sequence[0].Primitive.(game.ChooseFromZone)
	if !ok {
		t.Fatalf("primitive = %T, want game.ChooseFromZone", mode.Sequence[0].Primitive)
	}
	if choose.SourceZone != zone.Graveyard || choose.Destination.Zone != zone.Battlefield {
		t.Fatalf("choose = %#v, want graveyard -> battlefield", choose)
	}
	return choose
}

// TestLowerChosenGraveyardPutYourGraveyard proves the non-target EffectPut
// reanimation "Put a <filter> card from your graveyard onto the battlefield under
// your control." lowers to a controller-scoped ChooseFromZone (AllOwners is
// false) whose chosen card enters under the controller's control. This is the
// EffectPut counterpart of the already-supported EffectReturn reanimation; before
// this lowering the put form fell through to a misleading counter-placement
// diagnostic.
func TestLowerChosenGraveyardPutYourGraveyard(t *testing.T) {
	t.Parallel()
	choose := chosenPutPrimitive(t,
		"Put a creature card from your graveyard onto the battlefield under your control.")
	if choose.AllOwners {
		t.Fatal("AllOwners = true, want false for \"from your graveyard\"")
	}
	if !slices.Equal(choose.Filter.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("filter = %#v, want creature", choose.Filter)
	}
	if choose.Riders.EntersTapped {
		t.Fatal("EntersTapped = true, want false")
	}
}

// TestLowerChosenGraveyardPutAnyGraveyard proves the "from a graveyard" wording
// widens the candidate pool across every player's graveyard (AllOwners is true),
// the shape Extract from Darkness and Soul of Windgrace use, while still putting
// the chosen card onto the battlefield under the resolving controller's control.
func TestLowerChosenGraveyardPutAnyGraveyard(t *testing.T) {
	t.Parallel()
	choose := chosenPutPrimitive(t,
		"Put a creature card from a graveyard onto the battlefield under your control.")
	if !choose.AllOwners {
		t.Fatal("AllOwners = false, want true for \"from a graveyard\"")
	}
	if !slices.Equal(choose.Filter.RequiredTypes, []types.Card{types.Creature}) {
		t.Fatalf("filter = %#v, want creature", choose.Filter)
	}
}

// TestLowerChosenGraveyardPutTapped proves the "tapped" entry rider on the
// non-target put reanimation lowers to a ChooseFromZone whose chosen card enters
// tapped.
func TestLowerChosenGraveyardPutTapped(t *testing.T) {
	t.Parallel()
	choose := chosenPutPrimitive(t,
		"Put a creature card from a graveyard onto the battlefield tapped under your control.")
	if !choose.AllOwners {
		t.Fatal("AllOwners = false, want true")
	}
	if !choose.Riders.EntersTapped {
		t.Fatal("EntersTapped = false, want true for \"tapped\"")
	}
}

// TestLowerChosenGraveyardPutRejectsOwnersControl proves a put that returns the
// chosen card under its owner's control rather than the resolving controller's
// fails closed: ReturnFromGraveyardChoice always seats the card under the
// chooser, so the owner's-control variant has no faithful lowering.
func TestLowerChosenGraveyardPutRejectsOwnersControl(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Owners Control Put",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Put a creature card from a graveyard onto the battlefield under its owner's control.",
	})
}
