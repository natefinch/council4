package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
)

// TestLowerTriggeringEventCounterPlacement proves that a "put that many <kind>
// counters on <this creature|it>" trigger body reads the quantity its trigger
// measured: cards drawn or discarded, combat or noncombat damage dealt, or life
// gained or lost. "that many" lowers to the matching per-event DynamicAmount on
// the source or event permanent, never a fixed amount misread from the counter's
// "+1".
func TestLowerTriggeringEventCounterPlacement(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		typeLine    string
		oracleText  string
		dynamicKind game.DynamicAmountKind
		object      game.ObjectReference
		kind        counter.Kind
	}{
		{
			name:        "discard count on self",
			typeLine:    "Creature — Shark Pirate",
			oracleText:  "Whenever you discard one or more cards, put that many +1/+1 counters on this creature.",
			dynamicKind: game.DynamicAmountEventCardCount,
			object:      game.SourcePermanentReference(),
			kind:        counter.PlusOnePlusOne,
		},
		{
			name:        "combat damage on event permanent",
			typeLine:    "Creature — Vampire",
			oracleText:  "Whenever a creature you control deals combat damage to a player, put that many +1/+1 counters on it.",
			dynamicKind: game.DynamicAmountEventDamage,
			object:      game.EventPermanentReference(),
			kind:        counter.PlusOnePlusOne,
		},
		{
			name:        "life gained on self",
			typeLine:    "Creature — Elemental",
			oracleText:  "Whenever you gain life, put that many +1/+1 counters on this creature.",
			dynamicKind: game.DynamicAmountEventLifeChange,
			object:      game.SourcePermanentReference(),
			kind:        counter.PlusOnePlusOne,
		},
		{
			name:        "dealt damage on self via it",
			typeLine:    "Creature — Hydra",
			oracleText:  "Whenever this creature is dealt damage, put that many +1/+1 counters on it.",
			dynamicKind: game.DynamicAmountEventDamage,
			object:      game.EventPermanentReference(),
			kind:        counter.PlusOnePlusOne,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			power := "2"
			face := lowerSingleFace(t, &ScryfallCard{
				Name:       "Test Triggering Counter",
				Layout:     "normal",
				TypeLine:   test.typeLine,
				OracleText: test.oracleText,
				Power:      &power,
				Toughness:  &power,
			})
			if len(face.TriggeredAbilities) != 1 {
				t.Fatalf("triggered abilities = %d, want 1", len(face.TriggeredAbilities))
			}
			mode := face.TriggeredAbilities[0].Content.Modes[0]
			add, ok := mode.Sequence[0].Primitive.(game.AddCounter)
			if !ok {
				t.Fatalf("primitive = %T, want game.AddCounter", mode.Sequence[0].Primitive)
			}
			if add.Object != test.object {
				t.Fatalf("object = %#v, want %#v", add.Object, test.object)
			}
			if add.CounterKind != test.kind {
				t.Fatalf("counter kind = %v, want %v", add.CounterKind, test.kind)
			}
			got := add.Amount.DynamicAmount()
			if !got.Exists {
				t.Fatalf("amount = %#v, want dynamic kind %v", add.Amount, test.dynamicKind)
			}
			if got.Val.Kind != test.dynamicKind {
				t.Fatalf("dynamic kind = %v, want %v", got.Val.Kind, test.dynamicKind)
			}
			if got.Val.Multiplier != 1 {
				t.Fatalf("dynamic multiplier = %d, want 1", got.Val.Multiplier)
			}
		})
	}
}

// TestLowerTriggeringEventCounterPlacementFailsClosed proves that a "put that
// many <kind> counters" body fails closed when its enclosing trigger measures no
// quantity. The attack trigger publishes no per-event count, so "that many" has
// no referent and the placement stays unsupported rather than lowering to a
// wrong amount.
func TestLowerTriggeringEventCounterPlacementFailsClosed(t *testing.T) {
	t.Parallel()
	power := "2"
	_, diagnostics := lowerExecutableFaces(&ScryfallCard{
		Name:       "Test Unmeasured Trigger",
		Layout:     "normal",
		TypeLine:   "Creature — Warrior",
		OracleText: "Whenever this creature attacks, put that many +1/+1 counters on this creature.",
		Power:      &power,
		Toughness:  &power,
	})
	if len(diagnostics) == 0 {
		t.Fatal("lowered without diagnostics; expected fail-closed")
	}
}
