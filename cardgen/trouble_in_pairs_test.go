package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerTroubleInPairs(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Trouble in Pairs",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{2}{W}{W}",
		OracleText: "If an opponent would begin an extra turn, that player skips that turn instead.\nWhenever an opponent attacks you with two or more creatures, draws their second card each turn, or casts their second spell each turn, you draw a card.",
	})
	if len(face.StaticAbilities) != 1 || len(face.TriggeredAbilities) != 3 {
		t.Fatalf("face = %#v", face)
	}
	rule := face.StaticAbilities[0].Body.RuleEffects[0]
	if rule.Kind != game.RuleEffectSkipExtraTurns || rule.AffectedPlayer != game.PlayerOpponent {
		t.Fatalf("rule = %#v", rule)
	}
	attack := face.TriggeredAbilities[0].Trigger.Pattern
	if attack.Event != game.EventAttackerDeclared ||
		attack.Controller != game.TriggerControllerOpponent ||
		attack.Player != game.TriggerPlayerYou ||
		attack.AttackerCountAtLeast != 2 ||
		!attack.OneOrMore {
		t.Fatalf("attack trigger = %#v", attack)
	}
	if draw := face.TriggeredAbilities[1].Trigger.Pattern; draw.Event != game.EventCardDrawn ||
		draw.Player != game.TriggerPlayerOpponent ||
		draw.Controller != game.TriggerControllerAny ||
		draw.PlayerEventOrdinalThisTurn != 2 {
		t.Fatalf("draw trigger = %#v", draw)
	}
	if cast := face.TriggeredAbilities[2].Trigger.Pattern; cast.Event != game.EventSpellCast || cast.PlayerEventOrdinalThisTurn != 2 {
		t.Fatalf("cast trigger = %#v", cast)
	}
}
