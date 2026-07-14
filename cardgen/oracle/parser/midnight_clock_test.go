package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestParseControllerHandAndGraveyardShuffle covers the reusable multi-zone
// shuffle Midnight Clock needs: "Shuffle your hand and graveyard into your
// library." is a controller-scoped shuffle whose two source zones cannot ride
// the single FromZone field, so it is carried by a dedicated flag.
func TestParseControllerHandAndGraveyardShuffle(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Shuffle your hand and graveyard into your library.", Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	effect := document.Abilities[0].Sentences[0].Effects[0]
	if effect.Kind != EffectShuffle ||
		!effect.Exact ||
		effect.Context != EffectContextController ||
		effect.ToZone != zone.Library ||
		!effect.ShuffleControllerHandAndGraveyardIntoLibrary {
		t.Fatalf("effect = %#v", effect)
	}
}

// TestParseTwelfthCounterThresholdTrigger covers the reusable Nth-counter
// threshold trigger: "When the twelfth hour counter is put on this artifact,
// ..." parses to a self-sourced counter-added trigger keyed to the hour counter
// kind with a threshold of twelve.
func TestParseTwelfthCounterThresholdTrigger(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("When the twelfth hour counter is put on this artifact, draw a card.", Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	trigger := document.Abilities[0].Trigger
	if trigger == nil || trigger.TriggerEvent == nil {
		t.Fatalf("ability has no counter trigger event: %#v", document.Abilities[0])
	}
	event := trigger.TriggerEvent
	if event.Kind != TriggerEventKindCounterAdded {
		t.Fatalf("event kind = %v, want CounterAdded", event.Kind)
	}
	if event.Counter.Threshold != 12 {
		t.Fatalf("counter threshold = %d, want 12", event.Counter.Threshold)
	}
	if !event.Counter.Known || event.Counter.Kind != counter.Hour {
		t.Fatalf("counter = (known=%v, kind=%v), want (true, hour)", event.Counter.Known, event.Counter.Kind)
	}
}
