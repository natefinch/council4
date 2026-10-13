package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

func declaredTargetBody(t *testing.T, noun string) AbilityContent {
	t.Helper()
	text := "Exile target card from a graveyard. Return target " + noun + " to its owner's hand."
	document, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	effect := document.Abilities[0].Sentences[1].Effects[0]
	reference := effect.References[0]
	t.Logf("parser clause=%d exact=%v targets=%d reference node=%d order=%v kind=%v pronoun=%v noun=%v producer=%d",
		effect.ClauseID, effect.Exact, len(effect.Targets), reference.NodeID, reference.Order,
		reference.Kind, reference.Pronoun, reference.SubjectNoun, reference.ProducerClauseID)
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	return compilation.Abilities[0].Content
}
func TestClauseDeclaredTargetOwnerProofControls(t *testing.T) {
	for _, noun := range []string{"creature", "creature or Vehicle"} {
		t.Run(noun, func(t *testing.T) {
			for _, tc := range []struct {
				name   string
				mutate func(*AbilityContent, *CompiledReference, AbilityContent)
			}{
				{"no body scope", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) { b.subjectScope = nil }},
				{"no source scope", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) { b.Source.scope = nil }},
				{"foreign reference", func(_ *AbilityContent, r *CompiledReference, f AbilityContent) { *r = f.References[0] }},
				{"default reference", func(_ *AbilityContent, r *CompiledReference, _ AbilityContent) { *r = CompiledReference{} }},
				{"node", func(_ *AbilityContent, r *CompiledReference, _ AbilityContent) { r.NodeID++ }},
				{"empty order", func(_ *AbilityContent, r *CompiledReference, _ AbilityContent) { r.Order = shared.SourceOrder{} }},
				{"player role", func(_ *AbilityContent, r *CompiledReference, _ AbilityContent) {
					r.SubjectNoun = parser.ObjectNounPlayer
				}},
				{"card role", func(_ *AbilityContent, r *CompiledReference, _ AbilityContent) { r.CardIdentity = true }},
				{"resource role", func(_ *AbilityContent, r *CompiledReference, _ AbilityContent) { r.Kind = ReferencePaidCostSubject }},
				{"explicit producer", func(_ *AbilityContent, r *CompiledReference, _ AbilityContent) { r.ProducerClauseID = 1 }},
				{"duplicate reference", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.References = append(b.References, b.References[0])
				}},
				{"foreign clause", func(b *AbilityContent, _ *CompiledReference, f AbilityContent) { b.Effects[1] = f.Effects[1] }},
				{"default clause", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) { b.Effects[1].ClauseID = 0 }},
				{"duplicate clause", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Effects = append(b.Effects, b.Effects[1])
				}},
				{"nonexact clause", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) { b.Effects[1].Exact = false }},
				{"different destination", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Effects[1].ToZone = zone.Battlefield
				}},
				{"source role", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Effects[1].Context = parser.EffectContextSource
				}},
				{"foreign declaration", func(b *AbilityContent, _ *CompiledReference, f AbilityContent) {
					b.Effects[1].Targets[0] = f.Targets[1]
				}},
				{"plural declaration", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Effects[1].Targets[0].Cardinality.Max++
				}},
				{"card declaration", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Effects[1].Targets[0].Selector.Zone = zone.Graveyard
				}},
				{"player declaration", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Effects[1].Targets[0].Selector.Kind = SelectorPlayer
				}},
				{"stack declaration", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Effects[1].Targets[0].Selector.Kind = SelectorSpell
				}},
				{"inexact declaration", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) { b.Effects[1].Targets[0].Exact = false }},
				{"missing canonical declaration", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) { b.Targets = b.Targets[:1] }},
				{"duplicate canonical declaration", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Targets = append(b.Targets, b.Targets[1])
				}},
				{"wrong copied atom", func(b *AbilityContent, _ *CompiledReference, _ AbilityContent) {
					b.Effects[1].References[0].Pronoun = ReferencePronounIt
				}},
			} {
				t.Run(tc.name, func(t *testing.T) {
					body, foreign := declaredTargetBody(t, noun), declaredTargetBody(t, noun)
					body.subjectScope.owner = &body
					reference := body.References[0]
					if occurrence, ok := declaredTargetOwnerOccurrence(body, reference); !ok || occurrence != 1 {
						t.Fatal("normal canonical owner declaration lost its proof")
					}
					tc.mutate(&body, &reference, foreign)
					if _, ok := declaredTargetOwnerOccurrence(body, reference); ok {
						t.Fatal("malformed or foreign owner declaration accepted")
					}
				})
			}
		})
	}
}

func TestClauseDeclaredTargetOwnerDoesNotRebindResultReadbacks(t *testing.T) {
	for _, text := range []string{
		"Exile target card from a graveyard. Put it into its owner's hand.",
		"Exile target creature. Return it to the battlefield under its owner's control.",
	} {
		document, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		compilation, diagnostics := Compile(document, Context{})
		if len(diagnostics) != 0 {
			t.Fatal(diagnostics)
		}
		body := compilation.Abilities[0].Content
		for _, reference := range body.References {
			if reference.Binding != ReferenceBindingPriorInstructionResult || reference.PriorInstruction != 0 {
				t.Fatalf("legitimate prior-exiled result was retargeted: %#v", reference)
			}
		}
	}
}
func TestClauseDeclaredTargetOwnerReference(t *testing.T) {
	for _, noun := range []string{"creature", "creature or Vehicle"} {
		t.Run(noun, func(t *testing.T) {
			body := declaredTargetBody(t, noun)
			effect, target, reference := body.Effects[1], body.Targets[1], body.References[0]
			t.Logf("compiled clause=%d order=%v target exact=%v order=%v kind=%v node=%d refOrder=%v binding=%v occurrence=%d prior=%d",
				effect.ClauseID, effect.Order, target.Exact, target.Order, target.Selector.Kind,
				reference.NodeID, reference.Order, reference.Binding, reference.Occurrence, reference.PriorInstruction)
			if reference.Binding != ReferenceBindingTarget || reference.Occurrence != 1 ||
				reference.SubjectDomain() != ReferenceSubjectPermanent ||
				reference.SubjectLifetime() != ReferenceLifetimeOriginalObject {
				t.Fatal("same-clause owner reference was stolen by unrelated exiled card")
			}
		})
	}
}
