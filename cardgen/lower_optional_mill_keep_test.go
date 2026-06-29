package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerOptionalMillKeepSingleType verifies the typed optional-take mill
// ("Mill N cards. You may put a [type] card from among the cards milled this way
// into your hand.") lowers to one Dig that mills N, may take up to one matching
// card into hand, and leaves the rest in the graveyard.
func TestLowerOptionalMillKeepSingleType(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Mill Keep Single",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Mill four cards. You may put a creature card from among the cards milled this way into your hand.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Look != game.Fixed(4) || dig.Take != game.Fixed(1) {
		t.Fatalf("dig = %+v, want Look 4 Take 1", dig)
	}
	if !dig.TakeUpTo {
		t.Fatal("dig.TakeUpTo = false, want true (the put is optional)")
	}
	if dig.Remainder != game.DigRemainderGraveyard {
		t.Fatalf("dig.Remainder = %v, want graveyard", dig.Remainder)
	}
	if dig.Player != game.ControllerReference() {
		t.Fatalf("dig.Player = %+v, want controller", dig.Player)
	}
	if !dig.Filter.Exists {
		t.Fatal("dig.Filter absent, want a creature-card filter")
	}
	if got := dig.Filter.Val.RequiredTypes; !reflect.DeepEqual(got, []types.Card{types.Creature}) {
		t.Fatalf("dig.Filter.Val.RequiredTypes = %v, want [Creature]", got)
	}
}

// TestLowerOptionalMillKeepGraveyardSourceFailsClosed verifies a "put from your
// graveyard" form whose source is the whole graveyard rather than the cards just
// milled is not lowered as the optional mill-keep dig.
func TestLowerOptionalMillKeepGraveyardSourceFailsClosed(t *testing.T) {
	t.Parallel()
	lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
		Name:       "Test Mill Keep Graveyard",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Mill five cards. Then you may put an artifact card from your graveyard on top of your library.",
	})
}
