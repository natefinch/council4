package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestReferenceSubjectEnteredObservation(t *testing.T) {
	compilation, diagnostics := compileSource(
		"At the beginning of your upkeep, reveal the top card of your library. If it's a creature card, put it onto the battlefield. That creature gains haste until end of turn. Sacrifice it at the beginning of the next end step.",
		pipelineContext{CardName: "Killer Instinct"},
	)
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	content := compilation.Abilities[0].Content
	for _, target := range content.Targets {
		t.Logf("compiled target=%v zone=%v exact=%v domain=%v", target.Selector.Kind,
			target.Selector.Zone, target.Exact, targetSubjectDomain(target))
	}
	for _, effect := range content.Effects {
		t.Logf("effect=%v clause=%d exact=%v source=%v from=%v to=%v", effect.Kind,
			effect.ClauseID, effect.Exact, effect.CardSource, effect.FromZone, effect.ToZone)
	}
	found := false
	for _, reference := range content.References {
		t.Logf("reference=%d kind=%v noun=%v card=%v binding=%v producer=%d observation=%v prior=%d domain=%v",
			reference.NodeID, reference.Kind, reference.SubjectNoun, reference.CardIdentity,
			reference.Binding, reference.ProducerClauseID, reference.LibraryCardObservation,
			reference.PriorInstruction, reference.SubjectDomain())
		if reference.Kind != ReferenceThatObject || reference.SubjectNoun != parser.ObjectNounCreature {
			continue
		}
		found = true
		if reference.Binding != ReferenceBindingPriorInstructionResult || reference.PriorInstruction != 1 ||
			reference.SubjectDomain() != ReferenceSubjectPermanent ||
			reference.SubjectLifetime() != ReferenceLifetimeActualProduct {
			t.Error("post-placement creature does not name the actual entered permanent")
		}
	}
	if !found {
		t.Fatal("complete original lost its explicit creature subject")
	}
}

func TestReferenceSubjectEnteredTarget(t *testing.T) {
	for _, text := range []string{
		"Return target creature card from your graveyard to the battlefield. That creature gains haste until end of turn.",
		"You may return target creature card from your graveyard to the battlefield and gain 1 life. It gains haste until end of turn.",
	} {
		t.Run(text, func(t *testing.T) {
			document, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			for _, sentence := range document.Abilities[0].Sentences {
				for _, effect := range sentence.Effects {
					for _, target := range effect.Targets {
						t.Logf("parser producer=%d kind=%v exact=%v to=%v target=%v zone=%v count=%v",
							effect.ClauseID, effect.Kind, effect.Exact, effect.ToZone,
							target.Selection.Kind, target.Selection.Zone, target.Cardinality)
					}
				}
			}
			compilation, diagnostics := Compile(document, Context{})
			if len(diagnostics) != 0 {
				t.Fatal(diagnostics)
			}
			content := compilation.Abilities[0].Content
			for _, reference := range content.References {
				t.Logf("reference=%d binding=%v producer=%d prior=%d domain=%v",
					reference.NodeID, reference.Binding, reference.ProducerClauseID,
					reference.PriorInstruction, reference.SubjectDomain())
				if reference.Binding != ReferenceBindingPriorInstructionResult ||
					reference.PriorInstruction != 0 || reference.SubjectDomain() != ReferenceSubjectPermanent {
					t.Error("later creature subject does not name the actual entered target")
				}
				if occurrence, ok := reference.ProducerTargetOccurrence(); !ok || occurrence != 0 {
					t.Error("entered producer lost its exact card-target occurrence")
				}
			}
			if len(content.References) != 1 {
				t.Fatal("normal composition lost its singular subject")
			}
		})
	}
}
