package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/zone"
)

const chosenLandSearch = "Search your library for up to two basic land cards which have the same name as the chosen land, put them onto the battlefield tapped, then shuffle."

func compileChosenNameBody(t *testing.T, text string) AbilityContent {
	t.Helper()
	document, diagnostics := parser.Parse(text, parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 || len(compilation.Abilities) != 1 {
		t.Fatal(diagnostics)
	}
	return compilation.Abilities[0].Content
}

func TestChosenPermanentNameAuthorityRejectsChangedOrForeignFacts(t *testing.T) {
	const text = "Choose a land you control. You gain 2 life. " + chosenLandSearch
	for _, tc := range []struct {
		name   string
		mutate func(*AbilityContent)
	}{
		{"missing canonical reference", func(b *AbilityContent) { b.References = nil }},
		{"missing canonical proof", func(b *AbilityContent) { b.References[0].Subject = ReferenceSubjectProof{} }},
		{"duplicate canonical reference", func(b *AbilityContent) { b.References = append(b.References, b.References[0]) }},
		{"missing choice", func(b *AbilityContent) { b.Effects[0].ClauseID = 0 }},
		{"duplicate choice ID", func(b *AbilityContent) { b.Effects[1].ClauseID = b.Effects[0].ClauseID }},
		{"forward choice", func(b *AbilityContent) { b.Effects[0].VerbOrder = shared.SourceOrder{Start: 90, End: 91} }},
		{"default choice order", func(b *AbilityContent) { b.Effects[0].VerbOrder = shared.SourceOrder{} }},
		{"competing choice", func(b *AbilityContent) { b.Effects[1].Kind = EffectChoosePermanent }},
		{"wrong noun", func(b *AbilityContent) { b.References[0].SubjectNoun = parser.ObjectNounCreature }},
		{"card identity", func(b *AbilityContent) { b.References[0].CardIdentity = true }},
		{"library observation", func(b *AbilityContent) { b.References[0].LibraryCardObservation = true }},
		{"reached card", func(b *AbilityContent) { b.References[0].ReachedCardProducerClauseID = 8 }},
		{"foreign choice", func(b *AbilityContent) { b.Effects[0] = compileChosenNameBody(t, text).Effects[0] }},
		{"choice kind", func(b *AbilityContent) { b.Effects[0].Kind = EffectDraw }},
		{"choice actor", func(b *AbilityContent) { b.Effects[0].Context = parser.EffectContextEachOpponent }},
		{"choice controller", func(b *AbilityContent) { b.Effects[0].Selector.Controller = ControllerOpponent }},
		{"choice noun", func(b *AbilityContent) { b.Effects[0].Selector.Kind = SelectorCreature }},
		{"choice source", func(b *AbilityContent) { b.Effects[0].CardSource = parser.EffectCardSourceTopOfPlayerLibrary }},
		{"choice zone", func(b *AbilityContent) { b.Effects[0].Selector.Zone = zone.Graveyard }},
		{"optional choice", func(b *AbilityContent) { b.Effects[0].Optional = true }},
		{"delayed choice", func(b *AbilityContent) { b.Effects[0].DelayedTiming++ }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := compileChosenNameBody(t, text)
			reference := body.Effects[2].References[0]
			if !body.OwnsSubject(reference) {
				t.Fatal("missing original authority")
			}
			// The authority owns the original body, not an unrelated value copy.
			owner := reference.Subject.scope.owner
			tc.mutate(owner)
			if owner.OwnsSubject(reference) || reference.SubjectProducerMatches(owner.Effects[0]) {
				t.Fatal("changed canonical facts retained chosen-name authority")
			}
			resolveChosenPermanentSearchNames(owner)
			if owner.Effects[2].SearchNameFromChosenPermanent {
				t.Fatal("changed authority retained same-name operand")
			}
		})
	}
	for _, tc := range []struct {
		name   string
		mutate func(*CompiledReference)
	}{
		{"default", func(r *CompiledReference) { *r = CompiledReference{} }},
		{"missing proof", func(r *CompiledReference) { r.Subject = ReferenceSubjectProof{} }},
		{"node", func(r *CompiledReference) { r.NodeID++ }},
		{"choice ID", func(r *CompiledReference) { r.ChoiceClauseID++ }},
		{"missing choice ID", func(r *CompiledReference) { r.ChoiceClauseID = 0 }},
		{"producer ID", func(r *CompiledReference) { r.ProducerClauseID++ }},
		{"order", func(r *CompiledReference) { r.Order = shared.SourceOrder{} }},
		{"source binding", func(r *CompiledReference) { r.Binding = ReferenceBindingSource }},
		{"card noun", func(r *CompiledReference) { r.SubjectNoun = parser.ObjectNounCard }},
		{"player noun", func(r *CompiledReference) { r.SubjectNoun = parser.ObjectNounPlayer }},
		{"stack noun", func(r *CompiledReference) { r.SubjectNoun = parser.ObjectNounSpell }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := compileChosenNameBody(t, text)
			reference := body.Effects[2].References[0]
			tc.mutate(&reference)
			if reference.SubjectSupported() || body.OwnsSubject(reference) || reference.SubjectProducerMatches(body.Effects[0]) {
				t.Fatal("forged/default subject accepted")
			}
		})
	}
}

