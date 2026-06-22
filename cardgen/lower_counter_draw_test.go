package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestLowerCounterPlacementDrawSelf(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Exemplar of Light",
		Layout:     "normal",
		ManaCost:   "{2}{W}{W}",
		TypeLine:   "Creature — Angel",
		OracleText: "Flying\nWhenever you gain life, put a +1/+1 counter on this creature.\nWhenever you put one or more +1/+1 counters on this creature, draw a card. This ability triggers only once each turn.",
		Power:      new("3"),
		Toughness:  new("3"),
	})
	if len(face.TriggeredAbilities) != 2 {
		t.Fatalf("triggered abilities = %d, want 2", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[1]
	pattern := got.Trigger.Pattern
	if pattern.Event != game.EventCountersAdded {
		t.Errorf("Pattern.Event = %v, want EventCountersAdded", pattern.Event)
	}
	if pattern.Source != game.TriggerSourceSelf {
		t.Errorf("Pattern.Source = %v, want TriggerSourceSelf", pattern.Source)
	}
	if pattern.CauseController != game.TriggerControllerYou {
		t.Errorf("Pattern.CauseController = %v, want TriggerControllerYou", pattern.CauseController)
	}
	if !pattern.OneOrMore || !pattern.MatchCounterKind || pattern.CounterKind != counter.PlusOnePlusOne {
		t.Errorf("Pattern counter fields = %#v", pattern)
	}
	if got.MaxTriggersPerTurn != 1 {
		t.Errorf("MaxTriggersPerTurn = %d, want 1", got.MaxTriggersPerTurn)
	}
}

func TestLowerCounterPlacementDrawThatMany(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Terrasymbiosis",
		Layout:     "normal",
		ManaCost:   "{2}{G}",
		TypeLine:   "Enchantment",
		OracleText: "Whenever you put one or more +1/+1 counters on a creature you control, you may draw that many cards. Do this only once each turn.",
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	pattern := got.Trigger.Pattern
	if pattern.Event != game.EventCountersAdded {
		t.Errorf("Pattern.Event = %v, want EventCountersAdded", pattern.Event)
	}
	if pattern.Controller != game.TriggerControllerYou {
		t.Errorf("Pattern.Controller = %v, want TriggerControllerYou", pattern.Controller)
	}
	if pattern.CauseController != game.TriggerControllerYou {
		t.Errorf("Pattern.CauseController = %v, want TriggerControllerYou", pattern.CauseController)
	}
	if !got.Optional {
		t.Error("Optional = false, want true")
	}
	if got.MaxTriggersPerTurn != 1 {
		t.Errorf("MaxTriggersPerTurn = %d, want 1", got.MaxTriggersPerTurn)
	}
	mode := got.Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %T, want game.Draw", mode.Sequence[0].Primitive)
	}
	dyn := draw.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountEventCounterCount {
		t.Errorf("draw amount = %#v, want DynamicAmountEventCounterCount", draw.Amount)
	}
}
