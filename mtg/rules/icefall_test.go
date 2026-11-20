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

// Source review: Coldsnap #85 printing revision 6cf90fc3-b08b-4261-a194-d6b06fdd59d8,
// Oracle ID 30baca42-50e2-45f8-8f99-8e2f9795b4da; pinned corpus SHA-256
// 6da9ced0e19bd384771ab8f14051cc45a8dc90302082b875fa6e039d2399a1b4. The
// current Scryfall card response fetched 2026-11-20 from
// https://api.scryfall.com/cards/6cf90fc3-b08b-4261-a194-d6b06fdd59d8 matches
// its complete Oracle text. Its rulings URI
// https://api.scryfall.com/cards/6cf90fc3-b08b-4261-a194-d6b06fdd59d8/rulings
// returned an empty list. Raw responses are in .cardwork/sources/recover-family/.
//
// Wizards Comprehensive Rules effective 2026-09-25, the pinned rules edition,
// https://media.wizards.com/2026/downloads/MagicCompRules%2020260925.txt;
// verified local copy .cardwork/sources/recover-family/MagicCompRules-20260925.txt
// (SHA-256 8d860e451f20f38865b725b42d82feb714c725373dd8f3b32b8652b3eeb070ca).
// Read §§ 115.1a, 601.2c/f-h and 608.2b-c for Icefall's artifact-or-land target,
// cast cost and target recheck; §§ 701.8a-b and 404.1 for destruction and the
// spell's graveyard destination; §§ 702.59a, 700.4, 603.1-3, 603.6c and 603.10a
// for Recover's graveyard-only trigger and creature death; §§ 118.12, 400.7 and
// 113.7a for the optional resolution payment and source movement/identity.
const icefallTestOracleID = "30baca42-50e2-45f8-8f99-8e2f9795b4da"

type icefallRecoverAgent struct {
	simPassPolicy
	pay      bool
	requests []game.ChoiceRequest
}

func (a *icefallRecoverAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		if a.pay {
			return []int{1}
		}
		return []int{0}
	}
	return request.DefaultSelection
}

func TestIcefallDestroyTargetAndRecoverChoices(t *testing.T) {
	def, err := cards.NewDefaultRegistry().LookupOracleID(icefallTestOracleID)
	if err != nil {
		t.Fatalf("exact Icefall registry lookup failed: %v", err)
	}
	if def.Name != "Icefall" {
		t.Fatalf("exact Icefall lookup returned %q", def.Name)
	}

	for _, test := range []struct {
		name       string
		targetType string
		pay        bool
		wantHand   bool
		wantExile  bool
		wantMana   int
	}{
		{name: "destroys artifact and pays Recover", targetType: "artifact", pay: true, wantHand: true},
		{name: "destroys land and declines Recover", targetType: "land", wantExile: true, wantMana: 2},
	} {
		t.Run(test.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			engine := NewEngine(nil)
			var target *game.Permanent
			if test.targetType == "land" {
				target = addBasicLandPermanent(g, game.Player2, types.Mountain)
			} else {
				target = addArtifactPermanent(g, game.Player2)
			}
			icefall := addCardToHand(g, game.Player1, def)
			g.Players[game.Player1].ManaPool.Add(mana.R, 6)
			setMainPhasePriority(g, game.Player1)

			if !engine.applyAction(g, game.Player1, action.CastSpell(
				icefall, []game.Target{game.PermanentTarget(target.ObjectID)}, 0, nil,
			)) {
				t.Fatalf("casting Icefall targeting an opponent's %s was rejected", test.targetType)
			}
			if spell, ok := g.Stack.Peek(); !ok || spell.ManaSpentToCast != 4 ||
				g.Players[game.Player1].ManaPool.Total() != 2 {
				t.Fatalf("Icefall did not pay {2}{R}{R}: spell=%+v mana=%d", spell, g.Players[game.Player1].ManaPool.Total())
			}

			engine.resolveTopOfStack(g, &TurnLog{})
			if !g.Players[game.Player2].Graveyard.Contains(target.CardInstanceID) ||
				!g.Players[game.Player1].Graveyard.Contains(icefall) {
				t.Fatal("Icefall did not destroy its target and finish in its owner's graveyard")
			}

			dying := addCombatCreaturePermanent(g, game.Player1)
			if _, destroyed := destroyPermanent(g, dying.ObjectID); !destroyed {
				t.Fatal("expected a creature to die after Icefall entered its owner's graveyard")
			}
			if !engine.putTriggeredAbilitiesOnStack(g) || g.Stack.Size() != 1 {
				t.Fatal("Icefall's Recover ability did not trigger for its owner's creature")
			}
			trigger, ok := g.Stack.Peek()
			if !ok || trigger.Kind != game.StackTriggeredAbility || trigger.SourceCardID != icefall ||
				trigger.SourceZone != zone.Graveyard || trigger.TriggerEvent.PermanentID != dying.ObjectID {
				t.Fatalf("Recover trigger did not retain the Icefall source and death event: %+v", trigger)
			}

			agent := &icefallRecoverAgent{pay: test.pay}
			agents := simPassPolicies()
			agents[game.Player1] = agent
			engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
			if g.Players[game.Player1].Hand.Contains(icefall) != test.wantHand ||
				g.Players[game.Player1].Graveyard.Contains(icefall) ||
				g.Players[game.Player1].Exile.Contains(icefall) != test.wantExile {
				t.Fatalf("Recover result: hand=%v graveyard=%v exile=%v",
					g.Players[game.Player1].Hand.Contains(icefall),
					g.Players[game.Player1].Graveyard.Contains(icefall),
					g.Players[game.Player1].Exile.Contains(icefall))
			}
			if len(agent.requests) != 1 || agent.requests[0].Kind != game.ChoiceMay {
				t.Fatalf("Recover did not offer its optional {R}{R} payment: %+v", agent.requests)
			}
			if g.Players[game.Player1].ManaPool.Total() != test.wantMana {
				t.Fatalf("Icefall's spell and Recover costs left %d mana, want %d",
					g.Players[game.Player1].ManaPool.Total(), test.wantMana)
			}
		})
	}
}
