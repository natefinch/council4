package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/zone"
)

func compiledEnteredSubjectBody(t *testing.T) AbilityContent {
	t.Helper()
	document, diagnostics := parser.Parse(
		"Return target creature card from your graveyard to the battlefield. That creature gains haste until end of turn.",
		parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
		t.Fatal(diagnostics)
	}
	return compilation.Abilities[0].Content
}

func TestReferenceSubjectScopeAndImmutableFacts(t *testing.T) {
	content := compiledEnteredSubjectBody(t)
	foreign := compiledEnteredSubjectBody(t)
	reference := content.References[0]
	if !content.OwnsSubject(reference) || !reference.SubjectProducerMatches(content.Effects[0]) ||
		!reference.ProducerTargetMatches(content.Targets[0]) || !reference.EnteredSubjectSupported() {
		t.Fatal("normally compiled subject lost its original scope/occurrence/entered proof")
	}
	if foreign.OwnsSubject(reference) || reference.SubjectProducerMatches(foreign.Effects[0]) ||
		reference.ProducerTargetMatches(foreign.Targets[0]) {
		t.Fatal("identical NodeIDs/ClauseIDs/orders from another body were accepted")
	}
	for _, tc := range []struct {
		name   string
		mutate func(*CompiledReference)
	}{
		{"missing proof", func(r *CompiledReference) { r.Subject = ReferenceSubjectProof{} }},
		{"foreign binding", func(r *CompiledReference) { r.Binding = ReferenceBindingSource }},
		{"kind", func(r *CompiledReference) { r.Kind = ReferenceThisObject }},
		{"pronoun", func(r *CompiledReference) { r.Pronoun = ReferencePronounIt }},
		{"node", func(r *CompiledReference) { r.NodeID++ }},
		{"producer", func(r *CompiledReference) { r.ProducerClauseID++ }},
		{"card domain", func(r *CompiledReference) { r.CardIdentity = !r.CardIdentity }},
		{"noun", func(r *CompiledReference) { r.SubjectNoun = parser.ObjectNounCard }},
		{"observation", func(r *CompiledReference) { r.LibraryCardObservation = true }},
		{"occurrence", func(r *CompiledReference) { r.Occurrence++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			copy := reference
			tc.mutate(&copy)
			if copy.SubjectSupported() || copy.SubjectProducerMatches(content.Effects[0]) {
				t.Fatal("modified semantic fact retained a validated proof")
			}
		})
	}
	copy := reference
	copy.Text = "opaque"
	copy.PriorInstruction = 7
	if !copy.SubjectSupported() || !copy.SubjectProducerMatches(content.Effects[0]) {
		t.Fatal("mechanical instruction reindexing or display metadata changed referent identity")
	}
	target := content.Targets[0]
	target.Text = "opaque"
	if !reference.ProducerTargetMatches(target) {
		t.Fatal("display text changed exact target occurrence")
	}
	target.Selector.Zone = zone.Exile
	if reference.ProducerTargetMatches(target) {
		t.Fatal("another card lifetime retained the original target proof")
	}
	producer := content.Effects[0]
	producer.ToZone = zone.Hand
	if reference.SubjectProducerMatches(producer) {
		t.Fatal("a non-entry operation retained actual-entry proof")
	}
}

func TestReferenceSubjectConditionCopiesShareFinalProof(t *testing.T) {
	for _, text := range []string{
		"Whenever a land you control enters, tap target creature. If that land is an Island, draw a card.",
		"When this creature enters, exile target card from a graveyard. You gain 1 life. If it was a creature card, draw a card.",
		"Whenever you cast an instant or sorcery spell, tap target creature. If that spell has mana value 5 or greater, draw a card.",
		"When this creature enters, exile target creature. Return it to the battlefield under its owner's control. If that creature was a Human, draw a card.",
	} {
		t.Run(text, func(t *testing.T) {
			compilation, diagnostics := compileSource(text, pipelineContext{CardName: "Final Proof"})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			for _, condition := range content.Conditions {
				if condition.ObjectReference == nil {
					t.Fatal("normal condition lost its subject")
				}
				reference := *condition.ObjectReference
				if !content.OwnsSubject(reference) || reference.Binding != condition.ObjectBinding {
					t.Fatal("condition does not carry its body's final validated subject")
				}
				copies := 0
				for _, canonical := range content.References {
					if canonical.NodeID == reference.NodeID {
						copies++
						if canonical.Subject != reference.Subject || canonical.Binding != reference.Binding ||
							canonical.PriorInstruction != reference.PriorInstruction ||
							canonical.Occurrence != reference.Occurrence {
							t.Fatal("condition and canonical copies disagree after a binding rewrite")
						}
					}
				}
				if copies != 1 {
					t.Fatal("condition has no unique canonical reference occurrence")
				}
			}
		})
	}
}

func TestReferenceSubjectCardLineageOwnership(t *testing.T) {
	for _, text := range []string{
		"Reveal the top card of your library. If it's a land card, put it onto the battlefield. If it's a snow card, you gain 2 life.",
		"Reveal the top card of your library. If it's a land card, you may put it onto the battlefield. If it's a snow card, you gain 2 life.",
	} {
		t.Run(text, func(t *testing.T) {
			document, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			compilation, diagnostics := Compile(document, Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			reference := *content.Conditions[len(content.Conditions)-1].ObjectReference
			reached, ok := reference.CardLineageSubject(content.Effects)
			if !ok || reference.ProducerClauseID != 1 || !reference.LibraryCardObservation ||
				reached.ProducerClauseID != 2 || reached.LibraryCardObservation ||
				reached.SubjectDomain() != ReferenceSubjectCard || !reached.EnteredSubjectSupported() {
				t.Fatal("independent card predicate lost its exact input/reached lineage")
			}
			for _, foreign := range [][]CompiledEffect{
				nil,
				{content.Effects[0]},
				{content.Effects[1], content.Effects[1]},
				compiledEnteredSubjectBody(t).Effects,
			} {
				if _, ok := reference.CardLineageSubject(foreign); ok {
					t.Fatal("unavailable/ambiguous/foreign lineage producer was accepted")
				}
			}
			reordered := []CompiledEffect{content.Effects[2], content.Effects[0], content.Effects[1]}
			reindexed, ok := reference.CardLineageSubject(reordered)
			if !ok || reindexed.PriorInstruction != 2 || reindexed.ProducerClauseID != 2 {
				t.Fatal("mechanical position remapping changed proven lineage")
			}
		})
	}
}
