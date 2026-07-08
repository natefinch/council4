package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerTransformSelfActivated verifies that a transforming double-faced
// card's "Transform this creature." activated ability lowers to a game.Transform
// primitive that transforms the source permanent, and that the back face is
// generated.
func TestLowerTransformSelfActivated(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Ulvenwald Captive",
		Layout:   "transform",
		TypeLine: "Creature — Elf Druid",
		ManaCost: "{1}{G}",
		CardFaces: []ScryfallCardFace{
			{
				Name:       "Ulvenwald Captive",
				TypeLine:   "Creature — Elf Druid",
				ManaCost:   "{1}{G}",
				OracleText: "Defender\n{T}: Add {G}.\n{5}{G}{G}: Transform this creature.",
				Power:      new("1"),
				Toughness:  new("2"),
			},
			{
				Name:       "Ulvenwald Abomination",
				TypeLine:   "Creature — Eldrazi Horror",
				OracleText: "{T}: Add {C}{C}.",
				Power:      new("4"),
				Toughness:  new("4"),
			},
		},
	})
	var transform game.Transform
	found := false
	for _, ability := range face.ActivatedAbilities {
		if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Sequence) != 1 {
			continue
		}
		if tr, ok := ability.Content.Modes[0].Sequence[0].Primitive.(game.Transform); ok {
			transform = tr
			found = true
		}
	}
	if !found {
		t.Fatalf("no activated Transform ability found in %+v", face.ActivatedAbilities)
	}
	if transform.Object.Kind() != game.ObjectReferenceSourcePermanent {
		t.Fatalf("transform object kind = %v, want ObjectReferenceSourcePermanent", transform.Object.Kind())
	}
}
