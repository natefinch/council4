package agent

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
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

func TestSimpleCasterPrefersLandThenNonSelfCast(t *testing.T) {
	land := action.PlayLand(id.ID(1))
	selfCast := action.CastSpell(id.ID(2), []game.Target{game.PlayerTarget(game.Player1)}, 0, nil)
	opponentCast := action.CastSpell(id.ID(2), []game.Target{game.PlayerTarget(game.Player2)}, 0, nil)

	got := SimpleCaster{}.ChooseAction(rules.PlayerObservation{Player: game.Player1}, []action.Action{
		selfCast,
		opponentCast,
		land,
		action.Pass(),
	})
	if got.Kind != action.ActionPlayLand {
		t.Fatalf("SimpleCaster chose %+v, want play land", got)
	}

	got = SimpleCaster{}.ChooseAction(rules.PlayerObservation{Player: game.Player1}, []action.Action{
		selfCast,
		opponentCast,
		action.Pass(),
	})
	if !sameCast(got, opponentCast) {
		t.Fatalf("SimpleCaster chose %+v, want %+v", got, opponentCast)
	}

	got = SimpleCaster{}.ChooseAction(rules.PlayerObservation{Player: game.Player1}, []action.Action{
		selfCast,
		action.Pass(),
	})
	if !sameCast(got, selfCast) {
		t.Fatalf("SimpleCaster chose %+v, want %+v", got, selfCast)
	}
}

func sameCast(a, b action.Action) bool {
	return a.Kind == action.ActionCastSpell &&
		b.Kind == action.ActionCastSpell &&
		a.CastSpell.CardID == b.CastSpell.CardID &&
		slices.Equal(a.CastSpell.Targets, b.CastSpell.Targets)
}