func TestChosenPermanentNameSkippedChoiceScopesRefuse(t *testing.T) {
	for _, text := range []string{
		"If you control a creature, choose a land you control. " + chosenLandSearch,
		"If you control a creature, choose a land you control. You gain 2 life. " + chosenLandSearch,
		"Choose a land you control. If you control a creature, choose a land you control. " + chosenLandSearch,
		"You may choose a land you control. " + chosenLandSearch,
		chosenLandSearch + " Choose a land you control.",
	} {
		t.Run(text, func(t *testing.T) {
			body := compileChosenNameBody(t, text)
			for _, effect := range body.Effects {
				if effect.Kind == EffectSearch && effect.SearchNameFromChosenPermanent {
					t.Fatal("skipped/ambiguous/forward choice received authority")
				}
			}
		})
	}
}
func TestChosenPermanentNameAcrossIndependentClauses(t *testing.T) {
	for _, text := range []string{
		"Choose a land you control. " + chosenLandSearch,
		"Choose a land you control. You gain 2 life. " + chosenLandSearch,
		"You gain 2 life. Choose a land you control. " + chosenLandSearch,
		"Choose a land you control. " + chosenLandSearch + " You gain 2 life.",
		"You gain 1 life. Choose a land you control. You gain 2 life. " + chosenLandSearch + " You gain 3 life.",
		"Choose a land you control. You gain 2 life. " + chosenLandSearch + " You gain 3 life. " + chosenLandSearch,
	} {
		t.Run(text, func(t *testing.T) {
			body := compileChosenNameBody(t, text)
			choice, searches := -1, 0
			for i, effect := range body.Effects {
				if effect.Kind == EffectChoosePermanent {
					choice = i
				}

			}
			for _, effect := range body.Effects {
				if effect.Kind != EffectSearch {
					continue
				}
				searches++
				if !effect.SearchNameFromChosenPermanent || effect.UnsupportedDetail != "" {
					t.Fatalf("same-name authority missing: flag=%v detail=%q references=%+v", effect.SearchNameFromChosenPermanent, effect.UnsupportedDetail, effect.References)
				}
				found := 0
				for _, reference := range effect.References {
					if reference.Kind == ReferenceThatObject {
						found++
						if reference.PriorInstruction != choice || !body.OwnsSubject(reference) ||
							!reference.SubjectProducerMatches(body.Effects[choice]) {
							t.Fatalf("not the canonical original choice: %+v", reference)
						}
					}
				}
				if found != 1 {
					t.Fatalf("chosen-name references=%d", found)
				}
			}
			if searches == 0 {
				t.Fatal("missing search")
			}
		})
	}
}

func TestChosenPermanentNameTypedTransportWithoutDisplayMetadata(t *testing.T) {
	document, diagnostics := parser.Parse("Choose a land you control. You gain 2 life. "+chosenLandSearch,
		parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	ability := &document.Abilities[0]
	for i := range ability.SemanticReferences {
		reference := &ability.SemanticReferences[i]
		reference.Text, reference.Tokens, reference.Span = "opaque", nil, shared.Span{}
	}
	for si := range ability.Sentences {
		for ei := range ability.Sentences[si].Effects {
			effect := &ability.Sentences[si].Effects[ei]
			effect.Text, effect.Tokens = "opaque", nil
			effect.Span, effect.ClauseSpan, effect.VerbSpan = shared.Span{}, shared.Span{}, shared.Span{}
		}
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	body := compilation.Abilities[0].Content
	search := body.Effects[2]
	if !search.SearchNameFromChosenPermanent {
		t.Fatal("typed choice authority depended on downstream text or byte spans")
	}
	reference := search.References[0]
	if !body.OwnsSubject(reference) || reference.ChoiceClauseID != body.Effects[0].ClauseID ||
		!reference.SubjectProducerMatches(body.Effects[0]) {
		t.Fatal("typed transport lost canonical choice identity")
	}
}
