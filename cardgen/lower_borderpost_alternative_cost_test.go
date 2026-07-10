package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerBorderpostAlternativeCost(t *testing.T) {
	t.Parallel()
	const alternativeText = "You may pay {1} and return a basic land you control to its owner's hand rather than pay this spell's mana cost."
	face := lowerSingleFace(t, &ScryfallCard{
		Name:     "Mistvein Borderpost",
		Layout:   "normal",
		TypeLine: "Artifact",
		ManaCost: "{1}{U}{B}",
		OracleText: alternativeText + "\n" +
			"This artifact enters tapped.\n{T}: Add {U} or {B}.",
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one", face.AlternativeCosts)
	}
	alternative := face.AlternativeCosts[0]
	if !alternative.ManaCost.Exists || alternative.ManaCost.Val.String() != "{1}" {
		t.Fatalf("mana cost = %#v, want {1}", alternative.ManaCost)
	}
	if len(alternative.AdditionalCosts) != 1 {
		t.Fatalf("additional costs = %#v, want one", alternative.AdditionalCosts)
	}
	additional := alternative.AdditionalCosts[0]
	if additional.Kind != cost.AdditionalReturnToHand ||
		additional.Amount != 1 ||
		!additional.MatchPermanentType ||
		additional.PermanentType != types.Land ||
		additional.RequireSupertype != types.Basic {
		t.Fatalf("additional cost = %#v, want one basic land returned", additional)
	}
}
