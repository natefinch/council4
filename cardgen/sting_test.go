package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerStingConditionalFirstStrike(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Sting, the Glinting Dagger",
		Layout:     "normal",
		TypeLine:   "Legendary Artifact — Equipment",
		ManaCost:   "{2}",
		OracleText: "Equipped creature gets +1/+1 and has haste.\nAt the beginning of each combat, untap equipped creature.\nEquipped creature has first strike as long as it's blocking or blocked by a Goblin or Orc.\nEquip {2}",
	})
	var found bool
	for i := range face.StaticAbilities {
		body := face.StaticAbilities[i].Body
		if !body.Condition.Exists ||
			body.Condition.Val.SourceAttachedCombatCounterpartSubtypes != [2]types.Sub{types.Goblin, types.Orc} {
			continue
		}
		for _, effect := range body.ContinuousEffects {
			if effect.Layer == game.LayerAbility &&
				len(effect.AddKeywords) == 1 && effect.AddKeywords[0] == game.FirstStrike &&
				!effect.Group.Empty() {
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("static abilities = %#v, want conditional attached first strike", face.StaticAbilities)
	}
}
