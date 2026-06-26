package cardgen

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

// TestLowerControllerTurnSelfStatic proves that a leading "During your turn,"
// condition lowers the same self keyword/characteristic static the existing
// conditional self-static machinery produces, gated by the SourceControllerTurn
// runtime condition.
func TestLowerControllerTurnSelfStatic(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		oracleText string
		power      int
		toughness  int
		keywords   []game.Keyword
	}{
		"keyword only": {
			oracleText: "During your turn, this creature has first strike.",
			keywords:   []game.Keyword{game.FirstStrike},
		},
		"power toughness and keyword": {
			oracleText: "During your turn, this creature gets +1/+1 and has trample.",
			power:      1,
			toughness:  1,
			keywords:   []game.Keyword{game.Trample},
		},
		"power toughness only": {
			oracleText: "During your turn, this creature gets +2/+0.",
			power:      2,
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ability := lowerSelfStatic(t, test.oracleText)
			assertSelfCondition(t, &ability, &game.Condition{SourceControllerTurn: true})
			assertSelfContinuous(t, &ability, test.power, test.toughness, test.keywords)
		})
	}
}
