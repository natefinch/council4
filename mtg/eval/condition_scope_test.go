package eval

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestScorableCapturedConditionIsUncertain(t *testing.T) {
	t.Parallel()
	content := game.Mode{Sequence: []game.Instruction{{
		Primitive:     game.GainLife{Amount: game.Fixed(2), Player: game.ControllerReference()},
		ConditionGate: "scope",
	}}}.Ability()
	atoms := ScorableEffect(content)
	if len(atoms) != 1 || !atoms[0].IsDynamic {
		t.Fatalf("atoms=%#v, want captured condition flagged uncertain", atoms)
	}
}
