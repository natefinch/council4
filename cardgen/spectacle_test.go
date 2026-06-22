package cardgen

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game/cost"
)

const skewerTheCriticsText = `Spectacle {R} (You may cast this spell for its spectacle cost rather than its mana cost if an opponent lost life this turn.)
Skewer the Critics deals 3 damage to any target.`

func TestLowerSkewerTheCriticsSpectacle(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Skewer the Critics",
		Layout:     "normal",
		TypeLine:   "Sorcery",
		ManaCost:   "{2}{R}",
		OracleText: skewerTheCriticsText,
	})
	if len(face.AlternativeCosts) != 1 {
		t.Fatalf("alternative costs = %#v, want one Spectacle cost", face.AlternativeCosts)
	}
	alternative := face.AlternativeCosts[0]
	if alternative.Label != "Spectacle" ||
		alternative.Condition != cost.AlternativeConditionOpponentLostLifeThisTurn ||
		!alternative.ManaCost.Exists ||
		!slices.Equal(alternative.ManaCost.Val, cost.Mana{cost.R}) {
		t.Fatalf("alternative = %#v, want Spectacle {R} conditioned on opponent lost life", alternative)
	}
}
