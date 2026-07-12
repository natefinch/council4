package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestLowerFightRiggingHideawayPlay(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Fight Rigging",
		Layout:     "normal",
		TypeLine:   "Enchantment",
		ManaCost:   "{2}{G}",
		OracleText: "Hideaway 5 (When this enchantment enters, look at the top five cards of your library, exile one face down, then put the rest on the bottom in a random order.)\nAt the beginning of combat on your turn, put a +1/+1 counter on target creature you control. Then if you control a creature with power 7 or greater, you may play the exiled card without paying its mana cost.",
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want hideaway and combat", len(face.TriggeredAbilities))
	}
	combat := face.TriggeredAbilities[1]
	if combat.Trigger.Pattern.Event != game.EventBeginningOfStep ||
		combat.Trigger.Pattern.Step != game.StepBeginningOfCombat {
		t.Fatalf("trigger = %#v", combat.Trigger.Pattern)
	}
	mode := combat.Content.Modes[0]
	if len(mode.Targets) != 1 || len(mode.Sequence) != 2 {
		t.Fatalf("mode = %#v", mode)
	}
	if _, ok := mode.Sequence[0].Primitive.(game.AddCounter); !ok {
		t.Fatalf("first instruction = %#v", mode.Sequence[0])
	}
	if _, ok := mode.Sequence[1].Primitive.(game.PlayHideawayCard); !ok ||
		!mode.Sequence[1].Optional ||
		!mode.Sequence[1].Condition.Exists ||
		!mode.Sequence[1].Condition.Val.Condition.Exists ||
		!mode.Sequence[1].Condition.Val.Condition.Val.ControlsMatching.Exists {
		t.Fatalf("hideaway play = %#v", mode.Sequence[1])
	}
}
