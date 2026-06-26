package cardgen

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// digRevealPrimitive lowers a single-face card whose only ability is a spell or
// triggered body and returns the resolved Dig primitive, failing the test if the
// body did not lower to exactly one Dig instruction.
func digRevealPrimitive(t *testing.T, content game.AbilityContent) game.Dig {
	t.Helper()
	if len(content.Modes) != 1 {
		t.Fatalf("content.Modes = %d, want 1", len(content.Modes))
	}
	mode := content.Modes[0]
	if len(mode.Targets) != 0 || len(mode.Sequence) != 1 {
		t.Fatalf("mode = %+v, want no targets and one instruction", mode)
	}
	dig, ok := mode.Sequence[0].Primitive.(game.Dig)
	if !ok {
		t.Fatalf("primitive = %T, want game.Dig", mode.Sequence[0].Primitive)
	}
	return dig
}

// TestLowerOptionalDigRevealSingleType verifies the typed optional-reveal dig
// ("Look at the top N cards of your library. You may reveal a [type] card from
// among them and put it into your hand. Put the rest on the bottom of your
// library...") lowers to one Dig that looks at N, may take up to one matching
// card, reveals it, and bottoms the remainder.
func TestLowerOptionalDigRevealSingleType(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Reveal",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top four cards of your library. You may reveal an artifact card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
	})
	if !face.SpellAbility.Exists {
		t.Fatal("spell ability not lowered")
	}
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Look != game.Fixed(4) || dig.Take != game.Fixed(1) {
		t.Fatalf("dig = %+v, want Look 4 Take 1", dig)
	}
	if !dig.TakeUpTo {
		t.Fatal("dig.TakeUpTo = false, want true (the reveal is optional)")
	}
	if !dig.Reveal {
		t.Fatal("dig.Reveal = false, want true")
	}
	if dig.Remainder != game.DigRemainderLibraryBottom {
		t.Fatalf("dig.Remainder = %v, want library bottom", dig.Remainder)
	}
	if dig.Player != game.ControllerReference() {
		t.Fatalf("dig.Player = %+v, want controller", dig.Player)
	}
	if !dig.Filter.Exists {
		t.Fatal("dig.Filter absent, want an artifact-card filter")
	}
	if got := dig.Filter.Val.RequiredTypes; !reflect.DeepEqual(got, []types.Card{types.Artifact}) {
		t.Fatalf("dig.Filter.Val.RequiredTypes = %v, want [Artifact]", got)
	}
}

// TestLowerOptionalDigRevealSubtype verifies a creature-subtype reveal filter
// ("reveal a Dragon card") projects onto the Dig filter's SubtypesAny.
func TestLowerOptionalDigRevealSubtype(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Subtype",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top five cards of your library. You may reveal a Dragon card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
	})
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Look != game.Fixed(5) || dig.Take != game.Fixed(1) || !dig.TakeUpTo {
		t.Fatalf("dig = %+v, want Look 5 Take up-to 1", dig)
	}
	if !dig.Filter.Exists || !reflect.DeepEqual(dig.Filter.Val.SubtypesAny, []types.Sub{types.Sub("Dragon")}) {
		t.Fatalf("dig.Filter = %+v, want SubtypesAny [Dragon]", dig.Filter)
	}
}

// TestLowerOptionalDigRevealExcludedTypes verifies an excluded-type reveal
// filter ("reveal a noncreature, nonland card") projects onto ExcludedTypes.
func TestLowerOptionalDigRevealExcludedTypes(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Excluded",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top four cards of your library. You may reveal a noncreature, nonland card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
	})
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if !dig.Filter.Exists {
		t.Fatal("dig.Filter absent, want excluded-type filter")
	}
	if got := dig.Filter.Val.ExcludedTypes; !reflect.DeepEqual(got, []types.Card{types.Creature, types.Land}) {
		t.Fatalf("dig.Filter.Val.ExcludedTypes = %v, want [Creature Land]", got)
	}
}

// TestLowerOptionalDigRevealAnyNumber verifies the "any number" form bounds the
// take by the look count and keeps the optional/up-to semantics.
func TestLowerOptionalDigRevealAnyNumber(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Any Number",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top five cards of your library. You may reveal any number of creature cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.",
	})
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Look != game.Fixed(5) || dig.Take != game.Fixed(5) || !dig.TakeUpTo {
		t.Fatalf("dig = %+v, want Look 5 Take up-to 5", dig)
	}
	if !dig.Reveal || dig.Remainder != game.DigRemainderLibraryBottom {
		t.Fatalf("dig = %+v, want reveal and library-bottom remainder", dig)
	}
}

// TestLowerOptionalDigRevealUpToTwo verifies the bounded "up to two" form sets
// the take to two while remaining an optional upper bound.
func TestLowerOptionalDigRevealUpToTwo(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Up To Two",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top five cards of your library. You may reveal up to two instant and/or sorcery cards from among them and put them into your hand. Put the rest on the bottom of your library in a random order.",
	})
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Take != game.Fixed(2) || !dig.TakeUpTo {
		t.Fatalf("dig = %+v, want Take up-to 2", dig)
	}
}

// TestLowerOptionalDigRevealGraveyardRemainder verifies the graveyard-remainder
// variant routes the rest to the graveyard.
func TestLowerOptionalDigRevealGraveyardRemainder(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Optional Dig Graveyard",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top four cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest into your graveyard.",
	})
	dig := digRevealPrimitive(t, face.SpellAbility.Val)
	if dig.Remainder != game.DigRemainderGraveyard {
		t.Fatalf("dig.Remainder = %v, want graveyard", dig.Remainder)
	}
}

// TestLowerOptionalDigRevealMandatoryFailsClosed verifies a body whose reveal is
// not optional ("reveal a creature card", no "you may") does not lower through
// the optional-reveal dig path.
func TestLowerOptionalDigRevealMandatoryFailsClosed(t *testing.T) {
	t.Parallel()
	faces, _ := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Mandatory Dig Reveal",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		OracleText: "Look at the top four cards of your library. Reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in any order.",
	})
	if len(faces) == 0 {
		return
	}
	if faces[0].SpellAbility.Exists {
		if dig, ok := faces[0].SpellAbility.Val.Modes[0].Sequence[0].Primitive.(game.Dig); ok && dig.TakeUpTo {
			t.Fatal("mandatory reveal lowered through the optional-reveal dig path")
		}
	}
}
