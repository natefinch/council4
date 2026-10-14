package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestDeclaredPermanentReadbackBaseline(t *testing.T) {
	for _, action := range []string{
		"Return target creature to its owner's hand.",
		"Destroy target creature.",
		"Tap target creature.",
	} {
		for _, prefix := range []string{"", "Exile target card from a graveyard. ", "Target player gains 1 life. ", "Counter target spell. "} {
			t.Run(prefix+action, func(t *testing.T) {
				content := compiledTargetActionBody(t, prefix+action+" Its controller loses 2 life. Its owner gains 3 life.")
				want := len(content.Targets) - 1
				for _, effect := range content.Effects[len(content.Effects)-2:] {
					for _, reference := range effect.References {
						if reference.Binding != ReferenceBindingTarget || reference.Occurrence != want ||
							reference.ObjectProjectionDomain() != ReferenceSubjectPermanent ||
							!reference.SubjectSupported() {
							t.Fatalf("declared target %d readback: %+v", want, reference)
						}
					}
				}
			})
		}
	}
}
func TestDeclaredPermanentPrivateProofRefusesDrift(t *testing.T) {
	const text = "Exile target card from a graveyard. Tap target creature. You gain 1 life. Its controller loses 2 life."
	for _, test := range []struct {
		name string
		edit func(*AbilityContent, *CompiledReference)
	}{
		{"missing proof", func(_ *AbilityContent, r *CompiledReference) { r.Subject = ReferenceSubjectProof{} }},
		{"foreign scope", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope = compiledTargetActionBody(t, text).subjectScope
		}},
		{"foreign reference", func(_ *AbilityContent, r *CompiledReference) { *r = compiledTargetActionBody(t, text).References[0] }},
		{"missing identity", func(_ *AbilityContent, r *CompiledReference) { r.DeclaredPermanent = nil }},
		{"default identity", func(_ *AbilityContent, r *CompiledReference) {
			r.DeclaredPermanent = &parser.DeclaredPermanentSubject{}
		}},
		{"wrong producer", func(_ *AbilityContent, r *CompiledReference) {
			r.DeclaredPermanent = &parser.DeclaredPermanentSubject{ClauseID: 1}
		}},
		{"result instead", func(_ *AbilityContent, r *CompiledReference) { r.Binding = ReferenceBindingPriorInstructionResult }},
		{"library product", func(_ *AbilityContent, r *CompiledReference) { r.ProducerClauseID = 1 }},
		{"canonical result lineage", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.References[0].ProducerClauseID = 1
		}},
		{"canonical library observation", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.References[0].LibraryCardObservation = true
		}},
		{"canonical card identity", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.References[0].CardIdentity = true
		}},
		{"canonical reached lineage", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.References[0].ReachedCardProducerClauseID = 1
		}},
		{"foreign node", func(_ *AbilityContent, r *CompiledReference) { r.NodeID++ }},
		{"default order", func(_ *AbilityContent, r *CompiledReference) { r.Order = shared.SourceOrder{} }},
		{"foreign order", func(_ *AbilityContent, r *CompiledReference) { r.Order.Start++ }},
		{"plural", func(_ *AbilityContent, r *CompiledReference) { r.Pronoun = ReferencePronounTheir }},
		{"card noun", func(_ *AbilityContent, r *CompiledReference) { r.SubjectNoun = parser.ObjectNounCard }},
		{"foreign declaration", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.Effects[1] = compiledTargetActionBody(t, text).Effects[1]
		}},
		{"duplicate producer", func(c *AbilityContent, _ *CompiledReference) {
			o := c.subjectScope.owner
			o.Effects = append(o.Effects, o.Effects[1])
		}},
		{"missing producer", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.Effects[1].ClauseID = 0 }},
		{"changed producer kind", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.Effects[1].Kind = EffectDestroy }},
		{"producer not exact", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.Effects[1].Exact = false }},
		{"forward producer", func(c *AbilityContent, r *CompiledReference) {
			c.subjectScope.owner.Effects[1].Order.End = r.Order.End + 1
		}},
		{"foreign consumer", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.Effects[3] = compiledTargetActionBody(t, text).Effects[3]
		}},
		{"consumer missing", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.Effects[3].References = nil }},
		{"consumer context drift", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.Effects[3].Context = parser.EffectContextReferencedObjectOwner
		}},
		{"consumer node order drift", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.Effects[3].References[0].Order.Start++
		}},
		{"duplicate consumer reference", func(c *AbilityContent, _ *CompiledReference) {
			e := &c.subjectScope.owner.Effects[3]
			e.References = append(e.References, e.References[0])
		}},
		{"consumer clause drift", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.Effects[3].ClauseID++ }},
		{"duplicate consumer", func(c *AbilityContent, _ *CompiledReference) {
			o := c.subjectScope.owner
			o.Effects = append(o.Effects, o.Effects[3])
		}},
		{"duplicate canonical reference", func(c *AbilityContent, _ *CompiledReference) {
			o := c.subjectScope.owner
			o.References = append(o.References, o.References[0])
		}},
		{"missing canonical reference", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.References = nil }},
		{"foreign target", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.Targets[1] = compiledTargetActionBody(t, text).Targets[1]
		}},
		{"missing target", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.Targets = c.subjectScope.owner.Targets[:1]
		}},
		{"duplicate target", func(c *AbilityContent, _ *CompiledReference) {
			o := c.subjectScope.owner
			o.Targets = append(o.Targets, o.Targets[1])
		}},
		{"target plural", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.Targets[1].Cardinality.Max = 2 }},
		{"target optional width", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.Targets[1].Cardinality.Min = 0 }},
		{"target nonexact", func(c *AbilityContent, _ *CompiledReference) { c.subjectScope.owner.Targets[1].Exact = false }},
		{"target card domain", func(c *AbilityContent, _ *CompiledReference) {
			c.subjectScope.owner.Targets[1].Selector.Kind = SelectorCard
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			content := compiledTargetActionBody(t, text)
			reference := content.References[0]
			if !content.OwnsSubject(reference) {
				t.Fatal("control lacks canonical declared target proof")
			}
			test.edit(&content, &reference)
			if content.OwnsSubject(reference) {
				t.Fatal("foreign/default/ambiguous identity retained authority")
			}
		})
	}
}

