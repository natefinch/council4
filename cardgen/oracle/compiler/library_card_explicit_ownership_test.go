package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestCompilerLibraryCardExplicitObservationOwnership(t *testing.T) {
	for _, tc := range []struct {
		name string
		text string
		want []int
	}{
		{"look before reveal", "Look at the top card of your library. Reveal the top card of target opponent's library. Put the looked-at card into your hand.", []int{0}},
		{"reveal before look", "Reveal the top card of your library. Look at the top card of target opponent's library. Put the revealed card into your hand.", []int{0}},
		{"looked card and its owner", "Look at the top card of target player's library. Reveal the top card of target opponent's library. Put the looked-at card into that player's hand.", []int{0, 0}},
		{"revealed card and its owner", "Reveal the top card of target player's library. Look at the top card of target opponent's library. Put the revealed card into that player's hand.", []int{0, 0}},
		{"generic card remains latest", "Look at the top card of your library. Reveal the top card of target opponent's library. Put that card into your hand.", []int{1}},
		{"linked reveal preserves card", "Look at the top card of your library. Reveal it. Put the revealed card into your hand.", []int{0, 0}},
		{"linked reveal of earlier look", "Look at the top card of your library. Reveal the top card of target opponent's library. Reveal the looked-at card. Put the revealed card into your hand.", []int{0, 0}},
		{"independent products keep owners", "Look at the top card of target player's library. Reveal the top card of target opponent's library. Put the looked-at card into that player's hand. Put the revealed card into that player's graveyard.", []int{0, 0, 1, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compilation, diagnostics := compileSource(tc.text, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
				t.Fatalf("compile diagnostics=%v abilities=%d", diagnostics, len(compilation.Abilities))
			}
			content := compilation.Abilities[0].Content
			if len(content.References) != len(tc.want) {
				t.Fatalf("references=%#v, want prior instructions %v", content.References, tc.want)
			}
			references := append([]CompiledReference(nil), content.References...)
			effects := append([]CompiledEffect(nil), content.Effects...)
			for i := range references {
				wantBinding := ReferenceBindingPriorInstructionResult
				if references[i].Kind == ReferenceThatPlayer {
					wantBinding = ReferenceBindingLibraryOwner
				}
				if references[i].Binding != wantBinding || references[i].PriorInstruction != tc.want[i] ||
					references[i].ProducerClauseID != effects[tc.want[i]].ClauseID {
					t.Errorf("reference %d has binding %v and producer %d/%d, want %v/%d/%d",
						references[i].NodeID, references[i].Binding, references[i].PriorInstruction,
						references[i].ProducerClauseID, wantBinding, tc.want[i], effects[tc.want[i]].ClauseID)
				}
				references[i].Binding = ReferenceBindingUnsupported
				references[i].PriorInstruction = -1
				references[i].Text = "opaque reference"
				references[i].Span = shared.Span{}
			}
			for i := range effects {
				effects[i].Text = "opaque effect"
				effects[i].Span, effects[i].ClauseSpan, effects[i].VerbSpan = shared.Span{}, shared.Span{}, shared.Span{}
			}
			rebound := bindReferences(references, content.Targets, effects, nil)
			for i := range rebound {
				if rebound[i].Binding != content.References[i].Binding || rebound[i].PriorInstruction != tc.want[i] ||
					rebound[i].NodeID != content.References[i].NodeID ||
					rebound[i].ProducerClauseID != effects[tc.want[i]].ClauseID {
					t.Errorf("opaque/zero-span reference changed exact identity: before=%#v after=%#v", content.References[i], rebound[i])
				}
			}
		})
	}
}
