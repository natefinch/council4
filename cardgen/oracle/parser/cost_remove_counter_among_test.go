package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

func TestParseRemoveCounterAmongCostObject(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		source      string
		wantAmount  int
		wantFromX   bool
		wantKind    counter.Kind
		wantKnown   bool
		wantNoun    ObjectNoun
		wantControl ControllerRelation
	}{
		{
			name:        "two +1/+1 from among creatures",
			source:      "Remove two +1/+1 counters from among creatures you control: Draw a card.",
			wantAmount:  2,
			wantKind:    counter.PlusOnePlusOne,
			wantKnown:   true,
			wantNoun:    ObjectNounCreature,
			wantControl: ControllerRelationYouControl,
		},
		{
			name:        "X +1/+1 from among creatures",
			source:      "Remove X +1/+1 counters from among creatures you control: Draw a card.",
			wantFromX:   true,
			wantKind:    counter.PlusOnePlusOne,
			wantKnown:   true,
			wantNoun:    ObjectNounCreature,
			wantControl: ControllerRelationYouControl,
		},
		{
			name:        "two +1/+1 from among artifacts",
			source:      "Remove two +1/+1 counters from among artifacts you control: Draw a card.",
			wantAmount:  2,
			wantKind:    counter.PlusOnePlusOne,
			wantKnown:   true,
			wantNoun:    ObjectNounArtifact,
			wantControl: ControllerRelationYouControl,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentRemoveCounter {
				t.Fatalf("kind = %v, want remove counter", component.Kind)
			}
			if !component.RemoveCounterAmong {
				t.Fatalf("component = %#v, want among removal", component)
			}
			if component.SourceSelf {
				t.Fatalf("component = %#v, want not source-self", component)
			}
			if component.AmountFromX != test.wantFromX ||
				(!test.wantFromX && (!component.AmountKnown || component.AmountValue != test.wantAmount)) {
				t.Fatalf("amount = (%d known %t fromX %t), want %d fromX %t",
					component.AmountValue, component.AmountKnown, component.AmountFromX, test.wantAmount, test.wantFromX)
			}
			if component.CounterKindKnown != test.wantKnown || component.CounterKind != test.wantKind {
				t.Fatalf("counter = (%v known %t), want %v known %t",
					component.CounterKind, component.CounterKindKnown, test.wantKind, test.wantKnown)
			}
			if component.ObjectNoun != test.wantNoun || component.ObjectController != test.wantControl {
				t.Fatalf("object = (%v, %v), want (%v, %v)",
					component.ObjectNoun, component.ObjectController, test.wantNoun, test.wantControl)
			}
		})
	}
}

func TestParseRemoveCounterAmongUnkindedRecognizedWithoutKind(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t, "Remove three counters from among creatures you control: Draw a card.")
	if component.Kind != CostComponentRemoveCounter || !component.RemoveCounterAmong {
		t.Fatalf("component = %#v, want among removal", component)
	}
	if component.CounterKindKnown {
		t.Fatalf("component = %#v, want no known counter kind for unspecified counters", component)
	}
	if !component.AmountKnown || component.AmountValue != 3 {
		t.Fatalf("amount = (%d known %t), want 3", component.AmountValue, component.AmountKnown)
	}
}

func TestParseRemoveCounterFromPermanentYouControl(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		source    string
		wantNoun  ObjectNoun
		wantKind  counter.Kind
		wantKnown bool
	}{
		{
			name:     "any counter from a permanent",
			source:   "Remove a counter from a permanent you control: Draw a card.",
			wantNoun: ObjectNounPermanent,
		},
		{
			name:      "+1/+1 counter from a permanent",
			source:    "Remove a +1/+1 counter from a permanent you control: Draw a card.",
			wantNoun:  ObjectNounPermanent,
			wantKind:  counter.PlusOnePlusOne,
			wantKnown: true,
		},
		{
			name:     "any counter from a creature",
			source:   "Remove a counter from a creature you control: Draw a card.",
			wantNoun: ObjectNounCreature,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			component := soleCostComponent(t, test.source)
			if component.Kind != CostComponentRemoveCounter || !component.RemoveCounterAmong {
				t.Fatalf("component = %#v, want among removal", component)
			}
			if component.SourceSelf {
				t.Fatalf("component = %#v, want not source-self", component)
			}
			if !component.AmountKnown || component.AmountValue != 1 {
				t.Fatalf("amount = (%d known %t), want 1", component.AmountValue, component.AmountKnown)
			}
			if component.CounterKindKnown != test.wantKnown || component.CounterKind != test.wantKind {
				t.Fatalf("counter = (%v known %t), want %v known %t",
					component.CounterKind, component.CounterKindKnown, test.wantKind, test.wantKnown)
			}
			if component.ObjectNoun != test.wantNoun ||
				component.ObjectController != ControllerRelationYouControl {
				t.Fatalf("object = (%v, %v), want (%v, you control)",
					component.ObjectNoun, component.ObjectController, test.wantNoun)
			}
		})
	}
}

func TestParseRemoveCounterFromSourceStillRecognized(t *testing.T) {
	t.Parallel()
	component := soleCostComponent(t, "Remove a +1/+1 counter from this creature: Draw a card.")
	if component.Kind != CostComponentRemoveCounter || component.RemoveCounterAmong || !component.SourceSelf {
		t.Fatalf("component = %#v, want single-source self removal", component)
	}
	if !component.CounterKindKnown || component.CounterKind != counter.PlusOnePlusOne {
		t.Fatalf("counter = (%v known %t), want +1/+1", component.CounterKind, component.CounterKindKnown)
	}
}
