package parser

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/counter"
)

// TestParseRemoveCounterAndSacrificeSelfCost covers the Quest/Expedition family's
// combined "Remove N <kind> counters from this <permanent> and sacrifice it" cost.
// The single comma-delimited phrase splits into two typed components: a
// self-source counter removal and a self-source sacrifice, so each half lowers
// independently.
func TestParseRemoveCounterAndSacrificeSelfCost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		source      string
		amount      int
		counterKind counter.Kind
	}{
		{
			name:        "quest counters from this enchantment",
			source:      "Remove five quest counters from this enchantment and sacrifice it: Draw a card.",
			amount:      5,
			counterKind: counter.Quest,
		},
		{
			name:        "single quest counter",
			source:      "Remove a quest counter from this enchantment and sacrifice it: Draw a card.",
			amount:      1,
			counterKind: counter.Quest,
		},
		{
			name:        "pressure counters from this land",
			source:      "{1}{R}, {T}, Remove two pressure counters from this land and sacrifice it: Draw a card.",
			amount:      2,
			counterKind: counter.Pressure,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := Parse(test.source, Context{})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			if len(document.Abilities) != 1 {
				t.Fatalf("abilities = %d, want 1", len(document.Abilities))
			}
			syntax := document.Abilities[0].CostSyntax
			if syntax == nil {
				t.Fatal("cost syntax = nil")
			}
			var remove, sacrifice *CostComponent
			for i := range syntax.Components {
				switch syntax.Components[i].Kind {
				case CostComponentRemoveCounter:
					remove = &syntax.Components[i]
				case CostComponentSacrifice:
					sacrifice = &syntax.Components[i]
				default:
				}
			}
			if remove == nil {
				t.Fatalf("components = %#v, want a remove-counter component", syntax.Components)
			}
			if !remove.SourceSelf {
				t.Fatalf("remove component = %#v, want source-self", remove)
			}
			if !remove.AmountKnown || remove.AmountValue != test.amount {
				t.Fatalf("remove amount = (%d, %v), want %d", remove.AmountValue, remove.AmountKnown, test.amount)
			}
			if !remove.CounterKindKnown || remove.CounterKind != test.counterKind {
				t.Fatalf("remove counter = (%v, %v), want %v", remove.CounterKind, remove.CounterKindKnown, test.counterKind)
			}
			if sacrifice == nil {
				t.Fatalf("components = %#v, want a sacrifice component", syntax.Components)
			}
			if !sacrifice.SourceSelf {
				t.Fatalf("sacrifice component = %#v, want source-self", sacrifice)
			}
		})
	}
}

// TestParseSacrificeItNotSplitWithoutLeadingCost verifies the split only applies
// when a leading cost precedes the "and sacrifice it" clause. A bare "Sacrifice
// it" is not a recognized self-sacrifice and must not be silently promoted.
func TestParseSacrificeItNotSplitWithoutLeadingCost(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse("Sacrifice a creature and a Forest: Draw a card.", Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	syntax := document.Abilities[0].CostSyntax
	if syntax == nil || len(syntax.Components) != 1 {
		t.Fatalf("components = %#v, want a single unsplit sacrifice component", syntax)
	}
	if syntax.Components[0].Kind != CostComponentSacrifice {
		t.Fatalf("kind = %v, want sacrifice", syntax.Components[0].Kind)
	}
}
