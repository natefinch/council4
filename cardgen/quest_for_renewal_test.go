package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestLowerQuestForRenewalConditionalUntap(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Quest for Renewal",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{1}{G}",
		OracleText: "Whenever a creature you control becomes tapped, you may put a quest counter on this enchantment.\nAs long as there are four or more quest counters on this enchantment, untap all creatures you control during each other player's untap step.",
	})
	if len(face.StaticAbilities) != 1 {
		t.Fatalf("static abilities = %#v, want one", face.StaticAbilities)
	}
	body := face.StaticAbilities[0].Body
	if !body.Condition.Exists || !body.Condition.Val.SourceCounterKindKnown ||
		body.Condition.Val.SourceCounterKind != counter.Quest ||
		body.Condition.Val.SourceCountersAtLeast != 4 {
		t.Fatalf("condition = %#v, want four quest counters", body.Condition)
	}
	effect := body.RuleEffects[0]
	if effect.Kind != game.RuleEffectUntapDuringOtherPlayersUntapStep ||
		effect.AffectedController != game.ControllerYou ||
		len(effect.PermanentTypes) != 1 || effect.PermanentTypes[0] != types.Creature {
		t.Fatalf("rule effect = %#v, want controller creatures extra untap", effect)
	}
}
