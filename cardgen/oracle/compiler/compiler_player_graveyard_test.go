package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

// TestCompilePlayerGraveyardExile confirms the parser's recognized
// whole-graveyard exile owner relation flows into the compiled effect so
// lowering can build the target-player MoveCard.
func TestCompilePlayerGraveyardExile(t *testing.T) {
	t.Parallel()
	tests := []struct {
		source string
		owner  parser.GraveyardZoneExileKind
	}{
		{"Exile target player's graveyard.", parser.GraveyardZoneExileTargetPlayer},
		{"Exile target opponent's graveyard.", parser.GraveyardZoneExileTargetOpponent},
	}
	for _, test := range tests {
		t.Run(test.source, func(t *testing.T) {
			t.Parallel()
			compilation, diagnostics := compileSource(test.source, pipelineContext{InstantOrSorcery: true})
			if len(diagnostics) != 0 {
				t.Fatalf("diagnostics = %#v", diagnostics)
			}
			effects := compilation.Abilities[0].Content.Effects
			if len(effects) != 1 {
				t.Fatalf("effects = %#v, want one", effects)
			}
			if effects[0].Kind != EffectExile {
				t.Fatalf("kind = %v, want EffectExile", effects[0].Kind)
			}
			if effects[0].GraveyardZoneExile != test.owner {
				t.Fatalf("owner = %q, want %q", effects[0].GraveyardZoneExile, test.owner)
			}
		})
	}
}
