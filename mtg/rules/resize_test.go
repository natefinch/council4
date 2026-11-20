package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Coldsnap #117 printing revision c5dded16-e516-430e-a528-c49a483f03b4,
// Oracle ID 2c092ed4-4dbb-41ec-9e07-e59b9676802e; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The current
// Scryfall card response retrieved 2026-11-20 confirms mana cost {1}{G}, Instant,
// and Oracle text: "Target creature gets +3/+3 until end of turn." followed by
// "Recover {1}{G} (When a creature is put into your graveyard from the battlefield,
// you may pay {1}{G}. If you do, return this card from your graveyard to your hand.
// Otherwise, exile this card.)" Its rulings URI
// https://api.scryfall.com/cards/c5dded16-e516-430e-a528-c49a483f03b4/rulings
// returned an empty list on 2026-11-20; raw card and ruling responses are retained
// in .cardwork/sources/resize/.
//
// Official Wizards Comprehensive Rules effective 2026-09-25 (the corpus rules pin),
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt;
// verified local copy .cardwork/sources/resize/MagicCompRules-20260925.txt,
// SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
// Read §§ 115.1a and 601.2a,c,f-h,i for target choice and casting payment;
// 608.2b-c for resolving the target effect; 611.2a/c, 613.4c and 514.2 for
// the until-end-of-turn +3/+3; 702.59a and 700.4 for Recover/dies; 603.1-3,
// 603.6c and 603.10a for zone-change trigger capture; 113.7a/.8, 118.12,
// 400.7 and 404.1-2 for the resolving trigger, optional payment, and graveyard
// object movement. The prior retained handoff's TestRecoverResizeTriggerGapProbe
// observed no trigger on baseline 74beb1ae5ec08efd3f6ef7a7c16223c4caf06a41.
const resizeTestOracleID = "2c092ed4-4dbb-41ec-9e07-e59b9676802e"

type resizeRecoverAgent struct {
	simPassPolicy
	pay      bool
	requests []game.ChoiceRequest
}

func (a *resizeRecoverAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		if a.pay {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

func TestResizeSpellAndRecover(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(resizeTestOracleID)
	if err != nil {
		t.Fatalf("exact Resize registry lookup failed: %v", err)
	}
	if def.Name != "Resize" {
		t.Fatalf("exact Resize lookup returned %q", def.Name)
	}

	for _, test := range []struct {
		name      string
		pay       bool
		wantHand  bool
		wantExile bool
		wantMana  int
	}{
		{name: "pays Recover", pay: true, wantHand: true, wantMana: 2},
		{name: "declines Recover and exiles", wantExile: true, wantMana: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			target := addCreatureWithPowerToughness(g, game.Player2, 1, 1)
			resize := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.C, 3)
			g.Players[game.Player1].ManaPool.Add(mana.G, 3)
			setMainPhasePriority(g, game.Player1)

			if !engine.applyAction(g, game.Player1, action.CastSpell(
				resize, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
			)) {
				t.Fatal("casting Resize targeting a creature was rejected")
			}
			spell, ok := g.Stack.Peek()
			if !ok || spell.ManaSpentToCast != 2 || g.Players[game.Player1].ManaPool.Total() != 4 {
				t.Fatalf("Resize did not pay {1}{G}: spell=%+v mana=%d", spell, g.Players[game.Player1].ManaPool.Total())
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if got := effectivePower(g, target); got != 4 {
				t.Fatalf("target power after Resize = %d, want 4", got)
			}
			if got, ok := effectiveToughness(g, target); !ok || got != 4 {
				t.Fatalf("target toughness after Resize = %d (available %t), want 4", got, ok)
			}
			if !g.Players[game.Player1].Graveyard.Contains(resize) {
				t.Fatal("resolved Resize did not finish in its owner's graveyard")
			}

			dying := addCombatCreaturePermanent(g, game.Player1)
			if _, destroyed := destroyPermanent(g, dying.ObjectID); !destroyed {
				t.Fatal("expected a creature to die in Resize's owner's graveyard")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("Resize's Recover ability did not trigger for its owner's creature")
			}
			trigger, ok := g.Stack.Peek()
			if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.SourceCardID != resize ||
				trigger.SourceZone != zone.Graveyard || trigger.TriggerEvent.PermanentID != dying.ObjectID {
				t.Fatalf("Recover trigger did not retain the Resize source and death event: %+v", trigger)
			}

			agent := &resizeRecoverAgent{pay: test.pay}
			agents := simPassPolicies()
			agents[game.Player1] = agent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(resize) != test.wantHand ||
				g.Players[game.Player1].Graveyard.Contains(resize) ||
				g.Players[game.Player1].Exile.Contains(resize) != test.wantExile {
				t.Fatalf("Recover result: hand=%v graveyard=%v exile=%v",
					g.Players[game.Player1].Hand.Contains(resize),
					g.Players[game.Player1].Graveyard.Contains(resize),
					g.Players[game.Player1].Exile.Contains(resize))
			}
			if len(agent.requests) != 1 || agent.requests[0].Kind != game.ChoiceMay {
				t.Fatalf("Recover did not offer its optional payment: %+v", agent.requests)
			}
			if g.Players[game.Player1].ManaPool.Total() != test.wantMana {
				t.Fatalf("Resize spell and Recover costs left %d mana, want %d",
					g.Players[game.Player1].ManaPool.Total(), test.wantMana)
			}
		})
	}
}
