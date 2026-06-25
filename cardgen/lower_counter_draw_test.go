package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
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

// TestLowerLeavesBattlefieldCounterDraw proves Bloodtracker's "When this
// creature leaves the battlefield, draw a card for each +1/+1 counter on it."
// lowers to a leaves-the-battlefield trigger whose draw amount counts the
// triggering permanent's +1/+1 counters via DynamicAmountObjectCounters reading
// the event permanent's last-known information.
func TestLowerLeavesBattlefieldCounterDraw(t *testing.T) {
	t.Parallel()
	face := lowerSingleFace(t, &ScryfallCard{
		Name:       "Bloodtracker",
		Layout:     "normal",
		ManaCost:   "{3}{B}",
		TypeLine:   "Creature — Vampire Wizard",
		OracleText: "Flying\n{B}, Pay 2 life: Put a +1/+1 counter on this creature.\nWhen this creature leaves the battlefield, draw a card for each +1/+1 counter on it.",
		Power:      new("2"),
		Toughness:  new("2"),
	})
	if len(face.TriggeredAbilities) != 1 {
		t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
	}
	got := face.TriggeredAbilities[0]
	pattern := got.Trigger.Pattern
	if pattern.Event != game.EventZoneChanged {
		t.Errorf("Pattern.Event = %v, want EventZoneChanged", pattern.Event)
	}
	if pattern.Source != game.TriggerSourceSelf {
		t.Errorf("Pattern.Source = %v, want TriggerSourceSelf", pattern.Source)
	}
	if !pattern.MatchFromZone || pattern.FromZone != zone.Battlefield {
		t.Errorf("Pattern from-zone fields = %#v", pattern)
	}
	mode := got.Content.Modes[0]
	if len(mode.Sequence) != 1 {
		t.Fatalf("sequence length = %d, want 1", len(mode.Sequence))
	}
	draw, ok := mode.Sequence[0].Primitive.(game.Draw)
	if !ok {
		t.Fatalf("primitive = %T, want game.Draw", mode.Sequence[0].Primitive)
	}
	if draw.Player != game.ControllerReference() {
		t.Errorf("draw player = %#v, want ControllerReference", draw.Player)
	}
	dyn := draw.Amount.DynamicAmount()
	if !dyn.Exists || dyn.Val.Kind != game.DynamicAmountObjectCounters {
		t.Fatalf("draw amount = %#v, want DynamicAmountObjectCounters", draw.Amount)
	}
	if dyn.Val.CounterKind != counter.PlusOnePlusOne {
		t.Errorf("counted kind = %v, want PlusOnePlusOne", dyn.Val.CounterKind)
	}
	if dyn.Val.Object != game.EventPermanentReference() {
		t.Errorf("counted object = %#v, want EventPermanentReference", dyn.Val.Object)
	}
}
