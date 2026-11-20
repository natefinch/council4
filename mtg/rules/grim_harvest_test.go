package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Source review: Coldsnap #58 printing revision bdb1972f-2ef7-4fe2-8c8e-ab07f48a3176,
// Oracle ID 231ca538-b277-4c12-979f-57640a0e87f0; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. Current
// Scryfall card response on 2026-11-20 confirmed this exact Oracle text:
// "Return target creature card from your graveyard to your hand.\nRecover {2}{B}
// (When a creature is put into your graveyard from the battlefield, you may pay
// {2}{B}. If you do, return this card from your graveyard to your hand.
// Otherwise, exile this card.)" Its rulings URI
// https://api.scryfall.com/cards/bdb1972f-2ef7-4fe2-8c8e-ab07f48a3176/rulings
// returned an empty list on 2026-11-20 (raw responses retained in
// .cardwork/sources/recover-family/).
//
// Wizards of the Coast Comprehensive Rules effective 2026-09-25 (the corpus
// rules pin), https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt;
// verified local copy .cardwork/sources/recover-family/MagicCompRules-20260925.txt,
// SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca.
// Read §§ 601.2b-c/e-h/i and 608.2b for the spell's target, payment and
// resolution; §§ 702.59a and 700.4 define Recover and dying; §§ 603.1-3,
// 603.6c and 603.10a require the existing graveyard ability to trigger for
// creature deaths and let each event be processed; §§ 118.12 and 608.2g make
// the optional mana payment a resolution choice; §§ 400.3/.7 and 113.7a/.8
// require owner-zone movement and retain the independent trigger's controller
// and source relationship.
const grimHarvestTestOracleID = "231ca538-b277-4c12-979f-57640a0e87f0"

type grimHarvestPayAgent struct {
	simPassPolicy
	pay      bool
	requests []game.ChoiceRequest
}

func (a *grimHarvestPayAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		if a.pay {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

func TestGrimHarvestSpellAndRecoverChoices(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(grimHarvestTestOracleID)
	if err != nil {
		t.Fatalf("exact Grim Harvest registry lookup failed: %v", err)
	}
	if def.Name != "Grim Harvest" {
		t.Fatalf("exact Grim Harvest lookup returned %q", def.Name)
	}

	for _, test := range []struct {
		name      string
		pay       bool
		wantHand  bool
		wantExile bool
		wantMana  int
	}{
		{name: "pays Recover", pay: true, wantHand: true, wantMana: 1},
		{name: "declines Recover and exiles", wantExile: true, wantMana: 4},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			returnedCreature := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
				Name: "Grim Harvest target", Types: []types.Card{types.Creature},
			}})
			harvest := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.C, 3)
			g.Players[game.Player1].ManaPool.Add(mana.B, 3)
			setMainPhasePriority(g, game.Player1)

			if !engine.applyAction(g, game.Player1, action.CastSpell(
				harvest, []game.Target{currentCardTarget(t, g, returnedCreature)}, 0, nil,
			)) {
				t.Fatal("casting Grim Harvest targeting its owner's graveyard creature was rejected")
			}
			spell, ok := g.Stack.Peek()
			if !ok || spell.ManaSpentToCast != 2 || g.Players[game.Player1].ManaPool.Total() != 4 {
				t.Fatalf("Grim Harvest did not pay {1}{B}: spell=%+v mana=%d", spell, g.Players[game.Player1].ManaPool.Total())
			}
			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player1].Hand.Contains(returnedCreature) ||
				!g.Players[game.Player1].Graveyard.Contains(harvest) {
				t.Fatal("spell did not return the targeted creature and move itself to its owner's graveyard")
			}

			dying := addCombatCreaturePermanent(g, game.Player1)
			if _, destroyed := destroyPermanent(g, dying.ObjectID); !destroyed {
				t.Fatal("expected a creature to die after Grim Harvest entered its owner's graveyard")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("Grim Harvest's Recover ability did not trigger for its owner's creature")
			}
			trigger, ok := g.Stack.Peek()
			if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.SourceCardID != harvest ||
				trigger.SourceZone != zone.Graveyard || trigger.TriggerEvent.PermanentID != dying.ObjectID {
				t.Fatalf("Recover trigger did not retain the Grim Harvest source and death event: %+v", trigger)
			}

			agent := &grimHarvestPayAgent{pay: test.pay}
			agents := simPassPolicies()
			agents[game.Player1] = agent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(harvest) != test.wantHand ||
				g.Players[game.Player1].Graveyard.Contains(harvest) ||
				g.Players[game.Player1].Exile.Contains(harvest) != test.wantExile {
				t.Fatalf("Recover result: hand=%v graveyard=%v exile=%v",
					g.Players[game.Player1].Hand.Contains(harvest),
					g.Players[game.Player1].Graveyard.Contains(harvest),
					g.Players[game.Player1].Exile.Contains(harvest))
			}
			if len(agent.requests) != 1 || agent.requests[0].Kind != game.ChoiceMay {
				t.Fatalf("Recover did not offer its optional payment: %+v", agent.requests)
			}
			if g.Players[game.Player1].ManaPool.Total() != test.wantMana {
				t.Fatalf("Grim Harvest's spell and Recover costs left %d mana, want %d",
					g.Players[game.Player1].ManaPool.Total(), test.wantMana)
			}
		})
	}
}
