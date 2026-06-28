package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
)

// TestLowerCastTriggerManaSpentCondition proves a cast trigger gated by an
// intervening "if ... mana was spent to cast it" clause lowers its condition to
// an AggregateEventSpellManaSpentToCast comparison, and that stripping the
// intervening clause from the body parse leaves the effect intact (a Treasure
// token, as on Prompto Argentum). "at least N" maps to GreaterOrEqual N and
// "no mana" maps to LessOrEqual 0.
func TestLowerCastTriggerManaSpentCondition(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		clause  string
		wantOp  compare.Op
		wantVal int
	}{
		{"at least four", "if at least four mana was spent to cast it", compare.GreaterOrEqual, 4},
		{"at least eight that spell", "if at least eight mana was spent to cast that spell", compare.GreaterOrEqual, 8},
		{"no mana", "if no mana was spent to cast it", compare.LessOrEqual, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Caster",
				Layout:     "normal",
				TypeLine:   "Creature — Human Wizard",
				OracleText: "Whenever you cast a noncreature spell, " + tc.clause + ", create a Treasure token.",
				Power:      new("2"),
				Toughness:  new("2"),
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			ta := face.TriggeredAbilities[0]
			if ta.Trigger.Pattern.Event != game.EventSpellCast {
				t.Fatalf("event = %v, want EventSpellCast", ta.Trigger.Pattern.Event)
			}
			if !ta.Trigger.InterveningCondition.Exists {
				t.Fatalf("intervening condition missing; trigger = %+v", ta.Trigger)
			}
			aggregates := ta.Trigger.InterveningCondition.Val.Aggregates
			if len(aggregates) != 1 {
				t.Fatalf("aggregates = %#v, want exactly one", aggregates)
			}
			got := aggregates[0]
			if got.Aggregate != game.AggregateEventSpellManaSpentToCast {
				t.Errorf("aggregate = %v, want AggregateEventSpellManaSpentToCast", got.Aggregate)
			}
			if got.Op != tc.wantOp || got.Value != tc.wantVal {
				t.Errorf("comparison = {Op:%v Value:%d}, want {Op:%v Value:%d}", got.Op, got.Value, tc.wantOp, tc.wantVal)
			}
			if len(ta.Content.Modes) != 1 {
				t.Fatalf("content modes = %d, want 1 (effect must survive the intervening strip)", len(ta.Content.Modes))
			}
			if len(ta.Content.Modes[0].Sequence) == 0 {
				t.Fatal("effect sequence empty; intervening strip dropped the body")
			}
		})
	}
}
