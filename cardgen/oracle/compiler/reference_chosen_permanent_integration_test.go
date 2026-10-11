package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func compiledChosenPaymentSubject(t *testing.T) (AbilityContent, CompiledReference, CompiledEffect) {
	t.Helper()
	compilation, diagnostics := compileSource(
		"When this creature deals combat damage to a player, you may pay {3} and sacrifice it. If you do, choose a land you control. Then "+chosenLandSearch,
		pipelineContext{CardName: "Canoptek Wraith"})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	return content, content.Effects[2].References[0], content.Effects[1]
}

func TestChosenNameFoldedPaymentProjectionKeepsCanonicalAuthority(t *testing.T) {
	content, reference, choice := compiledChosenPaymentSubject(t)
	projected := choice
	projected.Payment = CompiledEffectPayment{}
	if !content.OwnsSubject(reference) || !reference.SubjectProducerMatches(choice) ||
		!reference.SubjectProducerMatches(projected) {
		t.Fatal("exact folded payment externalization lost original chosen subject")
	}
	for _, tc := range []struct {
		name   string
		mutate func(*CompiledEffectPayment)
	}{
		{"order", func(p *CompiledEffectPayment) { p.Order.Start++ }},
		{"form", func(p *CompiledEffectPayment) { p.Form = parser.EffectPaymentFormUnknown }},
		{"payer", func(p *CompiledEffectPayment) { p.Payer = parser.EffectPaymentPayerTargetPlayer }},
		{"node", func(p *CompiledEffectPayment) { p.SuccessConditionNodeID++ }},
		{"mana", func(p *CompiledEffectPayment) { p.ManaCost = nil }},
		{"cost", func(p *CompiledEffectPayment) { p.AdditionalCost = nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			changed := choice
			tc.mutate(&changed.Payment)
			if reference.SubjectProducerMatches(changed) {
				t.Fatal("nonempty changed payment acquired exact projected authority")
			}
		})
	}
}

func TestChosenNameFoldedProjectionDoesNotRepairChangedCanonicalScope(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*AbilityContent)
	}{
		{"missing condition", func(c *AbilityContent) { c.Conditions = nil }},
		{"duplicate condition", func(c *AbilityContent) { c.Conditions = append(c.Conditions, c.Conditions[0]) }},
		{"independent scope", func(c *AbilityContent) { c.Conditions[0].Ownership.ClauseIDs = []int{3} }},
		{"unsupported scope", func(c *AbilityContent) { c.Conditions[0].Ownership.Scope = parser.ConditionScopeUnsupported }},
		{"condition order", func(c *AbilityContent) { c.Conditions[0].Order.Start++ }},
		{"condition node", func(c *AbilityContent) { c.Conditions[0].NodeID++ }},
		{"condition kind", func(c *AbilityContent) { c.Conditions[0].Kind = ConditionUnless }},
		{"condition predicate", func(c *AbilityContent) { c.Conditions[0].Predicate++ }},
		{"canonical payment removed", func(c *AbilityContent) { c.Effects[1].Payment = CompiledEffectPayment{} }},
		{"canonical payer", func(c *AbilityContent) { c.Effects[1].Payment.Payer = parser.EffectPaymentPayerTargetPlayer }},
		{"canonical node", func(c *AbilityContent) { c.Effects[1].Payment.SuccessConditionNodeID++ }},
		{"canonical order", func(c *AbilityContent) { c.Effects[1].Payment.Order.Start++ }},
		{"canonical mana", func(c *AbilityContent) { c.Effects[1].Payment.ManaCost = nil }},
		{"canonical cost source", func(c *AbilityContent) {
			c.Effects[1].Payment.AdditionalCost.Components[0].SourceSelf = false
		}},
		{"canonical cost amount", func(c *AbilityContent) {
			c.Effects[1].Payment.AdditionalCost.Components[0].AmountValue++
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, reference, choice := compiledChosenPaymentSubject(t)
			projected := choice
			projected.Payment = CompiledEffectPayment{}
			tc.mutate(reference.Subject.scope.owner)
			if reference.SubjectSupported() || reference.SubjectProducerMatches(projected) {
				t.Fatal("empty projected envelope repaired altered canonical payment or gate")
			}
		})
	}
}

func TestChosenNameFreshFrameIgnoresOnlyEarlierImplicitSubjects(t *testing.T) {
	for _, prefix := range []string{
		"Manifest the top card of your library.",
		"Bolster 1.",
		"Choose a creature type.",
	} {
		t.Run(prefix, func(t *testing.T) {
			content := compileChosenNameBody(t, prefix+" Choose a land you control. "+chosenLandSearch)
			reference := content.Effects[2].References[0]
			if !content.Effects[2].SearchNameFromChosenPermanent || !content.OwnsSubject(reference) ||
				reference.ChoiceClauseID != content.Effects[1].ClauseID ||
				!reference.SubjectProducerMatches(content.Effects[1]) {
				t.Fatal("earlier implicit product poisoned a new original choice frame")
			}
			content.Effects[0].VerbOrder = content.Effects[1].VerbOrder
			content.Effects[0].VerbOrder.Start = content.Effects[1].VerbOrder.End
			content.Effects[0].VerbOrder.End = content.Effects[0].VerbOrder.Start + 1
			if reference.SubjectSupported() || reference.SubjectProducerMatches(content.Effects[1]) {
				t.Fatal("post-choice implicit subject retained the earlier-frame authority")
			}
		})
	}
}

func TestChosenNameFreshFrameDoesNotAcceptCompetingExplicitDeclarations(t *testing.T) {
	for _, text := range []string{
		"Bolster 1. Choose a land you control. Choose a land you control. " + chosenLandSearch,
		"Bolster 1. Choose a land you control. " + chosenLandSearch + " Choose a land you control.",
		"Bolster 1. Choose a creature you control. " + chosenLandSearch,
		"Bolster 1. If you control a creature, choose a land you control. " + chosenLandSearch,
		"Choose a land you control. Bolster 1. " + chosenLandSearch,
		"Choose a land you control. Manifest the top card of your library. " + chosenLandSearch,
		"Choose a land you control. Choose a creature type. " + chosenLandSearch,
	} {
		t.Run(text, func(t *testing.T) {
			content := compileChosenNameBody(t, text)
			for _, effect := range content.Effects {
				if effect.Kind == EffectSearch && effect.SearchNameFromChosenPermanent {
					t.Fatal("fresh frame replaced ambiguous/skipped/wrong-noun authority")
				}
			}
		})
	}
}
