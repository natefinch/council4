package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestLowerAugurOfAutumnCovenCast(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Augur of Autumn",
		Layout:     "normal",
		TypeLine:   "Creature — Human Druid",
		ManaCost:   "{1}{G}{G}",
		OracleText: "You may look at the top card of your library any time.\nYou may play lands from the top of your library.\nCoven — As long as you control three or more creatures with different powers, you may cast creature spells from the top of your library.",
	})
	var found bool
	for i := range face.StaticAbilities {
		body := face.StaticAbilities[i].Body
		if !body.Condition.Exists || len(body.Condition.Val.Aggregates) != 1 {
			continue
		}
		condition := body.Condition.Val.Aggregates[0]
		for _, effect := range body.RuleEffects {
			if condition.Aggregate == game.AggregateControllerCreaturePowerDiversity &&
				condition.Op == compare.GreaterOrEqual && condition.Value == 3 &&
				effect.Kind == game.RuleEffectCastSpellsFromZone &&
				effect.CastFromZone == zone.Library && effect.TopCardOnly &&
				len(effect.SpellTypes) == 1 && effect.SpellTypes[0] == types.Creature {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("static abilities = %#v, want coven creature cast permission", face.StaticAbilities)
	}
}
