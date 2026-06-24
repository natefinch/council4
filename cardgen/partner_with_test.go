package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestGenerateExecutablePartnerWithSource confirms a card whose only otherwise
// unsupported ability is "Partner with <name>" generates: the partner-with
// keyword lowers to the inert PartnerWithStaticBody and the card's other
// representable abilities lower as usual.
func TestGenerateExecutablePartnerWithSource(t *testing.T) {
	t.Parallel()
	source, diagnostics, err := GenerateExecutableCardSource(&ScryfallCard{
		Name:       "Brallin, Skyshark Rider",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Shaman",
		ManaCost:   "{3}{R}",
		OracleText: "Partner with Shabraz, the Skyshark (When this creature enters, target player may put Shabraz into their hand from their library, then shuffle.)\nWhenever you discard a card, put a +1/+1 counter on Brallin and it deals 1 damage to each opponent.\n{R}: Target Shark gains trample until end of turn.",
	}, "b")
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

// TestLowerPartnerWithStaticKeyword confirms the "Partner with <name>" keyword
// ability lowers to a single inert static ability carrying the PartnerWith
// keyword, mirroring the companion and banding represented-but-not-simulated
// precedents.
func TestLowerPartnerWithStaticKeyword(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Test Partner With Sage",
		Layout:     "normal",
		TypeLine:   "Legendary Creature — Human Shaman",
		ManaCost:   "{3}{R}",
		OracleText: "Partner with Other Sage (When this creature enters, target player may put Other Sage into their hand from their library, then shuffle.)",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %d, want 1", len(face.StaticAbilities))
	}
	body := face.StaticAbilities[0].Body
	if !game.BodyHasKeyword(&body, game.PartnerWith) {
		t.Fatal("lowered static ability does not carry the PartnerWith keyword")
	}
}
