package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
)

// TestLowerMoreThanMeetsTheEye proves the Transformers "More Than Meets the Eye"
// keyword lowers to a single alternative spell cost carrying the More Than Meets
// the Eye mechanic and its fixed mana cost. The rules layer reads that mechanic
// to make the resulting permanent enter converted, as its back face.
func TestLowerMoreThanMeetsTheEye(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:      "Test Bot",
		Layout:    "transform",
		TypeLine:  "Legendary Artifact Creature — Robot",
		ManaCost:  "{4}",
		Power:     new("3"),
		Toughness: new("3"),
		OracleText: "More Than Meets the Eye {R}{W} (You may cast this card converted for {R}{W}.)\n" +
			"Flying",
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alt := face.AlternativeCosts[0]
	if alt.Mechanic != cost.AlternativeMechanicMoreThanMeetsTheEye {
		t.Fatalf("mechanic = %v, want AlternativeMechanicMoreThanMeetsTheEye", alt.Mechanic)
	}
	if !alt.ManaCost.Exists {
		t.Fatal("alternative cost has no mana cost")
	}
	want := cost.Mana{cost.R, cost.W}
	if alt.ManaCost.Val.String() != want.String() {
		t.Fatalf("alternative mana cost = %s, want %s", alt.ManaCost.Val.String(), want.String())
	}
	if alt.Label != "More Than Meets the Eye" {
		t.Fatalf("label = %q, want More Than Meets the Eye", alt.Label)
	}
}
