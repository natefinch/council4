package parser

import "testing"

func putThoseCountersEffect(t *testing.T, source, cardName string) EffectSyntax {
	t.Helper()
	document, diagnostics := Parse(source, Context{CardName: cardName})
	if len(diagnostics) != 0 {
		t.Fatalf("Parse(%q) diagnostics = %#v", source, diagnostics)
	}
	var found *EffectSyntax
	for a := range document.Abilities {
		ability := &document.Abilities[a]
		for s := range ability.Sentences {
			sentence := &ability.Sentences[s]
			for i := range sentence.Effects {
				if sentence.Effects[i].Kind == EffectPut {
					found = &sentence.Effects[i]
				}
			}
		}
	}
	if found == nil {
		t.Fatalf("Parse(%q) found no EffectPut: %#v", source, document.Abilities)
	}
	return *found
}

// TestExactPutThoseCountersAccepts covers the counter-salvage form "put those
// counters on <dest>" reached after an intervening "if it had counters on it"
// clause, for both a target-creature destination (Iron Apprentice) and a
// self-name destination (The Ozolith).
func TestExactPutThoseCountersAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		card   string
	}{
		{
			"When this creature dies, if it had counters on it, put those counters on target creature you control.",
			"Iron Apprentice",
		},
		{
			"Whenever a creature you control leaves the battlefield, if it had counters on it, put those counters on The Ozolith.",
			"The Ozolith",
		},
	}
	for _, tc := range cases {
		effect := putThoseCountersEffect(t, tc.source, tc.card)
		if !effect.MoveThoseCounters {
			t.Errorf("MoveThoseCounters(%q) = false, want true", tc.source)
		}
	}
}
