package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestLibraryCardPredicateAndEventSubjectKeepDistinctBindings(t *testing.T) {
	compilation, diagnostics := compileSource(
		"Whenever a creature attacks you or a planeswalker you control, reveal the top card of your library. If it's a Forest card, remove that creature from combat. Then put the revealed card on the bottom of your library.",
		pipelineContext{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	effect := content.Effects[1]
	if len(effect.References) != 1 || effect.References[0].Binding != ReferenceBindingEventPermanent {
		t.Fatalf("event consumer context=%v exact=%v refs=%#v", effect.Context, effect.Exact, effect.References)
	}
	references := append([]CompiledReference(nil), content.References...)
	effects := append([]CompiledEffect(nil), content.Effects...)
	for i := range references {
		references[i].Binding = ReferenceBindingUnsupported
		references[i].Text = "opaque reference"
		references[i].Span = shared.Span{}
	}
	for i := range effects {
		effects[i].Text = "opaque effect"
		effects[i].Span, effects[i].ClauseSpan, effects[i].VerbSpan = shared.Span{}, shared.Span{}, shared.Span{}
	}
	rebound := bindReferences(references, content.Targets, effects, compilation.Abilities[0].Trigger)
	for i := range rebound {
		if rebound[i].Binding != content.References[i].Binding ||
			rebound[i].NodeID != content.References[i].NodeID ||
			rebound[i].ProducerClauseID != content.References[i].ProducerClauseID {
			t.Fatalf("opaque/zero-span compilation changed subject identity: before=%#v after=%#v",
				content.References[i], rebound[i])
		}
	}
}
