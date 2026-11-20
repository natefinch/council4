package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Krovikan Rot printing 17597c66-0d9f-41af-9160-0d92be88f450
// (CSP 63, Oracle ID f9a272c5-6ebc-4d29-a47c-a2d570184196); pinned and live
// Scryfall text agree; https://api.scryfall.com/cards/17597c66-0d9f-41af-9160-0d92be88f450/rulings
// returned no entries on 2026-11-20.
// Read the official CR 2026-09-25 at
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.pdf:
// 115.1a (target chosen while casting), 701.8a (destroy to owner's graveyard),
// 702.59a (Recover only in a graveyard; own creature death; pay/return or exile),
// 603.2-603.3 (triggering and stack placement), and 118.12 (optional resolution
// cost).
const krovikanRotTestOracleID = "f9a272c5-6ebc-4d29-a47c-a2d570184196"

type krovikanRotRecoverAgent struct {
	simPassPolicy
	pay      bool
	requests []game.ChoiceRequest
}

func (a *krovikanRotRecoverAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		if a.pay {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

func TestKrovikanRotDestroysSmallCreatureAndRecovers(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(krovikanRotTestOracleID)
	if err != nil {
		t.Fatalf("exact Krovikan Rot registry lookup failed: %v", err)
	}
	if def.Name != "Krovikan Rot" {
		t.Fatalf("exact Krovikan Rot lookup returned %q", def.Name)
	}

	for _, test := range []struct {
		name      string
		pay       bool
		wantHand  bool
		wantExile bool
		wantMana  int
	}{
		{name: "destroys a small creature and pays Recover", pay: true, wantHand: true},
		{name: "destroys a small creature and declines Recover", wantExile: true, wantMana: 3},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			target := addCombatCreaturePermanentWithPower(g, game.Player2, 2)
			tooPowerful := addCombatCreaturePermanentWithPower(g, game.Player2, 3)
			rot := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.B, 6)
			setMainPhasePriority(g, game.Player1)

			if engine.applyAction(g, game.Player1, action.CastSpell(
				rot, []game.Target{game.PermanentTarget(tooPowerful.ObjectID)}, 0, nil,
			)) {
				t.Fatal("Krovikan Rot accepted a creature with power 3")
			}
			if g.Stack.Size() != 0 || g.Players[game.Player1].ManaPool.Total() != 6 {
				t.Fatal("rejecting an illegal target changed the stack or mana pool")
			}

			if !engine.applyAction(g, game.Player1, action.CastSpell(
				rot, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
			)) {
				t.Fatal("casting Krovikan Rot targeting an opponent's 2/2 creature was rejected")
			}
			if spell, ok := g.Stack.Peek(); !ok || spell.ManaSpentToCast != 3 ||
				g.Players[game.Player1].ManaPool.Total() != 3 {
				t.Fatalf("Krovikan Rot did not pay {2}{B}: spell=%+v mana=%d",
					spell, g.Players[game.Player1].ManaPool.Total())
			}

			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
				!g.Players[game.Player1].Graveyard.Contains(rot) {
				t.Fatal("Krovikan Rot did not destroy its target and finish in its owner's graveyard")
			}

			dying := addCombatCreaturePermanentWithPower(g, game.Player1, 2)
			if _, destroyed := destroyPermanent(g, dying.ObjectID); !destroyed {
				t.Fatal("expected a creature to die into Krovikan Rot's owner's graveyard")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("Krovikan Rot's Recover ability did not trigger for its owner's creature")
			}
			trigger, ok := g.Stack.Peek()
			if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.SourceCardID != rot ||
				trigger.SourceZone != zone.Graveyard || trigger.TriggerEvent.PermanentID != dying.ObjectID {
				t.Fatalf("Recover trigger did not retain the Krovikan Rot source and death event: %+v", trigger)
			}

			agent := &krovikanRotRecoverAgent{pay: test.pay}
			agents := simPassPolicies()
			agents[game.Player1] = agent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(rot) != test.wantHand ||
				g.Players[game.Player1].Graveyard.Contains(rot) ||
				g.Players[game.Player1].Exile.Contains(rot) != test.wantExile {
				t.Fatalf("Recover result: hand=%v graveyard=%v exile=%v",
					g.Players[game.Player1].Hand.Contains(rot),
					g.Players[game.Player1].Graveyard.Contains(rot),
					g.Players[game.Player1].Exile.Contains(rot))
			}
			if len(agent.requests) != 1 || agent.requests[0].Kind != game.ChoiceMay {
				t.Fatalf("Recover did not offer its optional {1}{B}{B} payment: %+v", agent.requests)
			}
			if g.Players[game.Player1].ManaPool.Total() != test.wantMana {
				t.Fatalf("Recover mana remaining = %d, want %d",
					g.Players[game.Player1].ManaPool.Total(), test.wantMana)
			}
		})
	}
}
