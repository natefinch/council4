package agent

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/rules"
)

func TestFirstLegalChoosesFirstAction(t *testing.T) {
	want := action.PlayLand(id.ID(42))

	got := FirstLegal{}.ChooseAction(rules.PlayerObservation{}, []action.Action{
		want,
		action.Pass(),
	})

	if got.Kind != want.Kind || got.PlayLand != want.PlayLand {
		t.Fatalf("ChooseAction() = %+v, want %+v", got, want)
	}
}

func TestFirstLegalPassesWithNoLegalActions(t *testing.T) {
	got := FirstLegal{}.ChooseAction(rules.PlayerObservation{}, nil)
	if got.Kind != action.ActionPass {
		t.Fatalf("ChooseAction() kind = %v, want %v", got.Kind, action.ActionPass)
	}
}
