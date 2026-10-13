package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestPermanentUnionTargetSubjectProof(t *testing.T) {
	const text = "Counter target spell. Return target creature or Vehicle to its owner's hand. Its controller loses 2 life."
	compile := func() AbilityContent {
		doc, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := Compile(doc, Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		return compilation.Abilities[0].Content
	}
	content, foreign := compile(), compile()
	subject, ok := content.TargetSubject(1)
	if !ok || !content.OwnsSubject(subject) ||
		subject.SubjectDomain() != ReferenceSubjectPermanent ||
		subject.SubjectLifetime() != ReferenceLifetimeOriginalObject {
		t.Fatalf("missing exact original-permanent union proof: %#v", subject)
	}
	target := content.Targets[1]
	localized, ok := subject.LocalizeSubjectOccurrence(target, 0)
	if !ok || localized.Occurrence != 0 || localized.SubjectDomain() != ReferenceSubjectPermanent {
		t.Fatal("owned union did not localize mechanically")
	}
	for _, tc := range []struct {
		name   string
		mutate func(*CompiledTarget)
	}{
		{"foreign", func(t *CompiledTarget) { *t = foreign.Targets[1] }},
		{"empty order", func(t *CompiledTarget) { t.Order = shared.SourceOrder{} }},
		{"order", func(t *CompiledTarget) { t.Order.Start++ }},
		{"card zone", func(t *CompiledTarget) { t.Selector.Zone = zone.Graveyard }},
		{"stack domain", func(t *CompiledTarget) { t.Selector.Kind = SelectorSpell }},
		{"cardinality", func(t *CompiledTarget) { t.Cardinality.Max++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := target
			tc.mutate(&changed)
			if _, ok := subject.LocalizeSubjectOccurrence(changed, 0); ok {
				t.Fatal("foreign or modified target occurrence localized")
			}
		})
	}
	if _, ok := (AbilityContent{Targets: content.Targets}).TargetSubject(1); ok {
		t.Fatal("unowned target acquired a subject proof")
	}
	if foreign.OwnsSubject(subject) {
		t.Fatal("foreign body accepted identical target facts")
	}
}
