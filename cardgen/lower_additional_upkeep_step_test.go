package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerAdditionalUpkeepStepParadoxHaze proves that Paradox Haze lowers to an
// Enchant player static ability plus a triggered ability whose beginning-of-step
// pattern is scoped to the enchanted player's first upkeep each turn and whose
// body is a single AddExtraUpkeepStep primitive.
func TestLowerAdditionalUpkeepStepParadoxHaze(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Paradox Haze",
		Layout:     "normal",
		TypeLine:   "Enchantment — Aura",
		ManaCost:   "{2}{U}",
		OracleText: "Enchant player\nAt the beginning of enchanted player's first upkeep each turn, that player gets an additional upkeep step after this step.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	trigger := face.TriggeredAbilities[0].Trigger
	if trigger.Type != game.TriggerAt {
		t.Fatalf("trigger type = %v, want TriggerAt", trigger.Type)
	}
	pattern := trigger.Pattern
	if pattern.Event != game.EventBeginningOfStep {
		t.Fatalf("pattern event = %v, want EventBeginningOfStep", pattern.Event)
	}
	if pattern.Step != game.StepUpkeep {
		t.Fatalf("pattern step = %v, want StepUpkeep", pattern.Step)
	}
	if !pattern.StepPlayerIsSourceEnchantedPlayer {
		t.Fatal("pattern must be scoped to the enchanted player's step")
	}
	if !pattern.FirstUpkeepStepEachTurn {
		t.Fatal("pattern must be gated on the first upkeep step each turn")
	}
	seq := face.TriggeredAbilities[0].Content.Modes[0].Sequence
	if len(seq) != 1 {
		t.Fatalf("sequence = %#v, want one instruction", seq)
	}
	if _, ok := seq[0].Primitive.(game.AddExtraUpkeepStep); !ok {
		t.Fatalf("primitive = %T, want game.AddExtraUpkeepStep", seq[0].Primitive)
	}

	if len(face.StaticAbilities) == 0 {
		t.Fatal("expected an Enchant player static ability")
	}
	var enchant bool
	for _, static := range face.StaticAbilities {
		for _, kw := range static.Body.KeywordAbilities {
			if _, ok := kw.(game.EnchantKeyword); ok {
				enchant = true
			}
		}
	}
	if !enchant {
		t.Fatalf("static abilities = %#v, want an Enchant keyword", face.StaticAbilities)
	}
}
