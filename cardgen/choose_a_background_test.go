package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestGenerateExecutableChooseABackgroundSource confirms a card whose only
// otherwise unsupported ability is "Choose a Background" generates: the keyword
// lowers to the inert ChooseABackgroundStaticBody and the card's other
// representable abilities lower as usual.
func TestGenerateExecutableChooseABackgroundSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Jaheira, Friend of the Forest",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Elf Druid",
		ManaCost:   "{2}{G}",
		OracleText: "Tokens you control have \"{T}: Add {G}.\"\nChoose a Background (You can have a Background as a second commander.)",
	}, "j")
	if err != nil {
		t.Fatal(err)
	}
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if source == "" {
		t.Fatal("empty generated source")
	}
}

// TestLowerChooseABackgroundStaticKeyword confirms the "Choose a Background"
// keyword ability lowers to a single inert static ability carrying the
// ChooseABackground keyword, mirroring the companion and partner-with
// represented-but-not-simulated precedents.
func TestLowerChooseABackgroundStaticKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Background Sage",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Druid",
		ManaCost:   "{2}{G}",
		OracleText: "Choose a Background (You can have a Background as a second commander.)",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if !game.BodyHasKeyword(&body, game.ChooseABackground) {
		t.Fatal("lowered static ability does not carry the ChooseABackground keyword")
	}
}
