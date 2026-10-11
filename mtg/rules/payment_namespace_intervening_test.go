package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestCompiledPaymentActorAcrossIndependentRider(t *testing.T) {
	def := compileCounterTaxCard(t, "Intervening Payment Actor",
		"If you control a creature, counter target spell unless its controller pays {3}. You gain 1 life. If they do, draw a card.",
		"Instant")
	for _, tc := range []struct {
		name              string
		creature, accept  bool
		unpayable         bool
		wantDraw, counter bool
	}{
		{name: "paid", creature: true, accept: true, wantDraw: true},
		{name: "declined", creature: true, counter: true},
		{name: "unpayable", creature: true, accept: true, unpayable: true, counter: true},
		{name: "skipped", accept: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			if tc.creature {
				addCreaturePermanent(g, game.Player1)
			}
			addCardToLibrary(g, game.Player1, vanillaCreature("Draw", 1, 1))
			if !tc.unpayable {
				g.Players[game.Player2].ManaPool.Add(mana.C, 3)
			}
			target := addStackSpell(g, game.Player2, "Target", []types.Card{types.Sorcery})
			obj := pushCounterTaxSpell(g, def, target)
			obj.ResolutionResults = map[string]game.InstructionResolutionResult{
				"payment-clause-1": {Accepted: true, Succeeded: true},
			}
			agents := [game.NumPlayers]PlayerAgent{}
			if !tc.accept {
				agents[game.Player2] = &choiceOnlyAgent{choices: [][]int{{0}}}
			}
			log := TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &log)
			wantHand := 0
			if tc.wantDraw {
				wantHand = 1
			}
			if g.Players[game.Player1].Life != 41 || g.Players[game.Player1].Hand.Size() != wantHand {
				t.Fatalf("independent controller rider stole payer result: life=%d hand=%d", g.Players[game.Player1].Life, g.Players[game.Player1].Hand.Size())
			}
			if _, remains := stackObjectByID(g, target.ID); remains == tc.counter {
				t.Fatal("wrong counter disposition")
			}
			if !tc.creature && len(log.Choices) != 0 {
				t.Fatal("skipped payment offered a choice")
			}
			if result := obj.ResolutionResults["payment-clause-1"]; !result.Accepted || !result.Succeeded ||
				len(obj.ResolutionResults) != 1 {
				t.Fatal("local payment result did not restore outer frame")
			}
		})
	}
}
