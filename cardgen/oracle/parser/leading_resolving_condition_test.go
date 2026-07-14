package parser

import "testing"

// firstModifyPTEffect parses source expected to yield a single ability and
// returns its first EffectModifyPT effect, so tests assert on the pump clause's
// recovered subject and exactness rather than on source text.
func firstModifyPTEffect(t *testing.T, source string) (EffectSyntax, bool) {
	t.Helper()
	document, diagnostics := Parse(source, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("source %q diagnostics = %#v", source, diagnostics)
	}
	if len(document.Abilities) != 1 {
		t.Fatalf("source %q abilities = %d, want one", source, len(document.Abilities))
	}
	for _, sentence := range document.Abilities[0].Sentences {
		for _, effect := range sentence.Effects {
			if effect.Kind == EffectModifyPT {
				return effect, true
			}
		}
	}
	return EffectSyntax{}, false
}

// TestStripLeadingResolvingGateRecoversGroupPumpParity proves the headline shape
// this change unblocks: with a leading "If X is 10 or more," resolving gate, the
// pump clause "creatures you control get +X/+X and gain haste until end of turn"
// recovers its group subject (EffectStaticSubjectControlledCreatures) and exact
// flag, and the trailing keyword grant links to that prior subject — identical to
// the standalone clause without the gate. Before this change the leading
// condition suppressed the subject (StaticSubject=None, Exact=false), dropping the
// buffed group. The gate is still recorded as an "If" condition boundary so
// lowering can gate the recovered effects.
func TestStripLeadingResolvingGateRecoversGroupPumpParity(t *testing.T) {
	t.Parallel()
	const gated = "If X is 10 or more, creatures you control get +X/+X and gain haste until end of turn."
	const standalone = "Creatures you control get +X/+X and gain haste until end of turn."

	gatedDoc, diagnostics := Parse(gated, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("gated diagnostics = %#v", diagnostics)
	}
	if len(gatedDoc.Abilities) != 1 {
		t.Fatalf("gated abilities = %d, want one", len(gatedDoc.Abilities))
	}
	boundaries := gatedDoc.Abilities[0].ConditionBoundaries
	if len(boundaries) != 1 || boundaries[0].Kind != ConditionIntroIf {
		t.Fatalf("gated condition boundaries = %#v, want one ConditionIntroIf", boundaries)
	}

	assertGroupPumpHaste := func(t *testing.T, label string, effects []EffectSyntax) {
		t.Helper()
		if len(effects) != 2 {
			t.Fatalf("%s effects = %d, want two (pump, keyword grant)", label, len(effects))
		}
		pump := effects[0]
		if pump.Kind != EffectModifyPT {
			t.Fatalf("%s effect 0 kind = %v, want EffectModifyPT", label, pump.Kind)
		}
		if !pump.Exact {
			t.Fatalf("%s pump Exact = false, want true", label)
		}
		if pump.StaticSubject.Kind != EffectStaticSubjectControlledCreatures {
			t.Fatalf("%s pump StaticSubject.Kind = %v, want EffectStaticSubjectControlledCreatures", label, pump.StaticSubject.Kind)
		}
		if !pump.PowerDelta.VariableX || !pump.ToughnessDelta.VariableX {
			t.Fatalf("%s pump delta = %+v/%+v, want +X/+X (VariableX)", label, pump.PowerDelta, pump.ToughnessDelta)
		}
		grant := effects[1]
		if grant.Kind != EffectGain {
			t.Fatalf("%s effect 1 kind = %v, want EffectGain", label, grant.Kind)
		}
		if !grant.Exact {
			t.Fatalf("%s keyword grant Exact = false, want true", label)
		}
		if grant.Context != EffectContextPriorSubject {
			t.Fatalf("%s keyword grant Context = %v, want EffectContextPriorSubject", label, grant.Context)
		}
	}

	if len(gatedDoc.Abilities[0].Sentences) != 1 {
		t.Fatalf("gated sentences = %d, want one", len(gatedDoc.Abilities[0].Sentences))
	}
	assertGroupPumpHaste(t, "gated", gatedDoc.Abilities[0].Sentences[0].Effects)

	standaloneDoc, diagnostics := Parse(standalone, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("standalone diagnostics = %#v", diagnostics)
	}
	if len(standaloneDoc.Abilities) != 1 || len(standaloneDoc.Abilities[0].Sentences) != 1 {
		t.Fatalf("standalone abilities = %#v", standaloneDoc.Abilities)
	}
	if len(standaloneDoc.Abilities[0].ConditionBoundaries) != 0 {
		t.Fatalf("standalone condition boundaries = %#v, want none", standaloneDoc.Abilities[0].ConditionBoundaries)
	}
	assertGroupPumpHaste(t, "standalone", standaloneDoc.Abilities[0].Sentences[0].Effects)
}

