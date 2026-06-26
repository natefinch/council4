package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestLowerOptionalRevealTakeSingleType verifies the typed optional-take reveal
// dig ("Reveal the top N cards of your library. You may put a [type] card from
// among them into your hand. Put the rest into your graveyard.") lowers to one
// Dig that looks at N, may take up to one matching card, reveals it, and
// graveyards the remainder.
func TestLowerOptionalRevealTakeSingleType(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reveal Take Single",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Reveal the top four cards of your library. You may put a land card from among them into your hand. Put the rest into your graveyard.",
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
	if !dig.Reveal {
		t.Fatal("dig.Reveal = false, want true")
	}
	if dig.Remainder != game.DigRemainderGraveyard {
		t.Fatalf("dig.Remainder = %v, want graveyard", dig.Remainder)
	}
	if dig.Player != game.ControllerReference() {
		t.Fatalf("dig.Player = %+v, want controller", dig.Player)
	}
	if !dig.Filter.Exists {
		t.Fatal("dig.Filter absent, want a land-card filter")
	}
	if got := dig.Filter.Val.RequiredTypes; !reflect.DeepEqual(got, []types.Card{types.Land}) {
		t.Fatalf("dig.Filter.Val.RequiredTypes = %v, want [Land]", got)
	}
}

// TestLowerOptionalRevealTakeOrFilter verifies a two-type "or" reveal filter
// ("a creature or enchantment card") projects onto the Dig filter's AnyTypes
// while remaining a take-one upper bound.
func TestLowerOptionalRevealTakeOrFilter(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Reveal Take Or",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Reveal the top five cards of your library. You may put a creature or enchantment card from among them into your hand. Put the rest into your graveyard.",
	})
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Look != game.Fixed(5) || dig.Take != game.Fixed(1) || !dig.TakeUpTo {
		t.Fatalf("dig = %+v, want Look 5 Take up-to 1", dig)
	}
	if dig.Remainder != game.DigRemainderGraveyard {
		t.Fatalf("dig.Remainder = %v, want graveyard", dig.Remainder)
	}
	if !dig.Filter.Exists {
		t.Fatal("dig.Filter absent, want a creature-or-enchantment filter")
	}
	if got := dig.Filter.Val.RequiredTypesAny; !reflect.DeepEqual(got, []types.Card{types.Creature, types.Enchantment}) {
		t.Fatalf("dig.Filter.Val.RequiredTypesAny = %v, want [Creature Enchantment]", got)
	}
}

// TestLowerOptionalRevealTakeAndOrFailsClosed verifies the inclusive
// "and/or" form ("a creature card and/or an enchantment card"), whose per-type
// take count a single flat Dig take cannot model, fails closed rather than
// lowering a silently-wrong single-take Dig.
func TestLowerOptionalRevealTakeAndOrFailsClosed(t *testing.T) {
	t.Parallel()
	faces, _ := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Reveal Take And Or",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Reveal the top five cards of your library. You may put a creature card and/or an enchantment card from among them into your hand. Put the rest into your graveyard.",
	})
	if len(faces) == 0 {
		return
	}
	if faces[0].SpellAbility.Exists {
		if len(faces[0].SpellAbility.Val.Modes) == 1 &&
			len(faces[0].SpellAbility.Val.Modes[0].Sequence) == 1 {
			if _, ok := faces[0].SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Dig); ok {
				t.Fatal("and/or reveal-take lowered to a single-take Dig")
			}
		}
	}
}

// TestLowerOptionalRevealTakeLibraryBottomFailsClosed verifies that a
// library-bottom remainder, which is not behaviorally equivalent to revealing
// the top N cards (the bottomed cards would have been publicly revealed first),
// fails closed instead of lowering through the reveal-take dig path.
func TestLowerOptionalRevealTakeLibraryBottomFailsClosed(t *testing.T) {
	t.Parallel()
	faces, _ := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Reveal Take Bottom",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Reveal the top four cards of your library. You may put a creature card from among them into your hand. Put the rest on the bottom of your library in any order.",
	})
	if len(faces) == 0 {
		return
	}
	if faces[0].SpellAbility.Exists {
		if len(faces[0].SpellAbility.Val.Modes) == 1 &&
			len(faces[0].SpellAbility.Val.Modes[0].Sequence) == 1 {
			if dig, ok := faces[0].SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Dig); ok && dig.TakeUpTo {
				t.Fatal("library-bottom reveal-take lowered through the reveal-take dig path")
			}
		}
	}
}

// TestLowerOptionalRevealTakeMandatoryFailsClosed verifies a body whose put is
// not optional ("put a creature card", no "you may") does not lower through the
// optional reveal-take dig path.
func TestLowerOptionalRevealTakeMandatoryFailsClosed(t *testing.T) {
	t.Parallel()
	faces, _ := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Reveal Take Mandatory",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Reveal the top four cards of your library. Put a creature card from among them into your hand. Put the rest into your graveyard.",
	})
	if len(faces) == 0 {
		return
	}
	if faces[0].SpellAbility.Exists {
		if len(faces[0].SpellAbility.Val.Modes) == 1 &&
			len(faces[0].SpellAbility.Val.Modes[0].Sequence) == 1 {
			if dig, ok := faces[0].SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Dig); ok && dig.TakeUpTo {
				t.Fatal("mandatory reveal-take lowered through the optional reveal-take dig path")
			}
		}
	}
}