func TestDeclaredPermanentProofLocalizesOnlyAdapterSlot(t *testing.T) {
	content := compiledTargetActionBody(t, "Exile target card from a graveyard. Tap target creature. Its owner gains 3 life.")
	reference := content.References[0]
	target := content.Targets[1]
	local, ok := reference.LocalizeSubjectOccurrence(target, 0)
	if !ok || !content.OwnsSubject(local) || local.Occurrence != 0 ||
		local.Subject.targetOccurrence != 1 || local.DeclaredPermanent.ClauseID != 2 {
		t.Fatalf("exact nonzero original target not preserved: %+v", local)
	}
	if _, ok := reference.LocalizeSubjectOccurrence(content.Targets[0], 0); ok {
		t.Fatal("prior CARD occurrence borrowed permanent subject")
	}
}

func TestDeclaredPermanentAndChosenNameNamespacesRefuseContradiction(t *testing.T) {
	declared := compiledTargetActionBody(t, "Exile target card from a graveyard. Tap target creature. Its owner gains 3 life.")
	declaredReference := declared.References[0]
	if !declared.OwnsSubject(declaredReference) || declaredReference.Binding != ReferenceBindingTarget {
		t.Fatal("declared control lacks target authority")
	}
	declaredReference.ChoiceClauseID = declaredReference.DeclaredPermanent.ClauseID
	if declaredReference.SubjectSupported() || declared.OwnsSubject(declaredReference) {
		t.Fatal("declared target borrowed chosen-name producer authority")
	}

	chosen := compileChosenNameBody(t, "Choose a land you control. You gain 2 life. "+chosenLandSearch)
	chosenReference := chosen.Effects[2].References[0]
	if !chosen.OwnsSubject(chosenReference) || chosenReference.Binding != ReferenceBindingPriorInstructionResult {
		t.Fatal("chosen-name control lacks product authority")
	}
	chosenReference.DeclaredPermanent = &parser.DeclaredPermanentSubject{ClauseID: chosenReference.ChoiceClauseID}
	if chosenReference.SubjectSupported() || chosen.OwnsSubject(chosenReference) {
		t.Fatal("chosen-name product borrowed original declared-target authority")
	}
}

func TestDeclaredPermanentDoesNotUseRetainedText(t *testing.T) {
	document, diagnostics := parser.Parse("Exile target card from a graveyard. Tap target creature. Its controller loses 2 life.", parser.Context{InstantOrSorcery: true})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	for i := range document.Abilities[0].SemanticReferences {
		document.Abilities[0].SemanticReferences[i].Text = "foreign card"
	}
	for i := range document.Abilities[0].Sentences {
		for j := range document.Abilities[0].Sentences[i].Effects {
			e := &document.Abilities[0].Sentences[i].Effects[j]
			e.Text = "foreign spell"
		}
	}
	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatal(diagnostics)
	}
	reference := compilation.Abilities[0].Content.References[0]
	if !reference.SubjectSupported() || reference.Occurrence != 1 ||
		reference.ObjectProjectionDomain() != ReferenceSubjectPermanent {
		t.Fatal("retained text altered declaration binding")
	}
}
