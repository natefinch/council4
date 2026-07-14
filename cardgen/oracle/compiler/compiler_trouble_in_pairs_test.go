package compiler

import (
	"testing"

	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func TestCompileTroubleInPairsMarkers(t *testing.T) {
	t.Parallel()
	const source = "If an opponent would begin an extra turn, that player skips that turn instead.\nWhenever an opponent attacks you with two or more creatures, draws their second card each turn, or casts their second spell each turn, you draw a card."
	compilation, diagnostics := compileSource(source, pipelineContext{CardName: "Trouble in Pairs"})
	if len(diagnostics) != 0 || len(compilation.Abilities) != 2 {
		t.Fatalf("compilation = %#v, diagnostics = %#v", compilation, diagnostics)
	}
	if compilation.Abilities[0].SkipExtraTurnsScope != parser.TriggerPlayerSelectorOpponent ||
		!compilation.Abilities[1].OpponentSecondActionTriplet {
		t.Fatalf("abilities = %#v", compilation.Abilities)
	}
}
