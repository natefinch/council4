package eval

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestAuthoredIntentIsUnknownNotNeutralOrExecutableMetadata(t *testing.T) {
	ref := game.ProgramReference{ID: "authored/control-next-turn", Version: "1"}
	atoms := ScorableEffect(game.Mode{Program: ref, Text: "Control the next turn"}.Ability())
	summary := ScorableAbility{Effect: atoms}
	if !summary.HasUnknownEffect() || !summary.HasUnknownProgram() || len(atoms) != 1 || atoms[0].Program != ref ||
		!strings.Contains(Describe(summary), "evaluation unknown") {
		t.Fatalf("missing explicit valuation uncertainty: %+v", summary)
	}
}
