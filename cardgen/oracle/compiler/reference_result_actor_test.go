package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestResultActorProjectsExactEventController(t *testing.T) {
	document, diagnostics := parser.Parse(
		"Whenever a creature deals combat damage to one of your opponents, its controller may pay 1 life. If they do, they draw a card.",
		parser.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compiled, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compiled.Abilities[0].Content
	if content.Conditions[0].Ownership.ResultProducerClauseID != 1 {
		t.Fatal("condition lost the actual payment action")
	}
	for _, reference := range content.References {
		if reference.Pronoun == ReferencePronounIts {
			if !content.OwnsSubject(reference) || reference.SubjectDomain() != ReferenceSubjectPermanent {
				t.Fatal("controller projection erased its event object")
			}
			continue
		}
		if !content.OwnsSubject(reference) || reference.Binding != ReferenceBindingEventPermanent ||
			reference.SubjectDomain() != ReferenceSubjectPlayer || reference.ObjectProjectionDomain() != ReferenceSubjectPermanent ||
			!reference.EventControllerResultActor() {
			t.Fatal("payer/result consumer borrowed the damaged event player")
		}
	}
	for _, tc := range []struct {
		name   string
		mutate func(*AbilityContent)
	}{
		{"missing producer", func(c *AbilityContent) { c.Conditions[0].Ownership.ResultProducerClauseID = 0 }},
		{"forward producer", func(c *AbilityContent) { c.Conditions[0].Ownership.ResultProducerClauseID = 2 }},
		{"wrong action actor", func(c *AbilityContent) { c.Effects[0].Context = parser.EffectContextController }},
		{"missing subject", func(c *AbilityContent) { c.Effects[0].SubjectReferences = nil }},
		{"foreign condition node", func(c *AbilityContent) { c.Conditions[0].Ownership.ResultSubjectReferenceNodeID = 99 }},
		{"wrong source binding", func(c *AbilityContent) { c.References[0].Binding = ReferenceBindingEventPlayer }},
		{"duplicate event subject", func(c *AbilityContent) { c.References = append(c.References, c.References[0]) }},
		{"duplicate producer", func(c *AbilityContent) { c.Effects[1].ClauseID = 1 }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compiled, _ := Compile(document, Context{})
			content := compiled.Abilities[0].Content
			tc.mutate(&content)
			if _, ok := resultReferencedActorProducer(content, content.References[1]); ok {
				t.Fatal("invalid owned event-controller action acquired a projection")
			}
		})
	}
}

func TestResultActorProjectionDoesNotBorrowZoneEventPlayer(t *testing.T) {
	document, diagnostics := parser.Parse(
		"Whenever a card is put into an opponent's graveyard from anywhere, you may have that player lose 2 life. If you do, you gain 2 life.",
		parser.Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compiled, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compiled.Abilities[0].Content
	if len(content.References) != 1 || !content.OwnsSubject(content.References[0]) {
		t.Fatal("missing canonical zone-event player reference")
	}
	if content.References[0].EventControllerResultActor() {
		t.Fatal("generic player projection invented an event-permanent controller result actor")
	}
}
