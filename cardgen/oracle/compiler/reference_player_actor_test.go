package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestPlayerActorTargetPrecedesUnrelatedCardResult(t *testing.T) {
	t.Parallel()
	for _, verb := range []string{"they gain 2 life", "they lose 2 life", "they draw two cards", "they mill two cards"} {
		t.Run(verb, func(t *testing.T) {
			t.Parallel()
			document, diagnostics := parser.Parse(
				"Exile target creature card from a graveyard. Target opponent sacrifices a creature of their choice. If that player does, "+verb+".",
				parser.Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			compilation, diagnostics := Compile(document, Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			if len(content.Effects) != 3 || len(content.Targets) != 2 {
				t.Fatalf("effects=%d targets=%d", len(content.Effects), len(content.Targets))
			}
			for _, effect := range content.Effects[1:] {
				if len(effect.References) == 0 {
					t.Fatal("player-owned clause lost its references")
				}
				for _, reference := range effect.References {
					if reference.Binding != ReferenceBindingTarget || reference.Occurrence != 1 {
						t.Fatalf("player actor bound to an unrelated card result: %+v", reference)
					}
				}
			}
		})
	}
}

func TestPlayerActorTargetPriorityRequiresTypedProof(t *testing.T) {
	t.Parallel()
	reference := CompiledReference{Kind: ReferencePronoun, Pronoun: ReferencePronounThey, NodeID: 8,
		Order: shared.SourceOrder{Start: 20, End: 21}}
	targets := []CompiledTarget{{
		Exact: true, Selector: CompiledSelector{Kind: SelectorOpponent},
		Cardinality: TargetCardinality{Min: 1, Max: 1}, Order: shared.SourceOrder{Start: 10, End: 11},
	}}
	effects := []CompiledEffect{
		{ClauseID: 1, Kind: EffectExile, VerbOrder: shared.SourceOrder{Start: 2, End: 3}},
		{ClauseID: 2, Context: parser.EffectContextEventPlayer, References: []CompiledReference{reference}},
	}
	if occurrence, ok := playerActorTargetAntecedent(reference, targets, effects, 0); !ok || occurrence != 0 {
		t.Fatalf("typed nearer player target was not preferred: occurrence=%d ok=%v", occurrence, ok)
	}
	zeroID := reference
	zeroID.NodeID = 0
	zeroEffects := append([]CompiledEffect(nil), effects...)
	zeroEffects[1].References = []CompiledReference{zeroID}
	if _, ok := playerActorTargetAntecedent(zeroID, targets, zeroEffects, 0); !ok {
		t.Fatal("valid first reference NodeID zero lost its explicit clause/order ownership")
	}
	for _, tt := range []struct {
		name string
		edit func(*CompiledReference, *[]CompiledTarget, *[]CompiledEffect)
	}{
		{"object pronoun", func(r *CompiledReference, _ *[]CompiledTarget, _ *[]CompiledEffect) { r.Pronoun = ReferencePronounIt }},
		{"card target", func(_ *CompiledReference, targets *[]CompiledTarget, _ *[]CompiledEffect) {
			(*targets)[0].Selector.Kind = SelectorCard
		}},
		{"plural target", func(_ *CompiledReference, targets *[]CompiledTarget, _ *[]CompiledEffect) {
			(*targets)[0].Cardinality.Max = 2
		}},
		{"optional target", func(_ *CompiledReference, targets *[]CompiledTarget, _ *[]CompiledEffect) {
			(*targets)[0].Cardinality.Min = 0
		}},
		{"nonexact target", func(_ *CompiledReference, targets *[]CompiledTarget, _ *[]CompiledEffect) {
			(*targets)[0].Exact = false
		}},
		{"older player target", func(_ *CompiledReference, targets *[]CompiledTarget, _ *[]CompiledEffect) {
			(*targets)[0].Order.Start = 1
		}},
		{"ambiguous players", func(_ *CompiledReference, targets *[]CompiledTarget, _ *[]CompiledEffect) {
			*targets = append(*targets, (*targets)[0])
		}},
		{"unowned reference", func(r *CompiledReference, _ *[]CompiledTarget, _ *[]CompiledEffect) { r.NodeID++ }},
		{"default reference identity", func(r *CompiledReference, _ *[]CompiledTarget, _ *[]CompiledEffect) {
			r.NodeID = 0
			r.Order = shared.SourceOrder{}
		}},
		{"missing owner clause", func(_ *CompiledReference, _ *[]CompiledTarget, effects *[]CompiledEffect) {
			(*effects)[1].ClauseID = 0
		}},
		{"foreign owner order", func(r *CompiledReference, _ *[]CompiledTarget, _ *[]CompiledEffect) { r.Order.Start++ }},
		{"unrelated controller clause", func(_ *CompiledReference, _ *[]CompiledTarget, effects *[]CompiledEffect) {
			(*effects)[1].Context = parser.EffectContextController
		}},
		{"unknown actor", func(_ *CompiledReference, _ *[]CompiledTarget, effects *[]CompiledEffect) {
			(*effects)[1].Context = parser.EffectContextUnknown
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ref := reference
			ts := append([]CompiledTarget(nil), targets...)
			es := append([]CompiledEffect(nil), effects...)
			tt.edit(&ref, &ts, &es)
			if occurrence, ok := playerActorTargetAntecedent(ref, ts, es, 0); ok {
				t.Fatalf("incomplete PLAYER proof preferred occurrence %d", occurrence)
			}
		})
	}
}
