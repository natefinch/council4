package compiler

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

const observedPostfixMatterText = "({C} represents colorless mana.)\nWhen this creature dies, reveal the top card of your library. You may put that card onto the battlefield if it's a permanent card with mana value 3 or less. Otherwise, put that card into your hand."

func observedPostfixCompilation(t *testing.T) (AbilityContent, *CompiledTrigger) {
	t.Helper()
	compilation, diagnostics := compileSource(observedPostfixMatterText, pipelineContext{CardName: "Matter Reshaper"})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	for _, ability := range compilation.Abilities {
		if ability.Kind == AbilityTriggered {
			return ability.Content, ability.Trigger
		}
	}
	t.Fatal("complete original input has no death trigger")
	return AbilityContent{}, nil
}

func TestObservedPostfixOptionalTypedBinding(t *testing.T) {
	content, _ := observedPostfixCompilation(t)
	if len(content.Effects) != 3 || len(content.Conditions) != 1 {
		t.Fatalf("effects=%#v conditions=%#v", content.Effects, content.Conditions)
	}
	for _, effect := range content.Effects {
		t.Logf("clause=%d optional=%v group=%v postfix=%+v connection=%v resultElse=%d",
			effect.ClauseID, effect.Optional, effect.OptionalActionClauseIDs,
			effect.OptionalPostfixElse, effect.Connection, effect.ResultElseOfClauseID)
	}
	condition := content.Conditions[0]
	t.Logf("condition=%+v", condition)
	if condition.ObjectReference == nil ||
		condition.ObjectReference.Binding != ReferenceBindingPriorInstructionResult ||
		condition.ObjectReference.ProducerClauseID != content.Effects[0].ClauseID ||
		condition.ObjectReference.NodeID != condition.SubjectRefID {
		t.Fatal("condition does not read the exact earlier observation")
	}
	ownership := content.Effects[2].OptionalPostfixElse
	if ownership.ActionClauseID != content.Effects[1].ClauseID ||
		ownership.ConditionNodeID != condition.NodeID {
		t.Fatal("Otherwise lost the exact postfix action/condition")
	}
}

func TestObservedPostfixSubjectProofNearMisses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*AbilityContent)
		refuse bool
	}{
		{"opaque zero-span metadata", func(c *AbilityContent) {
			for i := range c.References {
				c.References[i].Text, c.References[i].Span, c.References[i].Order = "opaque", shared.Span{}, shared.SourceOrder{}
			}
			for i := range c.Effects {
				e := &c.Effects[i]
				e.Text, e.Span, e.ClauseSpan, e.VerbSpan = "opaque", shared.Span{}, shared.Span{}, shared.Span{}
				e.Order, e.VerbOrder = shared.SourceOrder{}, shared.SourceOrder{}
			}
		}, false},
		{"missing subject", func(c *AbilityContent) {
			c.References = slices.DeleteFunc(c.References, func(r CompiledReference) bool {
				return r.NodeID == c.Conditions[0].SubjectRefID
			})
		}, true},
		{"duplicate subject", func(c *AbilityContent) {
			for _, r := range c.References {
				if r.NodeID == c.Conditions[0].SubjectRefID {
					c.References = append(c.References, r)
					break
				}
			}
		}, true},
		{"wrong subject identity", func(c *AbilityContent) { c.Conditions[0].SubjectRefID = 10000 }, true},
		{"missing producer identity", func(c *AbilityContent) {
			for i := range c.References {
				if c.References[i].NodeID == c.Conditions[0].SubjectRefID {
					c.References[i].ProducerClauseID = 0
				}
			}
		}, true},
		{"plural producer", func(c *AbilityContent) { c.Effects[0].Amount.Value = 2 }, true},
		{"duplicate producer identity", func(c *AbilityContent) { c.Effects = append(c.Effects, c.Effects[0]) }, true},
		{"missing producer", func(c *AbilityContent) { c.Effects = c.Effects[1:] }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			content, trigger := observedPostfixCompilation(t)
			tc.mutate(&content)
			for i := range content.References {
				content.References[i].Binding = ReferenceBindingUnsupported
			}
			condition := content.Conditions[0]
			condition.ObjectReference = nil
			if ok := bindContextualObjectCondition(&condition, content.References, content.Targets, content.Effects, trigger); ok == tc.refuse {
				t.Fatalf("bound=%v want refusal=%v: subject=%#v", ok, tc.refuse, condition.ObjectReference)
			}
		})
	}
}

