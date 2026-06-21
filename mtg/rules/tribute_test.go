package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func tributeCreatureDef(count int) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:      "Tribute Bearer",
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		ReplacementAbilities: []game.ReplacementAbility{
			game.TributeReplacement("As this creature enters, an opponent of your choice may put 3 +1/+1 counters on it.", count),
		},
	}}
}

// soleOpponentGame eliminates every player but Player1 and Player2 so a Tribute
// replacement faces exactly one opponent and prompts only that opponent's
// pay/decline choice (no controller "choose an opponent" step).
func soleOpponentGame() *game.Game {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for p := game.Player3; int(p) < game.NumPlayers; p++ {
		g.Players[p].Eliminated = true
	}
	return g
}

// TestTributePaidAddsCounters verifies that when the chosen opponent pays a
// creature's Tribute, the creature enters with N +1/+1 counters and its
// TributePaid flag is set.
func TestTributePaidAddsCounters(t *testing.T) {
	g := soleOpponentGame()
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}}}
	permanent := enterRiotCreature(t, g, engine, tributeCreatureDef(3), agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 3 {
		t.Fatalf("tribute counters = %d, want 3", got)
	}
	if !permanent.TributePaid {
		t.Fatal("TributePaid = false, want true after opponent paid")
	}
}

// TestTributeDeclinedLeavesFlagUnset verifies that when the chosen opponent
// declines a creature's Tribute, no counters are added and TributePaid stays
// false so a paired "if tribute wasn't paid" ability can react.
func TestTributeDeclinedLeavesFlagUnset(t *testing.T) {
	g := soleOpponentGame()
	engine := NewEngine(nil)
	agents := [game.NumPlayers]PlayerAgent{game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}}}
	permanent := enterRiotCreature(t, g, engine, tributeCreatureDef(3), agents)
	if got := permanent.Counters.Get(counter.PlusOnePlusOne); got != 0 {
		t.Fatalf("tribute counters = %d, want 0", got)
	}
	if permanent.TributePaid {
		t.Fatal("TributePaid = true, want false after opponent declined")
	}
}
