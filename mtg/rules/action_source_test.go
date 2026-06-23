package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

func TestRecordActionSourceSnapshotsAbilitySource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCreaturePermanentWithSupertype(g, game.Player1)

	actionLog := &ActionLog{Player: game.Player1}
	recordActionSource(g, actionLog, action.ActivateAbility(permanent.ObjectID, 0, nil, 0))

	if got := actionLog.PermanentSources[permanent.ObjectID]; got != permanent.CardInstanceID {
		t.Fatalf("PermanentSources[%v] = %v, want %v", permanent.ObjectID, got, permanent.CardInstanceID)
	}
}

func TestRecordActionSourceIgnoresNonAbilityActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	actionLog := &ActionLog{Player: game.Player1}
	recordActionSource(g, actionLog, action.Pass())
	if len(actionLog.PermanentSources) != 0 {
		t.Fatalf("PermanentSources = %v, want empty", actionLog.PermanentSources)
	}
}
