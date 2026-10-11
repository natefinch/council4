package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestPlayerActorOrderProjectionPreservesBoundIdentity(t *testing.T) {
	t.Parallel()
	canonical := CompiledReference{NodeID: 0, Kind: ReferencePronoun, Pronoun: ReferencePronounThey,
		Binding: ReferenceBindingTarget, Occurrence: 7, Order: shared.SourceOrder{Start: 9, End: 10}}
	for _, tt := range []struct {
		name    string
		context parser.EffectContextKind
		binding ReferenceBinding
		order   shared.SourceOrder
		want    shared.SourceOrder
	}{
		{"valid zero ID source", parser.EffectContextEventPlayer, ReferenceBindingSource, shared.SourceOrder{}, canonical.Order},
		{"cost snapshot stays bound", parser.EffectContextReferencedPlayer, ReferenceBindingPaidCost, shared.SourceOrder{}, canonical.Order},
		{"event stays bound", parser.EffectContextEventPlayer, ReferenceBindingEventPlayer, shared.SourceOrder{}, canonical.Order},
		{"unrelated controller", parser.EffectContextController, ReferenceBindingSource, shared.SourceOrder{}, shared.SourceOrder{}},
		{"drifted nonempty order", parser.EffectContextEventPlayer, ReferenceBindingSource,
			shared.SourceOrder{Start: 11, End: 12}, shared.SourceOrder{Start: 11, End: 12}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			owned := canonical
			owned.Order, owned.Binding, owned.Occurrence = tt.order, tt.binding, 3
			effects := []CompiledEffect{{ClauseID: 1, Context: tt.context, References: []CompiledReference{owned}}}
			projectPlayerActorReferenceOrders(effects, []CompiledReference{canonical})
			projected := effects[0].References[0]
			if projected.Order != tt.want || projected.Binding != owned.Binding || projected.Occurrence != owned.Occurrence ||
				projected.NodeID != owned.NodeID || projected.Kind != owned.Kind || projected.Pronoun != owned.Pronoun {
				t.Fatalf("projection changed bound identity: before=%+v after=%+v", owned, projected)
			}
		})
	}
}
