package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestLowerOptionalDigBattlefieldTapped verifies the typed optional
// dig-to-battlefield ("Look at the top N cards of your library. You may put a
// [type] card from among them onto the battlefield tapped. Put the rest on the
// bottom...") lowers to one Dig that looks at N, may put up to one matching
// card onto the battlefield tapped, and bottoms the remainder.
func TestLowerOptionalDigBattlefieldTapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Battlefield",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Look != game.Fixed(5) || dig.Take != game.Fixed(1) || !dig.TakeUpTo {
		t.Fatalf("dig = %+v, want Look 5 Take up-to 1", dig)
	}
	if dig.Destination != zone.Battlefield {
		t.Fatalf("dig.Destination = %v, want battlefield", dig.Destination)
	}
	if !dig.EntersTapped {
		t.Fatal("dig.EntersTapped = false, want true (entry says tapped)")
	}
	if dig.Reveal {
		t.Fatal("dig.Reveal = true, want false (battlefield put, not a reveal-to-hand)")
	}
	if dig.Remainder != game.DigRemainderLibraryBottom {
		t.Fatalf("dig.Remainder = %v, want library bottom", dig.Remainder)
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

// TestLowerOptionalDigBattlefieldUntapped verifies that without a "tapped"
// entry rider the put lowers with EntersTapped false.
func TestLowerOptionalDigBattlefieldUntapped(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Battlefield Untapped",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top four cards of your library. You may put a creature card from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Look != game.Fixed(4) || dig.Take != game.Fixed(1) || !dig.TakeUpTo {
		t.Fatalf("dig = %+v, want Look 4 Take up-to 1", dig)
	}
	if dig.Destination != zone.Battlefield {
		t.Fatalf("dig.Destination = %v, want battlefield", dig.Destination)
	}
	if dig.EntersTapped {
		t.Fatal("dig.EntersTapped = true, want false (no tapped entry)")
	}
}

// TestLowerOptionalDigBattlefieldUpToTwo verifies the bounded "up to two" form
// sets the take to two while remaining an optional upper bound and projects a
// mana-value-bounded permanent filter onto the battlefield.
func TestLowerOptionalDigBattlefieldUpToTwo(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Battlefield Up To Two",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top five cards of your library. You may put up to two permanent cards with mana value 5 or less from among them onto the battlefield. Put the rest on the bottom of your library in a random order.",
	})
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Take != game.Fixed(2) || !dig.TakeUpTo {
		t.Fatalf("dig = %+v, want Take up-to 2", dig)
	}
	if dig.Destination != zone.Battlefield {
		t.Fatalf("dig.Destination = %v, want battlefield", dig.Destination)
	}
}

// TestLowerOptionalDigBattlefieldAttackingFailsClosed verifies an entry rider
// the flat Dig cannot carry ("onto the battlefield tapped and attacking") does
// not lower through the optional dig-to-battlefield path, so the attacking
// rider is never silently dropped.
func TestLowerOptionalDigBattlefieldAttackingFailsClosed(t *testing.T) {
	t.Parallel()
	assertNotBattlefieldDig(t, "Test Dig Battlefield Attacking",
		"Look at the top five cards of your library. You may put a creature card from among them onto the battlefield tapped and attacking. Put the rest on the bottom of your library in a random order.")
}

// TestLowerOptionalDigBattlefieldConditionalFailsClosed verifies a conditional
// rider on the put ("if it has the same name as a permanent you control") does
// not lower through the optional dig-to-battlefield path, so the condition is
// never silently dropped.
func TestLowerOptionalDigBattlefieldConditionalFailsClosed(t *testing.T) {
	t.Parallel()
	assertNotBattlefieldDig(t, "Test Dig Battlefield Conditional",
		"Look at the top five cards of your library. You may put a creature card from among them onto the battlefield if it has the same name as a creature you control. Put the rest on the bottom of your library in a random order.")
}

// assertNotBattlefieldDig fails the test if the named single-faced spell lowers
// to a Dig whose destination is the battlefield, i.e. if it lowered through the
// optional dig-to-battlefield path.
func assertNotBattlefieldDig(t *testing.T, name, oracle string) {
	t.Helper()
	faces, _ := lowerExecutableFaces(&ScryfallCard{
		Name:       name,
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: oracle,
	})
	if len(faces) == 0 || !faces[0].SpellAbility.Exists {
		return
	}
	modes := faces[0].SpellAbility.Val.Modes
	if len(modes) != 1 || len(modes[0].Sequence) != 1 {
		return
	}
	if dig, ok := modes[0].Sequence[0].Primitive.(game.Dig); ok && dig.Destination == zone.Battlefield {
		t.Fatalf("%s lowered through the optional dig-to-battlefield path", name)
	}
}
