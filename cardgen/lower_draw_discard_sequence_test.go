package cardgen

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
)

func TestLowerDrawThenDiscardSequence(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Faithless Looting",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Sorcery",
		OracleText: "Draw two cards, then discard two cards.\nFlashback {2}{R}",
		Colors:     []string{"R"},
	})
	sequence := face.SpellAbility.Val.Modes[0].Sequence
	if len(sequence) != 2 {
		t.Fatalf("sequence = %#v, want draw then discard", sequence)
	}
	draw, ok := sequence[0].Primitive.(game.Draw)
	if !ok || draw.Player != game.ControllerReference() || draw.Amount != game.Fixed(2) {
		t.Fatalf("instruction 0 = %#v, want controller draw two", sequence[0])
	}
	discard, ok := sequence[1].Primitive.(game.Discard)
	if !ok ||
		discard.Player != game.ControllerReference() ||
		discard.Amount != game.Fixed(2) {
		t.Fatalf("instruction 1 = %#v, want controller discard two", sequence[1])
	}
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want flashback", face.StaticAbilities)
	}
	keyword, found := game.BodyKeywordAbility(&face.StaticAbilities[0].Body, game.Flashback)
	if !found {
		t.Fatalf("static abilities = %#v, want flashback", face.StaticAbilities)
	}
	flashback, ok := keyword.(game.FlashbackKeyword)
	if !ok || !slices.Equal(flashback.Cost, cost.Mana{cost.O(2), cost.R}) {
		t.Fatalf("flashback keyword = %#v, want {2}{R}", keyword)
	}
}

func TestGenerateFaithlessLootingExecutableSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Faithless Looting",
		Layout:     "normal",
		ManaCost:   "{R}",
		TypeLine:   "Sorcery",
		OracleText: "Draw two cards, then discard two cards.\nFlashback {2}{R}",
		Colors:     []string{"R"},
	}, "f")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	for _, want := range []string{
		"game.Draw",
		"game.Discard",
		"game.FlashbackKeyword",
		"cost.O(2)",
		"cost.R",
	} {
		if !strings.Contains(source, want) {
			t.Fatalf("generated source missing %q:\n%s", want, source)
		}
	}
}

func TestLowerFlashbackFailsClosedOutsideFixedManaCost(t *testing.T) {
	t.Parallel()
	for _, flashback := range []string{
		"Flashback {X}{R}",
		"Flashback {2}{R}, discard a card.",
	} {
		t.Run(flashback, func(t *testing.T) {
			t.Parallel()
			_, diagnostics := lowerExecutableFaces(&ScryfallCard{
				Name:       "Unsupported Flashback",
				Layout:     "normal",
				ManaCost:   "{R}",
				TypeLine:   "Sorcery",
				OracleText: "Draw two cards, then discard two cards.\n" + flashback,
				Colors:     []string{"R"},
			})
			if len(diagnostics) == 0 {
				t.Fatalf("expected unsupported diagnostic for %q", flashback)
			}
		})
	}
}
