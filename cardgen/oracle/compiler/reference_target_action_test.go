package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func compiledTargetActionBody(t *testing.T, text string) AbilityContent {
	t.Helper()
	document, diagnostics := parser.Parse(text, parser.Context{CardName: "Target Action", InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	return compilation.Abilities[0].Content
}

func TestTargetActionSubject(t *testing.T) {
	for _, prefix := range []string{"", "Target player gains 1 life. ", "Exile target card from your graveyard. ", "Counter target spell. "} {
		content := compiledTargetActionBody(t, prefix+"Gain control of target spell. You gain 2 life. You may choose new targets for it.")
		effect := content.Effects[len(content.Effects)-1]
		local := localizedTargetAction(t, content, effect)
		subject, ok := local.TargetActionSubject(effect)
		if !ok || subject.SubjectDomain() != ReferenceSubjectStackObject ||
			subject.SubjectLifetime() != ReferenceLifetimeStackOccurrence || subject.Occurrence != 0 {
			t.Fatalf("prefix %q lost exact stack subject: %+v", prefix, subject)
		}
	}
}

func localizedTargetAction(t *testing.T, content AbilityContent, effect CompiledEffect) AbilityContent {
	t.Helper()
	local := content
	local.Effects = []CompiledEffect{effect}
	local.References = effect.References
	local.Targets = effect.Targets
	if len(local.References) == 1 {
		target := content.Targets[local.References[0].Occurrence]
		local.Targets = []CompiledTarget{target}
		reference, ok := local.References[0].LocalizeSubjectOccurrence(target, 0)
		if !ok {
			t.Fatal("cannot localize target subject")
		}
		local.References = []CompiledReference{reference}
	}
	return local
}

func TestTargetActionSubjectRefusesUnownedOrAmbiguousIdentity(t *testing.T) {
	const text = "Gain control of target spell. You gain 2 life. You may choose new targets for it."
	for _, tt := range []struct {
		name   string
		mutate func(*AbilityContent, *CompiledEffect)
	}{
		{"missing proof", func(c *AbilityContent, _ *CompiledEffect) { c.References[0].Subject = ReferenceSubjectProof{} }},
		{"foreign reference", func(c *AbilityContent, _ *CompiledEffect) {
			foreign := compiledTargetActionBody(t, text)
			c.References[0] = foreign.References[0]
		}},
		{"foreign target", func(c *AbilityContent, _ *CompiledEffect) {
			c.Targets[0] = compiledTargetActionBody(t, text).Targets[0]
		}},
		{"unowned effect", func(_ *AbilityContent, e *CompiledEffect) { e.subjectScope = nil }},
		{"wrong clause", func(_ *AbilityContent, e *CompiledEffect) { e.ClauseID = 1 }},
		{"wrong effect order", func(_ *AbilityContent, e *CompiledEffect) { e.Order.Start++ }},
		{"wrong reference order", func(c *AbilityContent, _ *CompiledEffect) { c.References[0].Order.Start++ }},
		{"forward target", func(c *AbilityContent, _ *CompiledEffect) { c.Targets[0].Order.End = c.References[0].Order.End + 1 }},
		{"duplicate reference", func(c *AbilityContent, _ *CompiledEffect) { c.References = append(c.References, c.References[0]) }},
		{"duplicate canonical reference", func(c *AbilityContent, _ *CompiledEffect) {
			owner := c.subjectScope.owner
			owner.References = append(owner.References, owner.References[0])
		}},
		{"duplicate canonical clause", func(c *AbilityContent, e *CompiledEffect) {
			owner := c.subjectScope.owner
			owner.Effects = append(owner.Effects, *e)
		}},
		{"missing reference", func(c *AbilityContent, _ *CompiledEffect) { c.References = nil }},
		{"default reference", func(c *AbilityContent, _ *CompiledEffect) {
			c.References[0] = CompiledReference{Binding: ReferenceBindingTarget}
		}},
		{"source instead of target", func(c *AbilityContent, _ *CompiledEffect) { c.References[0].Binding = ReferenceBindingSource }},
		{"event instead of target", func(c *AbilityContent, _ *CompiledEffect) { c.References[0].Binding = ReferenceBindingEventStackObject }},
		{"foreign noun", func(c *AbilityContent, _ *CompiledEffect) { c.References[0].SubjectNoun = parser.ObjectNounCreature }},
		{"plural target", func(c *AbilityContent, _ *CompiledEffect) { c.Targets[0].Cardinality.Max = 2 }},
		{"optional target occurrence", func(c *AbilityContent, _ *CompiledEffect) { c.Targets[0].Cardinality.Min = 0 }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			content := compiledTargetActionBody(t, text)
			effect := content.Effects[2]
			local := localizedTargetAction(t, content, effect)
			local.Targets = append([]CompiledTarget(nil), local.Targets...)
			tt.mutate(&local, &effect)
			if _, ok := local.TargetActionSubject(effect); ok {
				t.Fatal("unowned or ambiguous target subject accepted")
			}
		})
	}
}

func TestTargetActionSubjectDeclaredTarget(t *testing.T) {
	content := compiledTargetActionBody(t, "Gain control of target spell.")
	for _, effect := range content.Effects {
		local := content
		local.Effects = []CompiledEffect{effect}
		local.References = effect.References
		local.Targets = effect.Targets
		subject, ok := local.TargetActionSubject(effect)
		if !ok || subject.SubjectDomain() != ReferenceSubjectStackObject ||
			subject.SubjectLifetime() != ReferenceLifetimeStackOccurrence || subject.Occurrence != 0 {
			t.Fatalf("declared stack subject failed: %+v", subject)
		}
	}
}
