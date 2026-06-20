package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// This text-blind test drives compileEffects with a constructed typed parser
// node only: no Oracle wording or tokens. It proves the compiler propagates the
// source-spell cost-reduction fields purely from typed syntax.

func TestCompileSourceSpellCostReductionFromTypedNodes(t *testing.T) {
	t.Parallel()
	sentences := []parser.Sentence{{
		Effects: []parser.EffectSyntax{{
			Kind:                           parser.EffectCast,
			Context:                        parser.EffectContextSource,
			SourceSpellCostReduction:       true,
			SourceSpellCostReductionAmount: 2,
		}},
	}}
	effects := compileEffects(sentences)
	if len(effects) != 1 {
		t.Fatalf("compiled effects = %d, want 1", len(effects))
	}
	if !effects[0].SourceSpellCostReduction {
		t.Fatal("SourceSpellCostReduction was not propagated to the compiled effect")
	}
	if effects[0].SourceSpellCostReductionAmount != 2 {
		t.Fatalf("SourceSpellCostReductionAmount = %d, want 2", effects[0].SourceSpellCostReductionAmount)
	}
}

func TestCompileSourceSpellCostReductionAbsentByDefault(t *testing.T) {
	t.Parallel()
	sentences := []parser.Sentence{{
		Effects: []parser.EffectSyntax{{
			Kind:    parser.EffectCast,
			Context: parser.EffectContextSource,
		}},
	}}
	effects := compileEffects(sentences)
	if len(effects) != 1 {
		t.Fatalf("compiled effects = %d, want 1", len(effects))
	}
	if effects[0].SourceSpellCostReduction || effects[0].SourceSpellCostReductionAmount != 0 {
		t.Fatalf("unexpected source-spell reduction = %#v", effects[0])
	}
}

func TestCompileSourceSpellCostReductionDynamicFromTypedNodes(t *testing.T) {
	t.Parallel()
	sentences := []parser.Sentence{{
		Effects: []parser.EffectSyntax{{
			Kind:                            parser.EffectCast,
			Context:                         parser.EffectContextSource,
			SourceSpellCostReductionDynamic: true,
		}},
	}}
	effects := compileEffects(sentences)
	if len(effects) != 1 {
		t.Fatalf("compiled effects = %d, want 1", len(effects))
	}
	if !effects[0].SourceSpellCostReductionDynamic {
		t.Fatal("SourceSpellCostReductionDynamic was not propagated to the compiled effect")
	}
	if effects[0].SourceSpellCostReduction {
		t.Fatal("dynamic reduction should not set the per-object reduction flag")
	}
}
