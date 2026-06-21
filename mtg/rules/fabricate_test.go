package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
)

func addFabricatePermanent(g *game.Game, controller game.PlayerID, count int) *game.Permanent {
	ability := game.FabricateTriggeredAbility(count)
	def := &game.CardDef{CardFace: game.CardFace{
		Name:               "Fabricate Creature",
		Types:              []types.Card{types.Creature},
		TriggeredAbilities: []game.TriggeredAbility{ability},
	}}
	return addCombatPermanent(g, controller, def)
}

func TestFabricateModalEntryChoosesCountersOrServoTokens(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		count       int
		mode        int
		wantCounter int
		wantTokens  int
	}{
		{name: "counters", count: 2, mode: 0, wantCounter: 2, wantTokens: 0},
		{name: "tokens", count: 2, mode: 1, wantCounter: 0, wantTokens: 2},
		{name: "single counter", count: 1, mode: 0, wantCounter: 1, wantTokens: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			source := addFabricatePermanent(g, game.Player1, test.count)

			emitEvent(g, game.Event{Kind: game.EventPermanentEnteredBattlefield, PermanentID: source.ObjectID})
			agents := [game.NumPlayers]PlayerAgent{
				game.Player1: &choiceOnlyAgent{choices: [][]int{{test.mode}}},
			}
			log := TurnLog{}
			if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, &log) {
				t.Fatal("fabricate entry trigger was not put on the stack")
			}
			engine.resolveTopOfStackWithChoices(g, agents, &log)

			if got := source.Counters.Get(counter.PlusOnePlusOne); got != test.wantCounter {
				t.Fatalf("+1/+1 counters = %d; want %d", got, test.wantCounter)
			}
			tokens := 0
			for _, permanent := range g.Battlefield {
				if permanent.Token && permanent.TokenDef != nil && permanent.TokenDef.Name == "Servo" {
					tokens++
				}
			}
			if tokens != test.wantTokens {
				t.Fatalf("Servo tokens = %d; want %d", tokens, test.wantTokens)
			}
		})
	}
}
