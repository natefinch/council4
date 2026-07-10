package parser

import "testing"

// phaseOutSentence returns the parsed sentence whose effects include an exact
// EffectPhaseOut, together with that effect. It fails the test when the source
// does not produce exactly one exact phase-out sentence. Targets live on the
// sentence (the phase-out recognizer covers the sentence and lets the shared
// target pass own the noun phrase), so callers assert cardinality on
// sentence.Targets.
func phaseOutSentence(t *testing.T, source string, context Context) (Sentence, EffectSyntax) {
	t.Helper()
	document, diagnostics := Parse(source, context)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %+v", diagnostics)
	}
	var matches []struct {
		sentence Sentence
		effect   EffectSyntax
	}
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectPhaseOut && effect.Exact {
					matches = append(matches, struct {
						sentence Sentence
						effect   EffectSyntax
					}{sentence, effect})
				}
			}
		}
	}
	if len(matches) != 1 {
		t.Fatalf("exact phase-out sentences = %d, want one", len(matches))
	}
	return matches[0].sentence, matches[0].effect
}

func TestParseAnyNumberTargetPhaseOut(t *testing.T) {
	t.Parallel()
	sentence, effect := phaseOutSentence(t,
		"Any number of target nonland permanents you control phase out.",
		Context{InstantOrSorcery: true},
	)
	if effect.Context != EffectContextController {
		t.Fatalf("effect = %+v, want controller context", effect)
	}
	if len(sentence.Targets) != 1 {
		t.Fatalf("targets = %+v, want one", sentence.Targets)
	}
	if sentence.Targets[0].Cardinality.Min != 0 || sentence.Targets[0].Cardinality.Max != 99 {
		t.Fatalf("cardinality = %+v, want any number (0..99)", sentence.Targets[0].Cardinality)
	}
}

func TestParseAnyNumberOfOtherTargetPhaseOut(t *testing.T) {
	t.Parallel()
	sentence, effect := phaseOutSentence(t,
		"When this creature enters, any number of other target creatures you control phase out.",
		Context{CardName: "Guardian of Faith"},
	)
	if effect.Context != EffectContextController {
		t.Fatalf("effect = %+v, want controller context", effect)
	}
	if len(sentence.Targets) != 1 {
		t.Fatalf("targets = %+v, want one", sentence.Targets)
	}
	if sentence.Targets[0].Cardinality.Min != 0 || sentence.Targets[0].Cardinality.Max != 99 {
		t.Fatalf("cardinality = %+v, want any number (0..99)", sentence.Targets[0].Cardinality)
	}
}

func TestParseSingleTargetPhaseOutIsExactController(t *testing.T) {
	t.Parallel()
	for name, source := range map[string]string{
		"bare creature":      "Target creature phases out.",
		"typed disjunction":  "Target artifact, creature, or land phases out.",
		"controlled subject": "Target creature you control phases out.",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			sentence, effect := phaseOutSentence(t, source, Context{InstantOrSorcery: true})
			if effect.Context != EffectContextController {
				t.Fatalf("effect = %+v, want controller context", effect)
			}
			if len(sentence.Targets) != 1 {
				t.Fatalf("targets = %+v, want one", sentence.Targets)
			}
		})
	}
}

// TestParsePluralPhaseOutWithoutTargetFailsClosed guards the Time and Tide
// regression: the plural "phase out" verb in a compound sentence that names no
// target must not be recognized as an exact phase-out effect, so it cannot lower
// lossily through the mass path. The mass form keeps its own strict recognizer.
func TestParsePluralPhaseOutWithoutTargetFailsClosed(t *testing.T) {
	t.Parallel()
	document, _ := Parse(
		"Simultaneously, all phased-out creatures phase in and all creatures with phasing phase out.",
		Context{InstantOrSorcery: true},
	)
	for _, ability := range document.Abilities {
		for _, sentence := range ability.Sentences {
			for _, effect := range sentence.Effects {
				if effect.Kind == EffectPhaseOut && effect.Exact {
					t.Fatalf("compound plural phase out produced an exact phase-out effect: %+v", effect)
				}
			}
		}
	}
}

// TestParseMassPhaseOutStillRecognized confirms the strict mass form still parses
// to an exact phase-out effect with no targets, unchanged by the targeted
// generalization.
func TestParseMassPhaseOutStillRecognized(t *testing.T) {
	t.Parallel()
	sentence, _ := phaseOutSentence(t,
		"All permanents you control phase out.",
		Context{InstantOrSorcery: true},
	)
	if len(sentence.Targets) != 0 {
		t.Fatalf("mass phase out targets = %+v, want none", sentence.Targets)
	}
}
