package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileLinkedRevealFieldsFromTypedSyntax(t *testing.T) {
	t.Parallel()
	effects := compileEffects([]parser.Sentence{{Effects: []parser.EffectSyntax{{
		Kind:                 parser.EffectPut,
		Player:               parser.EffectPlayerTargetOwner,
		CardSource:           parser.EffectCardSourcePriorInstructionResult,
		RequirePermanentCard: true,
	}}}})
	if len(effects) != 1 {
		t.Fatalf("effects = %d, want 1", len(effects))
	}
	got := effects[0]
	if got.Player != parser.EffectPlayerTargetOwner ||
		got.CardSource != parser.EffectCardSourcePriorInstructionResult ||
		!got.RequirePermanentCard {
		t.Fatalf("compiled effect = %#v", got)
	}
}
