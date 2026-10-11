package parser

import "testing"

func TestExploreRepeatSyntax(t *testing.T) {
	t.Parallel()
	for _, subject := range []string{"It", "This creature", "Test Scout", "Target creature you control"} {
		for _, count := range []struct {
			text  string
			value int
			x     bool
		}{
			{"once", 1, false},
			{"one times", 1, false},
			{"twice", 2, false},
			{"two times", 2, false},
			{"zero times", 0, false},
			{"X times", 0, true},
		} {
			t.Run(subject+"/"+count.text, func(t *testing.T) {
				t.Parallel()
				doc, diagnostics := Parse("When this creature enters, "+subject+" explores "+count.text+".", Context{CardName: "Test Scout"})
				if len(diagnostics) != 0 {
					t.Fatalf("diagnostics = %+v", diagnostics)
				}
				effects := doc.Abilities[0].Sentences[0].Effects
				if len(effects) != 1 || effects[0].Kind != EffectExplore || !effects[0].Exact {
					t.Fatalf("effect count=%d, want one exact Explore for %s/%s", len(effects), subject, count.text)
				}
				amount := effects[0].Amount
				if amount.Known == count.x || amount.VariableX != count.x || amount.Value != count.value {
					t.Fatalf("amount = %+v, want fixed %d or X=%v", amount, count.value, count.x)
				}
				if amount.Span.Start.Offset == amount.Span.End.Offset {
					t.Fatal("repeat count lost its original source span")
				}
			})
		}
	}
}

func TestExploreRepeatSyntaxRefusesUnknownSuffix(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"It explores -2 times.", "It explores Y times.", "It explores up to two times.",
		"It explores two times and dances.", "It explores X plus one times.",
		"They explores two times.", "Each creature you control explores two times.",
	} {
		t.Run(text, func(t *testing.T) {
			t.Parallel()
			doc, _ := Parse("When this creature enters, "+text, Context{CardName: "Test Scout"})
			for _, ability := range doc.Abilities {
				for _, sentence := range ability.Sentences {
					for _, effect := range sentence.Effects {
						if effect.Kind == EffectExplore && effect.Exact {
							t.Fatalf("unexpected exact Explore: %+v", effect)
						}
					}
				}
			}
		})
	}
}