func TestObservedLibraryOwnerShuffleTypedProvenance(t *testing.T) {
	compilation, diagnostics := compileSource(
		"The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.",
		pipelineContext{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	for _, e := range content.Effects {
		t.Logf("clause=%d kind=%v context=%v player=%v source=%v exact=%v amount=%+v selector=%v targets=%+v refs=%+v",
			e.ClauseID, e.Kind, e.Context, e.Player, e.CardSource, e.Exact, e.Amount, e.Selector.Kind, e.Targets, e.References)
	}
	for _, r := range content.References {
		t.Logf("reference=%+v", r)
	}
	if len(content.Conditions) != 1 || len(content.Effects) != 3 {
		t.Fatalf("effects/conditions=%d/%d", len(content.Effects), len(content.Conditions))
	}
	c := content.Conditions[0]
	t.Logf("condition=%+v", c)
	if c.Predicate != ConditionPredicateObjectMatches || c.ObjectReference == nil ||
		c.ObjectReference.NodeID != c.SubjectRefID ||
		c.ObjectReference.ProducerClauseID != content.Effects[1].ClauseID ||
		c.ObjectReference.Binding != ReferenceBindingPriorInstructionResult {
		t.Fatal("owner-library predicate lost its exact reveal producer")
	}
}

func TestObservedLibraryOwnerProofNearMisses(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*AbilityContent)
		refuse bool
	}{
		{"opaque zero metadata", func(*AbilityContent) {}, false},
		{"missing owner clause", func(c *AbilityContent) { c.Effects[1].LibraryOwnerClauseID = 0 }, true},
		{"wrong owner clause", func(c *AbilityContent) { c.Effects[1].LibraryOwnerClauseID = c.Effects[2].ClauseID }, true},
		{"duplicate owner identity", func(c *AbilityContent) { c.Effects = append(c.Effects, c.Effects[0]) }, true},
		{"player not permanent target", func(c *AbilityContent) { c.Effects[0].Targets[0].Selector.Kind = SelectorPlayer }, true},
		{"plural target", func(c *AbilityContent) { c.Effects[0].Targets[0].Cardinality.Max = 2 }, true},
		{"unproven quantity", func(c *AbilityContent) { c.Effects[1].Amount.Known = false }, true},
		{"plural observation", func(c *AbilityContent) { c.Effects[1].Amount.Value = 2 }, true},
		{"optional observation", func(c *AbilityContent) { c.Effects[1].Optional = true }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compilation, diagnostics := compileSource(
				"The owner of target permanent shuffles it into their library, then reveals the top card of their library. If it's a permanent card, they put it onto the battlefield.",
				pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			for i := range content.Effects {
				e := &content.Effects[i]
				e.Text, e.Span, e.ClauseSpan, e.VerbSpan = "opaque", shared.Span{}, shared.Span{}, shared.Span{}
				e.Order, e.VerbOrder = shared.SourceOrder{}, shared.SourceOrder{}
			}
			tc.mutate(&content)
			c := content.Conditions[0]
			subject := *c.ObjectReference
			subject.Binding, subject.Text, subject.Span, subject.Order = ReferenceBindingUnsupported, "opaque", shared.Span{}, shared.SourceOrder{}
			handled := bindLibraryCardReference(&subject, content.Effects)
			if !handled || (subject.Binding == ReferenceBindingPriorInstructionResult) == tc.refuse {
				t.Fatalf("handled=%v binding=%v want refusal=%v", handled, subject.Binding, tc.refuse)
			}
		})
	}
}
