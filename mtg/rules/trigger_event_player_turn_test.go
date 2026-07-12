package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestTriggerEventPlayerTurnMatchesDrawAndCastActor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		event game.Event
	}{
		{
			name:  "draw player",
			event: game.Event{Kind: game.EventCardDrawn, Player: game.Player2},
		},
		{
			name:  "spell controller",
			event: game.Event{Kind: game.EventSpellCast, Controller: game.Player2},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			g.Turn.ActivePlayer = game.Player3
			if triggerCastDuringTurnMatches(g, game.Player1, game.TriggerTurnEventPlayer, test.event) {
				t.Fatal("event-player turn matched a different active player")
			}
			g.Turn.ActivePlayer = game.Player2
			if !triggerCastDuringTurnMatches(g, game.Player1, game.TriggerTurnEventPlayer, test.event) {
				t.Fatal("event-player turn did not match the triggering player's turn")
			}
		})
	}
}
