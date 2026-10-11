package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestResultObjectTemptingBaseAvailabilityIndependentOfOpponentAcceptance(t *testing.T) {
	for _, tc := range []struct {
		name   string
		count  int
		hidden bool
	}{
		{"known empty base", 0, false},
		{"public base", 1, false},
		{"hidden base", 1, true},
		{"mixed base", 2, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for i := range tc.count {
				cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Base member", 1, 1))
				if tc.hidden && i == 0 {
					redirectResultCardToHand(g, cardID)
				}
			}
			obj := &game.StackObject{Controller: game.Player1}
			publisher := game.Instruction{
				Primitive: game.MoveTopOfLibrary{Player: game.GroupOfferMemberReference(),
					Amount: game.Fixed(2), Destination: zone.Graveyard},
				PublishResult: "tempt", Optional: true, TemptingOffer: true,
				OptionalActorGroup: opt.Val(game.OpponentsReference()),
			}
			agents := [game.NumPlayers]PlayerAgent{
				game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
				game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
				game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
			}
			engine := NewEngine(nil)
			engine.resolveInstructionWithChoices(g, obj, &publisher, agents, &TurnLog{})
			snapshots, present := obj.ResolutionResultObjects["tempt"]
			if present == tc.hidden || !tc.hidden && len(snapshots) != tc.count {
				t.Fatal("offer-acceptance erased a complete base or certified its incomplete mixed batch")
			}
			receipt := obj.ResolutionResults["tempt"]
			if receipt.Accepted || receipt.Succeeded != (tc.count > 0) ||
				receipt.Amount != 0 || receipt.AcceptedActors.Count() != 0 {
				t.Fatalf("snapshot provenance changed the offer-acceptance receipt: %#v", receipt)
			}
			gate := game.InstructionResultGate{Key: "tempt", Succeeded: game.TriTrue,
				ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}})}
			if got := instructionResultGateSatisfied(g, obj, gate); got != (!tc.hidden && tc.count > 0) {
				t.Fatal("filtered base result lost its independent known availability")
			}
			gate.Negate = true
			if got := instructionResultGateSatisfied(g, obj, gate); got != (!tc.hidden && tc.count == 0) {
				t.Fatal("Accepted=false inverted unknown base members into a known-empty result")
			}
			reader := actualCardQuantityReader("tempt", 1)
			reader.PublishResult = "consumer"
			engine.resolveInstructionWithChoices(g, obj, &reader, agents, &TurnLog{})
			if _, executed := obj.ResolutionResults["consumer"]; executed {
				t.Fatal("offer's opponent acceptance was substituted for actual base-quantity availability")
			}
		})
	}
}

func TestResultObjectNilTemptingBodyCannotMintAvailableEmpty(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	obj := &game.StackObject{Controller: game.Player1}
	publisher := temptingOfferCompoundInstruction()
	publisher.PublishResult = "body"
	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player3: &choiceOnlyAgent{choices: [][]int{{0}}},
		game.Player4: &choiceOnlyAgent{choices: [][]int{{0}}},
	}
	NewEngine(nil).resolveInstructionWithChoices(g, obj, &publisher, agents, &TurnLog{})
	if _, present := obj.ResolutionResultObjects["body"]; present {
		t.Fatal("nil multi-body envelope minted a result-object provider's known-empty batch")
	}
	if receipt, exists := obj.ResolutionResults["body"]; !exists || receipt.Accepted {
		t.Fatal("nil multi-body envelope lost its independent opponent-acceptance receipt")
	}
}
