package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileCounterResultOwnership(t *testing.T) {
	t.Parallel()
	for _, text := range []string{
		"Counter target spell unless its controller pays {3}. If that spell is countered this way, draw a card.",
		"Exile target creature. Counter target spell. If that spell is countered this way, draw a card.",
		"{1}: Counter target spell. If that spell is countered this way, you gain 2 life.",
		"When this artifact enters, counter target spell. If that spell is countered this way, draw a card.",
		"Choose one \u2014\n\u2022 Counter target spell. If that spell is countered this way, draw a card.\n\u2022 You gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			document, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}

			// Wording and spans are diagnostic-only inputs to compilation.
			ability := &document.Abilities[0]
			for i := range ability.ConditionSegments {
				ability.ConditionSegments[i].Text = "unrelated diagnostic"
			}
			compiled, diagnostics := Compile(document, Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compiled.Abilities[0].Content
			if len(content.Modes) > 0 {
				content = content.Modes[0].Content
			}
			found := false
			for _, condition := range content.Conditions {
				if condition.Predicate != ConditionPredicateCounterSucceeded {
					continue
				}
				found = true
				if condition.Ownership.ResultProducerClauseID <= 0 ||
					condition.Ownership.ResultSubjectReferenceNodeID < 0 {
					t.Fatalf("ownership = %#v", condition.Ownership)
				}
				for _, reference := range content.References {
					if reference.NodeID == condition.Ownership.ResultSubjectReferenceNodeID &&
						(reference.Binding != ReferenceBindingTarget ||
							reference.Occurrence != condition.Ownership.ResultSubjectTargetOccurrence) {
						t.Fatalf("counter subject lost owned target occurrence: %#v", reference)
					}
				}
			}
			if !found {
				t.Fatal("counter result predicate missing")
			}
		})
	}
}

func TestCounterResultSubjectBindingRefusesUnprovenOwner(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse("Exile target creature. Counter target spell. If that spell is countered this way, draw a card.",
		parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	for _, tt := range []struct {
		name   string
		mutate func(*CompiledCondition)
	}{
		{"missing producer", func(c *CompiledCondition) { c.Ownership.ResultProducerClauseID = 999 }},
		{"different producer", func(c *CompiledCondition) { c.Ownership.ResultProducerClauseID = content.Effects[0].ClauseID }},
		{"wrong target occurrence", func(c *CompiledCondition) { c.Ownership.ResultSubjectTargetOccurrence = 0 }},
		{"missing target occurrence", func(c *CompiledCondition) { c.Ownership.ResultSubjectTargetOccurrence = -1 }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			conditions := slices.Clone(content.Conditions)
			references := slices.Clone(content.References)
			tt.mutate(&conditions[0])
			bindCounterResultReferences(conditions, references, content.Targets, content.Effects)
			for _, reference := range references {
				if reference.NodeID == conditions[0].Ownership.ResultSubjectReferenceNodeID &&
					reference.Binding != ReferenceBindingUnsupported {
					t.Fatalf("unproven owner retained a plausible binding: %#v", reference)
				}
			}
		})
	}
}
