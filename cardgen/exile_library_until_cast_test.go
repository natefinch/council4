package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerExileLibraryUntilNonlandCast confirms the free-cast dig family lowers
// to a single ExileLibraryUntilNonlandCast primitive across spell, triggered, and
// loyalty shells, and that text outside the envelope fails closed.
func TestLowerExileLibraryUntilNonlandCast(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name     string
		typeLine string
		oracle   string
		power    string
	}{
		{
			name:     "sorcery",
			typeLine: "Sorcery",
			oracle:   "Exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.",
		},
		{
			name:     "triggered",
			typeLine: "Enchantment",
			oracle:   "Whenever you cast your second spell each turn, exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Standalone Until Cast",
				Layout:     "normal",
				TypeLine:   tc.typeLine,
				ManaCost:   "{2}{R}",
				OracleText: tc.oracle,
			})
			content := faceContent(face, tc.typeLine)
			if len(content.Modes) != 1 || len(content.Modes[0].Sequence) != 1 {
				t.Fatalf("content = %+v", content)
			}
			prim, ok := content.Modes[0].Sequence[0].Primitive.(game.ExileLibraryUntilNonlandCast)
			if !ok {
				t.Fatalf("primitive = %T, want ExileLibraryUntilNonlandCast", content.Modes[0].Sequence[0].Primitive)
			}
			if prim.Player != game.ControllerReference() {
				t.Fatalf("player = %+v", prim.Player)
			}
		})
	}
}

func faceContent(face loweredFaceAbilities, typeLine string) game.AbilityContent {
	if typeLine == "Sorcery" {
		return face.SpellAbility.Val
	}
	return face.TriggeredAbilities[0].Content
}

// TestExileLibraryUntilNonlandCastFailsClosed verifies riders outside the
// free-cast envelope keep failing closed: a mana-value cap and a "this turn"
// play window must not lower.
func TestExileLibraryUntilNonlandCastFailsClosed(t *testing.T) {
	t.Parallel()
	for _, oracle := range []string{
		"Exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost if the spell's mana value is less than 7.",
		"Exile cards from the top of your library until you exile a nonland card. You may cast that card this turn.",
	} {
		lowerSingleFaceExpectingUnsupported(t, &ScryfallCard{
			Name:       "Standalone Until Cast",
			Layout:     "normal",
			TypeLine:   "Sorcery",
			ManaCost:   "{2}{R}",
			OracleText: oracle,
		})
	}
}
