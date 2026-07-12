package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerMobilizeFixedEndToEnd proves a printed "Mobilize 2" creature lowers to
// exactly the reusable fixed Mobilize triggered body, text-blind through the
// stripped reminder text.
func TestLowerMobilizeFixedEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Mobilize Two Tester",
		Layout:   "normal",
		TypeLine: "Creature — Human Soldier",
		OracleText: "Mobilize 2 (Whenever this creature attacks, create two tapped and attacking 1/1 red Warrior creature tokens. " +
			"Sacrifice them at the beginning of the next end step.)",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	want := game.MobilizeTriggeredBody(game.MobilizeAmount{Fixed: 2})
	if !reflect.DeepEqual(face.TriggeredAbilities[0], want) {
		t.Fatalf("lowered ability = %#v, want fixed Mobilize 2 body", face.TriggeredAbilities[0])
	}
}

// TestLowerMobilizeDynamicGraveyardEndToEnd proves the "Mobilize X, where X is
// the number of creature cards in your graveyard" form lowers to the reusable
// dynamic Mobilize body (Avenger of the Fallen).
func TestLowerMobilizeDynamicGraveyardEndToEnd(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Mobilize Graveyard Tester",
		Layout:   "normal",
		TypeLine: "Creature — Zombie Warrior",
		OracleText: "Mobilize X, where X is the number of creature cards in your graveyard. " +
			"(Whenever this creature attacks, create X tapped and attacking 1/1 red Warrior creature tokens. " +
			"Sacrifice them at the beginning of the next end step.)",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	want := game.MobilizeTriggeredBody(game.MobilizeAmount{Dynamic: game.MobilizeDynamicCreatureCardsInGraveyard})
	if !reflect.DeepEqual(face.TriggeredAbilities[0], want) {
		t.Fatalf("lowered ability = %#v, want dynamic graveyard Mobilize body", face.TriggeredAbilities[0])
	}
}

// TestLowerMobilizeUnsupportedDynamicFailsClosed proves an unrepresentable
// dynamic Mobilize form ("where X is its power") fails closed rather than
// silently lowering to a wrong amount.
func TestLowerMobilizeUnsupportedDynamicFailsClosed(t *testing.T) {
	t.Parallel()
	face := lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:     "Mobilize Power Tester",
		Layout:   "normal",
		TypeLine: "Creature — Human Warrior",
		OracleText: "Mobilize X, where X is its power. " +
			"(Whenever this creature attacks, create X tapped and attacking 1/1 red Warrior creature tokens. " +
			"Sacrifice them at the beginning of the next end step.)",
	})
	if len(face.TriggeredAbilities) != 0 {
		t.Fatalf("unsupported Mobilize produced %d triggered abilities, want 0", len(face.TriggeredAbilities))
	}
}
