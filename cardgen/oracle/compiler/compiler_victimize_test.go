package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

const victimizeOracleText = "Choose two target creature cards in your graveyard. Sacrifice a creature. If you do, return the chosen cards to the battlefield tapped."

func TestCompileChosenCardsReferenceBindsTargetGroup(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		victimizeOracleText,
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Targets) != 1 ||
		content.Targets[0].Cardinality.Min != 2 ||
		content.Targets[0].Cardinality.Max != 2 {
		t.Fatalf("targets = %#v; want one exactly-two target group", content.Targets)
	}
	if len(content.References) != 1 {
		t.Fatalf("references = %#v; want one chosen-cards reference", content.References)
	}
	reference := content.References[0]
	if reference.Kind != ReferenceChosenCards ||
		reference.Binding != ReferenceBindingTarget ||
		reference.Occurrence != 0 {
		t.Fatalf("reference = %#v; want chosen-cards binding to target group 0", reference)
	}
	if len(content.Effects) != 2 ||
		!content.Effects[1].Exact ||
		len(content.Effects[1].References) != 1 ||
		content.Effects[1].References[0].NodeID != reference.NodeID {
		t.Fatalf("effects = %#v; want exact return preserving reference identity", content.Effects)
	}
}

func TestCompileChosenCardsReferencePreservesVariableTargetCardinality(t *testing.T) {
	t.Parallel()
	compilation, diagnostics := compileSource(
		"Choose up to two target creature cards in your graveyard. Return the chosen cards to the battlefield tapped.",
		pipelineContext{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.Targets) != 1 ||
		content.Targets[0].Cardinality.Min != 0 ||
		content.Targets[0].Cardinality.Max != 2 {
		t.Fatalf("targets = %#v; want one zero-to-two target group", content.Targets)
	}
	if len(content.References) != 1 ||
		content.References[0].Kind != ReferenceChosenCards ||
		content.References[0].Binding != ReferenceBindingTarget ||
		content.References[0].Occurrence != 0 {
		t.Fatalf("references = %#v; want chosen-cards binding to variable target group", content.References)
	}
}

func TestCompileChosenCardsReferenceIsTextBlind(t *testing.T) {
	t.Parallel()
	document, diagnostics := parser.Parse(
		victimizeOracleText,
		parser.Context{InstantOrSorcery: true},
	)
	if len(diagnostics) != 0 {
		t.Fatalf("parse diagnostics = %#v", diagnostics)
	}
	document.Abilities[0].Text = "compiler must not inspect this"
	document.Abilities[0].Tokens = nil
	document.Abilities[0].Sentences[0].Targets[0].Text = "or target text"
	returnEffect := &document.Abilities[0].Sentences[2].Effects[0]
	returnEffect.Text = "or effect text"
	returnEffect.Tokens = nil
	returnEffect.References[0].Text = "or reference text"
	document.Abilities[0].SemanticReferences[0].Text = "or semantic reference text"

	compilation, diagnostics := Compile(document, Context{})
	if len(diagnostics) != 0 {
		t.Fatalf("compile diagnostics = %#v", diagnostics)
	}
	content := compilation.Abilities[0].Content
	if len(content.References) != 1 ||
		content.References[0].Kind != ReferenceChosenCards ||
		content.References[0].Binding != ReferenceBindingTarget ||
		len(content.Effects) != 2 ||
		!content.Effects[1].Exact {
		t.Fatalf("content = %#v; want typed chosen-card semantics", content)
	}
}
