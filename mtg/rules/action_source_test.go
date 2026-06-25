package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestRecordActionSourceSnapshotsAbilitySource(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	permanent := addCreaturePermanentWithSupertype(g, game.Player1)

	actionLog := &ActionLog{Player: game.Player1}
	recordActionSource(g, game.Player1, actionLog, action.ActivateAbility(permanent.ObjectID, 0, nil, 0))

	if got := actionLog.PermanentSources[permanent.ObjectID]; got != permanent.CardInstanceID {
		t.Fatalf("PermanentSources[%v] = %v, want %v", permanent.ObjectID, got, permanent.CardInstanceID)
	}
}

func TestRecordActionSourceIgnoresNonAbilityActions(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	actionLog := &ActionLog{Player: game.Player1}
	recordActionSource(g, game.Player1, actionLog, action.Pass())
	if len(actionLog.PermanentSources) != 0 {
		t.Fatalf("PermanentSources = %v, want empty", actionLog.PermanentSources)
	}
	if actionLog.ManaAbility {
		t.Fatal("ManaAbility = true, want false for a pass")
	}
}

func TestRecordActionSourceFlagsManaAbility(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	forest := addBasicLandWithManaPermanent(g, game.Player1, "Forest", types.Forest, mana.G)

	actionLog := &ActionLog{Player: game.Player1}
	recordActionSource(g, game.Player1, actionLog, action.ActivateAbility(forest.ObjectID, 0, nil, 0))

	if !actionLog.ManaAbility {
		t.Fatal("ManaAbility = false, want true for a basic land's mana ability")
	}
}
