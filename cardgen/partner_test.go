package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestGenerateExecutablePartnerSource confirms a card whose only otherwise
// unsupported ability is the "Partner" ability word (including the
// "Partner—<variant>" forms such as "Partner—Character select") generates: the
// keyword lowers to the inert PartnerStaticBody and the card's other
// representable abilities lower as usual.
func TestGenerateExecutablePartnerSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:     "Donatello, the Brains",
		Layout:   "normal",
		TypeLine: "Legendary Creature — Turtle Warrior",
		ManaCost: "{2}{U}",
		OracleText: "If one or more tokens would be created under your control, those tokens plus a Mutagen token are created instead.\n" +
			"Partner—Character select (You can have two commanders if both have this ability.)",
		Power:     new("2"),
		Toughness: new("4"),
	}, "d")
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

// TestLowerPartnerStaticKeyword confirms the plain "Partner" keyword ability
// lowers to a single inert static ability carrying the Partner keyword,
// mirroring the companion, partner-with, and choose-a-background
// represented-but-not-simulated precedents.
func TestLowerPartnerStaticKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Partner Hero",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Soldier",
		ManaCost:   "{1}{W}",
		OracleText: "Partner (You can have two commanders if both have partner.)",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if !game.BodyHasKeyword(&body, game.Partner) {
		t.Fatal("lowered static ability does not carry the Partner keyword")
	}
}
