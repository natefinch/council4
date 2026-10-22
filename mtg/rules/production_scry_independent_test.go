package rules

import (
	"encoding/json"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/zone"
)

// Independent production-path check against Myr Custodian's complete WotC
// 2023-02-04 ruling and verified CR 2026-09-25 101.4, 608.2c, 701.22a-d.
func TestProductionMyrCustodianPrivateStagedReplay(t *testing.T) {
	registry := cards.NewDefaultRegistry()
	def, err := registry.LookupOracleID("43923f24-7e95-4934-9a5c-8d9a5e06ddea")
	if err != nil || def == nil || def.Name != "Myr Custodian" {
		t.Fatalf("exact production identity: %v, %v", def, err)
	}
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.Turn.TurnNumber = 1
	var original [game.NumPlayers][]id.ID
	for player := range game.NumPlayers {
		for range 3 {
			addCardToLibrary(g, game.PlayerID(player), vanillaCreature("Integration scry secret", 2, 2))
		}
		original[player] = g.Players[player].Library.All()
	}
	engine := NewEngineWithPrograms(nil, registry.Programs())
	source := addCardToHand(g, game.Player2, def)
	setMainPhasePriority(g, game.Player2)
	g.Players[game.Player2].ManaPool.Add(mana.C, 3)
	if !engine.applyAction(g, game.Player2, action.CastSpell(source, nil, 0, nil)) {
		t.Fatal("real paid production Myr Custodian cast rejected")
	}
	engine.resolveTopOfStack(g, &TurnLog{})
	permanent := permanentByCardID(g, source)
	if permanent == nil || !engine.putTriggeredAbilitiesOnStackWithChoices(g, simPassPolicies(), &TurnLog{}) {
		t.Fatal("real entry failed to produce the triggered ability")
	}
	addEffectSpellToStack(g, game.Player4, game.MovePermanent{
		Object: game.TargetPermanentReference(0), Destination: zone.Graveyard,
	}, []game.Target{game.PermanentTarget(permanent.ObjectID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	if permanentByCardID(g, source) != nil {
		t.Fatal("source departure positive control failed")
	}
	g.PlayerControlEffects = []game.PlayerControlEffect{{
		ID: g.IDGen.Next(), Player: game.Player3, Controller: game.Player4, Turn: 1,
	}}
	replay := g.Clone()
	var offers, placements []game.PlayerID
	decider := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
		if request.Player == game.Player2 {
			if request.Kind != game.ChoiceScry || len(request.Options) != 2 ||
				!request.Options[0].Card.Exists || request.Options[0].Card.Val.CardID != original[game.Player2][0] {
				t.Fatal("ability controller did not receive its own private first scry")
			}
			return []int{0}
		}
		if !slices.Equal(g.Players[game.Player2].Library.All(), []id.ID{
			original[game.Player2][1], original[game.Player2][2], original[game.Player2][0],
		}) {
			t.Fatal("opponents began before the controller's scry completed")
		}
		for _, player := range []game.PlayerID{game.Player1, game.Player3, game.Player4} {
			if !slices.Equal(g.Players[player].Library.All(), original[player]) {
				t.Fatal("an opponent's library committed before all placement choices")
			}
		}
		if request.Player == game.Player3 && (obs.Player != game.Player3 || obs.DecisionMaker != game.Player4) {
			t.Fatal("controlled opponent's private choice went to the wrong decision maker")
		}
		if request.Kind == game.ChoiceMay {
			if len(placements) != 0 {
				t.Fatal("private placement began before all participation choices")
			}
			offers = append(offers, request.Player)
			if request.Player == game.Player4 {
				return []int{0}
			}
			return []int{1}
		}
		if request.Kind != game.ChoiceScry || !request.Subject.Exists ||
			request.Subject.Val.CardID != original[request.Player][0] ||
			!slices.Equal(offers, []game.PlayerID{game.Player3, game.Player4, game.Player1}) {
			t.Fatal("private placement did not follow complete APNAP participation")
		}
		if request.Player == game.Player1 {
			if len(request.ScryDecisions) != 3 || !request.ScryDecisions[0].PlacementChosen ||
				request.ScryDecisions[0].BottomCount != 1 || request.ScryDecisions[1].Accepted {
				t.Fatal("later player lost the earlier public placement and decline")
			}
		}
		placements = append(placements, request.Player)
		if request.Player == game.Player3 {
			return []int{1}
		}
		return []int{0}
	}}
	trap := &controlScript{choice: func(PlayerObservation, game.ChoiceRequest) []int {
		t.Fatal("controlled player was asked to make its own choice")
		return nil
	}}
	log := &TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{decider, decider, trap, decider}, log)
	if !slices.Equal(placements, []game.PlayerID{game.Player3, game.Player1}) || len(log.Choices) != 6 {
		t.Fatal("printed choices skipped, duplicated, or out of order")
	}
	if !slices.Equal(g.Players[game.Player3].Library.All(), []id.ID{
		original[game.Player3][1], original[game.Player3][2], original[game.Player3][0],
	}) || !slices.Equal(g.Players[game.Player4].Library.All(), original[game.Player4]) ||
		!slices.Equal(g.Players[game.Player1].Library.All(), original[game.Player1]) {
		t.Fatal("simultaneous placements changed the wrong cards or players")
	}
	var cohort id.ID
	count := 0
	for _, event := range g.Events {
		if event.Kind != game.EventScry {
			continue
		}
		count++
		if event.Player == game.Player2 {
			continue
		}
		if cohort == 0 {
			cohort = event.SimultaneousID
		}
		if cohort == 0 || event.SimultaneousID != cohort {
			t.Fatal("simultaneous opponents lack a shared event cohort")
		}
	}
	if count != 3 {
		t.Fatal("source-departed ability did not produce the three actual scries")
	}
	var decisions [game.NumPlayers][][]int
	for _, decision := range log.Choices {
		if decision.Request.Private {
			if decision.Request.Subject.Exists {
				t.Fatal("private subject escaped into returned choice log")
			}
			for _, option := range decision.Request.Options {
				if option.Card.Exists {
					t.Fatal("private option identity escaped into returned choice log")
				}
			}
		}
		decisions[decision.DecisionMaker] = append(decisions[decision.DecisionMaker], slices.Clone(decision.Selected))
	}
	data, err := json.Marshal(GameResult{Turns: []TurnLog{*log}})
	if err != nil || strings.Contains(string(data), "Integration scry secret") {
		t.Fatal("returned game log leaked a private library identity", err)
	}
	var agents [game.NumPlayers]PlayerAgent
	for player := range game.NumPlayers {
		agents[player] = &choiceOnlyAgent{choices: decisions[player]}
	}
	replayLog := &TurnLog{}
	engine.resolveTopOfStackWithChoices(replay, agents, replayLog)
	if !reflect.DeepEqual(log, replayLog) {
		t.Fatal("sanitized decision-maker-indexed replay changed the choices")
	}
	for player := range game.NumPlayers {
		if !slices.Equal(g.Players[player].Library.All(), replay.Players[player].Library.All()) {
			t.Fatal("replay did not reproduce the actual library order")
		}
	}
}
