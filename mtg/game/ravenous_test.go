package game

import (
	"reflect"
	"testing"

	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
)

func TestRavenousTemplatesComposeReplacementAndTrigger(t *testing.T) {
	replacement := RavenousEntersWithCountersReplacement()
	wantPlacement := CounterPlacement{Kind: counter.PlusOnePlusOne, AmountFromX: true}
	if got := replacement.Replacement.EntersWithCounters; !reflect.DeepEqual(got, []CounterPlacement{wantPlacement}) {
		t.Fatalf("counter placements = %#v, want %#v", got, []CounterPlacement{wantPlacement})
	}

	trigger := RavenousDrawTriggeredAbility()
	if trigger.Trigger.Pattern.Event != EventPermanentEnteredBattlefield ||
		trigger.Trigger.Pattern.Source != TriggerSourceSelf {
		t.Fatalf("trigger pattern = %#v, want self entry", trigger.Trigger.Pattern)
	}
	if !BodyHasKeyword(&trigger, Ravenous) {
		t.Fatal("Ravenous trigger does not carry Ravenous keyword identity")
	}
	if !trigger.Trigger.InterveningCondition.Exists {
		t.Fatal("Ravenous trigger has no X threshold condition")
	}
	aggregates := trigger.Trigger.InterveningCondition.Val.Aggregates
	if len(aggregates) != 1 ||
		aggregates[0].Aggregate != AggregateEventPermanentCastX ||
		aggregates[0].Op != compare.GreaterOrEqual ||
		aggregates[0].Value != 5 {
		t.Fatalf("Ravenous threshold = %#v, want event cast X >= 5", aggregates)
	}
}
