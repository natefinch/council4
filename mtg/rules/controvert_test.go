package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Coldsnap #30 printing revision 0e670f6b-d16e-47fc-a5b7-7ca0d8763644,
// Oracle ID f7ad79e3-d128-4aee-8542-99fd3527f9d3; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// current Scryfall card response retrieved 2026-11-20 at
// https://api.scryfall.com/cards/0e670f6b-d16e-47fc-a5b7-7ca0d8763644 confirms
// "Counter target spell.\nRecover {2}{U}{U} (When a creature is put into your
// graveyard from the battlefield, you may pay {2}{U}{U}. If you do, return this
// card from your graveyard to your hand. Otherwise, exile this card.)" The
// printing's rulings URI
// https://api.scryfall.com/cards/0e670f6b-d16e-47fc-a5b7-7ca0d8763644/rulings
// returned an empty list on 2026-11-20 (raw responses in
// .cardwork/sources/controvert/).
//
// Wizards of the Coast Comprehensive Rules effective 2026-09-25 (the pinned
// rules_edition in testdata/card-preservation/corpus.json),
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt;
// verified local copy .cardwork/sources/controvert/MagicCompRules-20260925.txt,
// SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
// Read §§ 115.1a, 601.2b-c/e-i, 608.2b-c and 701.6a-b for casting the
// targeted counterspell, rechecking its target and putting a countered spell in
// its owner's graveyard; §§ 702.59a, 700.4 and 603.1-3 for Recover's
// graveyard-only creature-death trigger; §§ 118.12 and 608.2g for its optional
// resolution payment; §§ 400.7 and 404.1-2 for card identity and zone movement.
const controvertTestOracleID = "f7ad79e3-d128-4aee-8542-99fd3527f9d3"

type controvertRecoverAgent struct {
	simPassPolicy
	pay      bool
	requests []game.ChoiceRequest
}

func (a *controvertRecoverAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		if a.pay {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

func TestControvertCountersSpellAndRecovers(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(controvertTestOracleID)
	if err != nil {
		t.Fatalf("exact Controvert registry lookup failed: %v", err)
	}
	if def.Name != "Controvert" {
		t.Fatalf("exact Controvert lookup returned %q", def.Name)
	}

	for _, test := range []struct {
		name      string
		pay       bool
		wantHand  bool
		wantExile bool
		wantMana  int
	}{
		{name: "pays Recover", pay: true, wantHand: true},
		{name: "declines Recover and exiles", wantExile: true, wantMana: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := removalTokenEngine(t, authored.Controvert)
			target := productionCounterSubject(g, false)
			controvert := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.U, 8)
			setMainPhasePriority(g, game.Player1)

			if !engine.applyAction(g, game.Player1, action.CastSpell(
				controvert, []game.Target{game.StackObjectTarget(target.ID)}, 0, nil,
			)) {
				t.Fatal("casting Controvert targeting a spell was rejected")
			}
			spell, ok := g.Stack.Peek()
			if !ok || spell.ManaSpentToCast != 4 || g.Players[game.Player1].ManaPool.Total() != 4 {
				t.Fatalf("Controvert did not pay {2}{U}{U}: spell=%+v mana=%d", spell, g.Players[game.Player1].ManaPool.Total())
			}

			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Graveyard.Contains(target.SourceID) ||
				!g.Players[game.Player1].Graveyard.Contains(controvert) {
				t.Fatal("Controvert did not counter the spell and move itself to its owner's graveyard")
			}

			dying := addCombatCreaturePermanent(g, game.Player1)
			if _, destroyed := destroyPermanent(g, dying.ObjectID); !destroyed {
				t.Fatal("expected a creature to die after Controvert entered its owner's graveyard")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("Controvert's Recover ability did not trigger for its owner's creature")
			}
			trigger, ok := g.Stack.Peek()
			if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.SourceCardID != controvert ||
				trigger.SourceZone != zone.Graveyard || trigger.TriggerEvent.PermanentID != dying.ObjectID {
				t.Fatalf("Recover trigger did not retain the Controvert source and death event: %+v", trigger)
			}

			agent := &controvertRecoverAgent{pay: test.pay}
			agents := simPassPolicies()
			agents[game.Player1] = agent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(controvert) != test.wantHand ||
				g.Players[game.Player1].Graveyard.Contains(controvert) ||
				g.Players[game.Player1].Exile.Contains(controvert) != test.wantExile {
				t.Fatalf("Recover result: hand=%v graveyard=%v exile=%v",
					g.Players[game.Player1].Hand.Contains(controvert),
					g.Players[game.Player1].Graveyard.Contains(controvert),
					g.Players[game.Player1].Exile.Contains(controvert))
			}
			if len(agent.requests) != 1 || agent.requests[0].Kind != game.ChoiceMay ||
				agent.requests[0].Player != game.Player1 {
				t.Fatalf("Recover did not offer its optional payment to Controvert's owner: %+v", agent.requests)
			}
			if g.Players[game.Player1].ManaPool.Total() != test.wantMana {
				t.Fatalf("Controvert's spell and Recover costs left %d mana, want %d",
					g.Players[game.Player1].ManaPool.Total(), test.wantMana)
			}
		})
	}
}
