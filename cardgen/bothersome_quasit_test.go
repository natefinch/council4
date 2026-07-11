package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerBothersomeQuasitGoadedOpponentsCantBlock(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Bothersome Quasit",
		Layout:     "normal",
		TypeLine:   "Creature — Demon",
		ManaCost:   "{2}{R}",
		OracleText: "Menace\nGoaded creatures your opponents control can't block.\nWhenever you cast a noncreature spell, goad target creature an opponent controls.",
	})
	var found bool
	for i := range face.StaticAbilities {
		for _, effect := range face.StaticAbilities[i].Body.RuleEffects {
			if effect.Kind == game.RuleEffectCantBlock &&
				effect.AffectedController == game.ControllerOpponent &&
				effect.AffectedSelection.MatchGoaded &&
				len(effect.PermanentTypes) == 1 &&
				effect.PermanentTypes[0] == types.Creature {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("static abilities = %#v, want goaded opponent creature restriction", face.StaticAbilities)
	}
}
