package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: pinned Oracle corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4;
// Oracle ID 1cea6439-7ae5-4887-8c33-7da9fb36e2d4, printing
// 985bdb0c-ce6c-4506-8163-76f3b2fdf5fb (MH2 #49), {1}{U}, Instant. Live
// Scryfall endpoints retrieved 2026-11-19:
// https://api.scryfall.com/cards/985bdb0c-ce6c-4506-8163-76f3b2fdf5fb and
// https://api.scryfall.com/cards/985bdb0c-ce6c-4506-8163-76f3b2fdf5fb/rulings.
// The card endpoint confirms the pinned Oracle text; the rulings endpoint has
// two WotC rulings dated 2021-06-18: Replicate copies still happen if the
// original left the stack, and those copies are not cast.
//
// Read the official Comprehensive Rules, effective November 6, 2026:
// https://media.wizards.com/2026/downloads/MagicCompRules%2020261106.txt
// Sections 601.2b/c/f-i cover announcing additional costs and targets, paying
// the total cost, and the cast event; 603.2-3 cover the Replicate trigger;
// 608.2b/g/h cover target legality, an opponent's mana-payment choice, and
// information used by a resolving effect; 702.56a-b define Replicate as a
// repeatable additional cost with triggers per paid instance; 707.10/.10c
// establish that copies aren't cast and may receive new legal targets.
func TestLoseFocusReplicateScenario(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	target := addStackSpell(g, game.Player2, "Target spell", []types.Card{types.Instant})
	card := addCardToHand(g, game.Player1, authored.LoseFocus.New())
	g.Players[game.Player1].ManaPool.Add(mana.C, 1)
	g.Players[game.Player1].ManaPool.Add(mana.U, 2)
	g.Players[game.Player2].ManaPool.Add(mana.C, 2)
	setMainPhasePriority(g, game.Player1)

	cast := action.AnnounceSpell(action.CastSpellAction{
		CardID: card, SourceZone: zone.Hand,
		Targets:         []game.Target{game.StackObjectTarget(target.ID)},
		ReplicateCounts: []int{1},
	})
	if !engine.applyAction(g, game.Player1, cast) {
		t.Fatal("Lose Focus could not be cast with one Replicate payment")
	}
	original, ok := g.Stack.Peek()
	if !ok || original.ManaSpentToCast != 3 || len(original.ReplicateCounts) != 1 || original.ReplicateCounts[0] != 1 {
		t.Fatalf("cast did not pay {1}{U}{U} and one Replicate instance: %+v", original)
	}
	if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 3 {
		t.Fatal("one paid Replicate instance did not trigger above the target and original")
	}

	agents := conspireTestAgent(func(request game.ChoiceRequest) []int {
		if request.Prompt == "Choose new targets" {
			for _, option := range request.Options {
				if option.Label == "Keep current targets" {
					return []int{option.Index}
				}
			}
			t.Fatal("Replicate copy did not offer to keep its current target")
		}
		return request.DefaultSelection
	})
	agents[game.Player2] = optionalMayAgent{accept: true}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	if g.Stack.Size() != 3 {
		t.Fatal("Replicate trigger did not create exactly one copy")
	}

	engine.resolveTopOfStackWithChoices(g, agents, &log)
	if _, ok := stackObjectByID(g, target.ID); !ok || g.Players[game.Player2].ManaPool.Total() != 0 {
		t.Fatal("target controller did not pay {2} for the first resolution")
	}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	if _, ok := stackObjectByID(g, target.ID); ok ||
		!g.Players[game.Player2].Graveyard.Contains(target.SourceID) {
		t.Fatal("Lose Focus did not counter the target after its controller could no longer pay")
	}
	if !g.Players[game.Player1].Graveyard.Contains(card) {
		t.Fatal("resolved Lose Focus did not leave the stack for its owner's graveyard")
	}

	var paymentChoices int
	for _, choice := range log.Choices {
		if choice.Request.Player == game.Player2 {
			if choice.Request.Kind != game.ChoiceMay || choice.Request.Prompt != "Pay {2}?" {
				t.Fatalf("target controller received a nonpayment choice: %+v", choice.Request)
			}
			paymentChoices++
		}
	}
	if paymentChoices != 1 {
		t.Fatalf("target controller received %d choices, want one payable {2} offer", paymentChoices)
	}
}
