package parser

import "testing"

// tibaltsTrickeryOracle is the authoritative current Oracle text of Tibalt's
// Trickery ({1}{R} instant). The parser owns this wording; the compiler,
// lowering, and runtime consume only the typed nodes it produces.
const tibaltsTrickeryOracle = "Counter target spell. Choose 1, 2, or 3 at random. " +
	"Its controller mills that many cards, then exiles cards from the top of " +
	"their library until they exile a nonland card with a different name than " +
	"that spell. They may cast that card without paying its mana cost. Then " +
	"they put the exiled cards on the bottom of their library in a random order."

// collectEffects returns every parsed effect across all sentences of an ability
// in reading order, so a test can assert on the folded sequence the recognizer
// marked.
func collectEffects(ability *Ability) []*EffectSyntax {
	var effects []*EffectSyntax
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			effects = append(effects, &ability.Sentences[i].Effects[j])
		}
	}
	return effects
}

// TestParseTibaltsTrickerySequence verifies the recognizer marks the closed
// six-effect Tibalt's Trickery sequence: every effect is exact and folded into
// the TibaltsTrickery sequence, the head counter carries the inclusive random
// mill range [1, 3], exactly one zero-effect sentence is credited as the
// "Choose 1, 2, or 3 at random." mill-count prelude, and the whole ability is
// covered so no wording is left unrepresented.
func TestParseTibaltsTrickerySequence(t *testing.T) {
	t.Parallel()
	document, diagnostics := Parse(tibaltsTrickeryOracle,
		Context{CardName: "Tibalt's Trickery", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v, want none", diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("abilities = %d, want 1", len(document.Abilities))
	}
	ability := &document.Abilities[0]

	effects := collectEffects(ability)
	if len(effects) != 6 {
		t.Fatalf("effects = %d, want 6", len(effects))
	}
	wantKinds := []EffectKind{
		EffectCounter, EffectMill, EffectExile, EffectExile, EffectCast, EffectPut,
	}
	for i, effect := range effects {
		if effect.Kind != wantKinds[i] {
			t.Errorf("effect[%d].Kind = %v, want %v", i, effect.Kind, wantKinds[i])
		}
		if !effect.TibaltsTrickery {
			t.Errorf("effect[%d] (%v) not marked TibaltsTrickery", i, effect.Kind)
		}
		if !effect.Exact {
			t.Errorf("effect[%d] (%v) not marked Exact", i, effect.Kind)
		}
	}

	counter := effects[0]
	if counter.TibaltRandomMillMin != 1 || counter.TibaltRandomMillMax != 3 {
		t.Errorf("mill range = [%d, %d], want [1, 3]",
			counter.TibaltRandomMillMin, counter.TibaltRandomMillMax)
	}
	if counter.TibaltPreludeSpan.End.Offset <= counter.TibaltPreludeSpan.Start.Offset {
		t.Errorf("prelude span = %+v, want a non-empty credited span", counter.TibaltPreludeSpan)
	}

	preludes := 0
	for i := range ability.Sentences {
		if ability.Sentences[i].ChooseNumberAtRandomPrelude {
			preludes++
		}
	}
	if preludes != 1 {
		t.Errorf("credited random-choice preludes = %d, want exactly 1", preludes)
	}

	if report := AbilityCoverage(ability); !report.Complete || len(report.Uncovered) != 0 {
		t.Errorf("coverage incomplete: complete=%v blockers=%v uncovered=%v",
			report.Complete, report.Blockers, report.Uncovered)
	}
}

// TestParseTibaltsTrickeryFailsClosed verifies the recognizer refuses near-miss
// wording: a different-mana-value stop predicate, a fixed mill count with no
// random choose, and a mandatory (rather than optional) cast each leave every
// effect unmarked so the text-blind lowering never emits the Tibalt sequence
// for a card that does not match the exact template.
func TestParseTibaltsTrickeryFailsClosed(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		oracle string
	}{
		{
			name: "different mana value predicate",
			oracle: "Counter target spell. Choose 1, 2, or 3 at random. " +
				"Its controller mills that many cards, then exiles cards from the top of " +
				"their library until they exile a nonland card with a different mana value than " +
				"that spell. They may cast that card without paying its mana cost. Then " +
				"they put the exiled cards on the bottom of their library in a random order.",
		},
		{
			name: "fixed mill count without random choice",
			oracle: "Counter target spell. " +
				"Its controller mills three cards, then exiles cards from the top of " +
				"their library until they exile a nonland card with a different name than " +
				"that spell. They may cast that card without paying its mana cost. Then " +
				"they put the exiled cards on the bottom of their library in a random order.",
		},
		{
			name: "mandatory cast",
			oracle: "Counter target spell. Choose 1, 2, or 3 at random. " +
				"Its controller mills that many cards, then exiles cards from the top of " +
				"their library until they exile a nonland card with a different name than " +
				"that spell. They cast that card without paying its mana cost. Then " +
				"they put the exiled cards on the bottom of their library in a random order.",
		},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			document, _ := Parse(testCase.oracle,
				Context{CardName: "Almost Tibalt", InstantOrSorcery: true})
			for i := range document.Abilities {
				for _, effect := range collectEffects(&document.Abilities[i]) {
					if effect.TibaltsTrickery {
						t.Fatalf("near-miss %q marked effect %v as TibaltsTrickery",
							testCase.name, effect.Kind)
					}
				}
			}
		})
	}
}
