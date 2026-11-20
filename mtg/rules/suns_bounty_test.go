package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Coldsnap #18, printing revision bde59b0e-c75e-455b-b375-e0dac430a6d1,
// Oracle ID 5c15d9b1-5ec7-4476-a3e6-539d899a9957; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// current Scryfall card response retrieved 2026-11-20 from
// https://api.scryfall.com/cards/bde59b0e-c75e-455b-b375-e0dac430a6d1 matches:
// "{1}{W}, Instant. You gain 4 life.\nRecover {1}{W} (When a creature is put
// into your graveyard from the battlefield, you may pay {1}{W}. If you do,
// return this card from your graveyard to your hand. Otherwise, exile this
// card.)" Its rulings URI
// https://api.scryfall.com/cards/bde59b0e-c75e-455b-b375-e0dac430a6d1/rulings
// returned an empty list that day; raw responses are retained in
// .cardwork/sources/suns-bounty/.
//
// Wizards Comprehensive Rules effective 2026-09-25 (the corpus rules pin),
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt;
// verified local copy .cardwork/sources/suns-bounty/MagicCompRules-20260925.txt,
// SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
// Read §§ 601.2, 608.2, 119.3, 702.59a, 700.4, 603.1-3, 603.6c, 603.10a,
// 118.12, 400.7, and 404.1. These define paying and resolving the spell, gaining
// life, Recover's graveyard-only dies trigger, its optional resolution payment,
// and the source card's graveyard-to-hand or exile movement.
const sunsBountyTestOracleID = "5c15d9b1-5ec7-4476-a3e6-539d899a9957"

type sunsBountyRecoverAgent struct {
	simPassPolicy
	pay      bool
	requests []game.ChoiceRequest
}

func (a *sunsBountyRecoverAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		if a.pay {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

func TestSunsBountySpellAndRecoverChoices(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(sunsBountyTestOracleID)
	if err != nil {
		t.Fatalf("exact Sun's Bounty registry lookup failed: %v", err)
	}
	if def.Name != "Sun's Bounty" {
		t.Fatalf("exact Sun's Bounty lookup returned %q", def.Name)
	}

	for _, test := range []struct {
		name      string
		pay       bool
		wantHand  bool
		wantExile bool
		wantMana  int
	}{
		{name: "pays Recover", pay: true, wantHand: true},
		{name: "declines Recover and exiles", wantExile: true, wantMana: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			sunsBounty := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.W, 4)
			setMainPhasePriority(g, game.Player1)

			startingLife := g.Players[game.Player1].Life
			if !engine.applyAction(g, game.Player1, action.CastSpell(sunsBounty, nil, 0, nil)) {
				t.Fatal("casting Sun's Bounty was rejected")
			}
			if spell, ok := g.Stack.Peek(); !ok || spell.ManaSpentToCast != 2 ||
				g.Players[game.Player1].ManaPool.Total() != 2 {
				t.Fatalf("Sun's Bounty did not pay {1}{W}: spell=%+v mana=%d",
					spell, g.Players[game.Player1].ManaPool.Total())
			}

			engine.resolveTopOfStack(g, &TurnLog{})
			if got, want := g.Players[game.Player1].Life, startingLife+4; got != want {
				t.Fatalf("life after Sun's Bounty = %d, want %d", got, want)
			}
			if !g.Players[game.Player1].Graveyard.Contains(sunsBounty) {
				t.Fatal("resolved Sun's Bounty did not reach its owner's graveyard")
			}

			dying := addCombatCreaturePermanent(g, game.Player1)
			if _, destroyed := destroyPermanent(g, dying.ObjectID); !destroyed {
				t.Fatal("expected a creature to die after Sun's Bounty entered its owner's graveyard")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("Sun's Bounty's Recover ability did not trigger for its owner's creature")
			}
			trigger, ok := g.Stack.Peek()
			if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.SourceCardID != sunsBounty ||
				trigger.SourceZone != zone.Graveyard || trigger.TriggerEvent.PermanentID != dying.ObjectID {
				t.Fatalf("Recover trigger did not retain the Sun's Bounty source and death event: %+v", trigger)
			}

			agent := &sunsBountyRecoverAgent{pay: test.pay}
			agents := simPassPolicies()
			agents[game.Player1] = agent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(sunsBounty) != test.wantHand ||
				g.Players[game.Player1].Graveyard.Contains(sunsBounty) ||
				g.Players[game.Player1].Exile.Contains(sunsBounty) != test.wantExile {
				t.Fatalf("Recover result: hand=%v graveyard=%v exile=%v",
					g.Players[game.Player1].Hand.Contains(sunsBounty),
					g.Players[game.Player1].Graveyard.Contains(sunsBounty),
					g.Players[game.Player1].Exile.Contains(sunsBounty))
			}
			if len(agent.requests) != 1 || agent.requests[0].Kind != game.ChoiceMay {
				t.Fatalf("Recover did not offer its optional {1}{W} payment: %+v", agent.requests)
			}
			if got := g.Players[game.Player1].ManaPool.Total(); got != test.wantMana {
				t.Fatalf("Sun's Bounty's spell and Recover costs left %d mana, want %d", got, test.wantMana)
			}
		})
	}
}
