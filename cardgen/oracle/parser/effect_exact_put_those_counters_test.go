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

// TestExactPutItsCountersAccepts covers the singular-pronoun counter-salvage
// form "put its counters on <dest>", which names a triggering permanent's
// counters with "its" instead of "those" and round-trips to the same
// MoveThoseCounters salvage effect, for both a target-creature destination (Star
// Pupil) and a self destination (Buzzard-Wasp Colony).
func TestExactPutItsCountersAccepts(t *testing.T) {
	t.Parallel()
	cases := []struct {
		source string
		card   string
	}{
		{
			"When this creature dies, put its counters on target creature you control.",
			"Star Pupil",
		},
		{
			"Whenever another creature you control dies, if it had counters on it, put its counters on this creature.",
			"Buzzard-Wasp Colony",
		},
	}
	for _, tc := range cases {
		effect := putThoseCountersEffect(t, tc.source, tc.card)
		if !effect.MoveThoseCounters {
			t.Errorf("MoveThoseCounters(%q) = false, want true", tc.source)
		}
		if !effect.Exact {
			t.Errorf("Exact(%q) = false, want true", tc.source)
		}
	}
}

// TestExactPutItsNamedKindCountersFailsClosed verifies the kind-named singular
// form "put its +1/+1 counters on <dest>" stays out of the kind-agnostic salvage
// move (its all-of-one-kind semantics are not modeled), so MoveThoseCounters is
// not set.
func TestExactPutItsNamedKindCountersFailsClosed(t *testing.T) {
	t.Parallel()
	effect := putThoseCountersEffect(
		t,
		"When this creature leaves the battlefield, put its +1/+1 counters on target creature you control.",
		"Selfless Police Captain",
	)
	if effect.MoveThoseCounters {
		t.Error("MoveThoseCounters = true, want false for kind-named salvage")
	}
}