// TestStripLeadingResolvingGateAllowlist proves every allowlisted resolving-state
// gate, when it opens a sentence, is stripped so the gated group-pump clause is
// recognized as an exact controlled-creatures modification. Each wording is drawn
// from a real card whose gate exercises one of the leadingResolvingGateRecognizers
// entries, keeping the allowlist honest against printed text.
func TestStripLeadingResolvingGateAllowlist(t *testing.T) {
	t.Parallel()
	const body = ", creatures you control get +1/+1 until end of turn."
	cases := []struct {
		name string
		gate string
	}{
		{"spell X", "If X is 10 or more"},
		{"cast timing kicked", "If this spell was kicked"},
		{"cast timing bargained", "If this spell was bargained"},
		{"cast timing main phase", "If you cast this spell during your main phase"},
		{"cast timing from graveyard", "If this spell was cast from a graveyard"},
		{"controls", "If you control a creature with power 4 or greater"},
		{"graveyard count", "If there are seven or more cards in your graveyard"},
		{"adamant mana spent", "If at least three white mana was spent to cast this spell"},
		{"source counter state", "If this creature has a +1/+1 counter on it"},
		{"ability resolution ordinal", "If this is the second time this ability has resolved this turn"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			pump, ok := firstModifyPTEffect(t, tc.gate+body)
			if !ok {
				t.Fatalf("gate %q produced no EffectModifyPT", tc.gate)
			}
			if !pump.Exact {
				t.Fatalf("gate %q pump Exact = false, want true (gate not stripped)", tc.gate)
			}
			if pump.StaticSubject.Kind != EffectStaticSubjectControlledCreatures {
				t.Fatalf("gate %q pump StaticSubject.Kind = %v, want EffectStaticSubjectControlledCreatures", tc.gate, pump.StaticSubject.Kind)
			}
		})
	}
}

// TestLeadingResolvingGateNearMissDoesNotStrip proves the allowlist is curated,
// not a blanket "If" strip: a leading resolving gate outside the allowlist (a
// controller-life threshold) is left on the clause, so the pump subject stays
// unrecognized (Exact=false, StaticSubject=None) and the shape fails closed rather
// than silently buffing the wrong object. The same body without the gate recovers
// the group subject, confirming only the un-allowlisted gate suppresses it.
func TestLeadingResolvingGateNearMissDoesNotStrip(t *testing.T) {
	t.Parallel()
	const nearMiss = "If you have 40 or more life, creatures you control get +1/+1 until end of turn."
	const body = "Creatures you control get +1/+1 until end of turn."

	pump, ok := firstModifyPTEffect(t, nearMiss)
	if !ok {
		t.Fatal("near-miss produced no EffectModifyPT")
	}
	if pump.Exact {
		t.Fatal("near-miss pump Exact = true, want false (un-allowlisted gate must not strip)")
	}
	if pump.StaticSubject.Kind == EffectStaticSubjectControlledCreatures {
		t.Fatal("near-miss recovered the group subject, want it suppressed by the retained gate")
	}

	bare, ok := firstModifyPTEffect(t, body)
	if !ok {
		t.Fatal("bare body produced no EffectModifyPT")
	}
	if !bare.Exact || bare.StaticSubject.Kind != EffectStaticSubjectControlledCreatures {
		t.Fatalf("bare body pump = {Exact:%v Subject:%v}, want exact controlled creatures", bare.Exact, bare.StaticSubject.Kind)
	}
}
