package rules

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
)

// Source review: Rotting Rats HOP 39 printing
// 3895f0ce-ef97-4188-8b86-d0cc7067aacd (oracle 1f361ae9-98a2-40ff-a579-fa5b9f510df4),
// Oracle "When this creature enters, each player discards a card." from the
// pinned corpus; https://api.scryfall.com/cards/3895f0ce-ef97-4188-8b86-d0cc7067aacd/rulings
// returned one ruling dated 2009-02-01: "When Rotting Rats's "enters" ability
// resolves, first the player whose turn it is chooses a card to discard, then
// each other player in turn order does the same, then all cards are discarded at
// the same time. No one sees what the other players are discarding before
// deciding which card to discard." CR 101.4 (APNAP choices, then simultaneous
// actions), CR 101.4a (hidden-zone choices) and CR 701.9b (the affected player
// chooses the discarded card by default) from the rules effective September 25,
// 2026.

type groupDiscardChoice struct {
	player         game.PlayerID
	discardsBefore int
	options        []string
	observedHand   []string
}

type groupDiscardChooserAgent struct {
	g      *game.Game
	player game.PlayerID
	pick   string
	log    *[]groupDiscardChoice
}

func (*groupDiscardChooserAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *groupDiscardChooserAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	record := groupDiscardChoice{player: request.Player, discardsBefore: len(cleanupDiscardEvents(a.g))}
	for _, card := range obs.Hand() {
		record.observedHand = append(record.observedHand, card.Name)
	}
	selected := []int(nil)
	for _, option := range request.Options {
		if !option.Card.Exists {
			continue
		}
		record.options = append(record.options, option.Card.Val.Name)
		if option.Card.Val.Name == a.pick {
			selected = []int{option.Index}
		}
	}
	*a.log = append(*a.log, record)
	return selected
}

// TestRottingRatsEachPlayerChoosesDiscardInAPNAPOrder proves the production
// Rotting Rats trigger lets every player choose the card they discard from their
// own hand (CR 701.9b), asks in APNAP order starting with the active player, and
// discards nothing until every player has chosen (CR 101.4).
func TestRottingRatsEachPlayerChoosesDiscardInAPNAPOrder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	g.Turn.ActivePlayer = game.Player3
	players := []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4}
	var log []groupDiscardChoice
	var agents [game.NumPlayers]PlayerAgent
	for _, playerID := range players {
		for _, role := range []string{"Keep", "Toss", "Hold"} {
			addCardToHand(g, playerID, discardableCard(fmt.Sprintf("P%d %s", playerID, role)))
		}
		agents[playerID] = &groupDiscardChooserAgent{g: g, player: playerID, pick: fmt.Sprintf("P%d Toss", playerID), log: &log}
	}
	enterBattlefieldFromHand(t, engine, g, game.Player3, registryCardDef(t, "Rotting Rats"))
	if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, nil) || g.Stack.Size() != 1 {
		t.Fatalf("Rotting Rats trigger count = %d, want 1", g.Stack.Size())
	}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	wantOrder := []game.PlayerID{game.Player3, game.Player4, game.Player1, game.Player2}
	if len(log) != len(wantOrder) {
		t.Fatalf("choice requests = %d, want %d", len(log), len(wantOrder))
	}
	for i, record := range log {
		if record.player != wantOrder[i] {
			t.Fatalf("choice %d asked player %d, want %d (APNAP order)", i, record.player, wantOrder[i])
		}
		if record.discardsBefore != 0 {
			t.Fatalf("player %d chose after %d discards, want all choices before any discard", record.player, record.discardsBefore)
		}
		prefix := fmt.Sprintf("P%d ", record.player)
		if len(record.options) != 3 {
			t.Fatalf("player %d options = %v, want their three hand cards", record.player, record.options)
		}
		for _, name := range record.options {
			if !strings.HasPrefix(name, prefix) {
				t.Fatalf("player %d was offered %q from another player's hand", record.player, name)
			}
		}
		for _, name := range record.observedHand {
			if !strings.HasPrefix(name, prefix) {
				t.Fatalf("player %d observed %q from another player's hand", record.player, name)
			}
		}
	}

	discards := cleanupDiscardEvents(g)
	if len(discards) != len(players) {
		t.Fatalf("discard events = %d, want %d", len(discards), len(players))
	}
	for _, playerID := range players {
		hand := handCardNames(g, playerID)
		slices.Sort(hand)
		want := []string{fmt.Sprintf("P%d Hold", playerID), fmt.Sprintf("P%d Keep", playerID)}
		if !slices.Equal(hand, want) {
			t.Fatalf("player %d hand = %v, want %v", playerID, hand, want)
		}
		if !graveyardHasCardNamed(g, playerID, fmt.Sprintf("P%d Toss", playerID)) {
			t.Fatalf("player %d graveyard lacks their chosen discard", playerID)
		}
	}
	assertZoneChangeBeforeEachDiscard(t, g)
}

// TestGroupDiscardZeroCardsDiscardsNothing proves a player-group discard of
// zero cards neither prompts nor discards (it is not coerced to one card).
func TestGroupDiscardZeroCardsDiscardsNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	players := []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4}
	var log []groupDiscardChoice
	var agents [game.NumPlayers]PlayerAgent
	for _, playerID := range players {
		addCardToHand(g, playerID, discardableCard(fmt.Sprintf("P%d Card", playerID)))
		agents[playerID] = &groupDiscardChooserAgent{g: g, player: playerID, log: &log}
	}
	addEffectSpellToStack(g, game.Player1, game.Discard{
		Amount:      game.Fixed(0),
		PlayerGroup: game.AllPlayersReference(),
	}, nil)

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(log) != 0 {
		t.Fatalf("choice requests = %d, want none for a zero-card discard", len(log))
	}
	if got := len(cleanupDiscardEvents(g)); got != 0 {
		t.Fatalf("discard events = %d, want 0", got)
	}
	for _, playerID := range players {
		if got := g.Players[playerID].Hand.Size(); got != 1 {
			t.Fatalf("player %d hand size = %d, want 1", playerID, got)
		}
	}
}

// TestDiscardChoicesKeepUndiscardedHandCardsOutOfLog proves the chosen-discard
// request is private on both the player-group route (Rotting Rats) and the
// single-player route: each chooser still receives their own hand as options,
// but the returned game log redacts them, so cards that stay in hand are never
// revealed (CR 400.2, CR 701.9b).
func TestDiscardChoicesKeepUndiscardedHandCardsOutOfLog(t *testing.T) {
	setup := func(t *testing.T) (*game.Game, *Engine, [game.NumPlayers]PlayerAgent, *[]groupDiscardChoice) {
		t.Helper()
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		log := &[]groupDiscardChoice{}
		var agents [game.NumPlayers]PlayerAgent
		for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
			for _, role := range []string{"Keep", "Toss", "Hold"} {
				addCardToHand(g, playerID, discardableCard(fmt.Sprintf("P%d %s", playerID, role)))
			}
			agents[playerID] = &groupDiscardChooserAgent{g: g, player: playerID, pick: fmt.Sprintf("P%d Toss", playerID), log: log}
		}
		return g, engine, agents, log
	}
	for _, tc := range []struct {
		name    string
		players []game.PlayerID
		resolve func(t *testing.T, g *game.Game, engine *Engine, agents [game.NumPlayers]PlayerAgent, turnLog *TurnLog)
	}{
		{
			name:    "Rotting Rats group discard",
			players: []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4},
			resolve: func(t *testing.T, g *game.Game, engine *Engine, agents [game.NumPlayers]PlayerAgent, turnLog *TurnLog) {
				enterBattlefieldFromHand(t, engine, g, game.Player1, registryCardDef(t, "Rotting Rats"))
				if !engine.putTriggeredAbilitiesOnStackWithChoices(g, agents, turnLog) || g.Stack.Size() != 1 {
					t.Fatalf("Rotting Rats trigger count = %d, want 1", g.Stack.Size())
				}
				engine.resolveTopOfStackWithChoices(g, agents, turnLog)
			},
		},
		{
			name:    "single-player chosen discard",
			players: []game.PlayerID{game.Player1},
			resolve: func(_ *testing.T, g *game.Game, engine *Engine, agents [game.NumPlayers]PlayerAgent, turnLog *TurnLog) {
				addEffectSpellToStack(g, game.Player1, game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)}, nil)
				engine.resolveTopOfStackWithChoices(g, agents, turnLog)
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, engine, agents, log := setup(t)
			turnLog := &TurnLog{}
			tc.resolve(t, g, engine, agents, turnLog)

			if len(*log) != len(tc.players) {
				t.Fatalf("choice requests = %d, want %d", len(*log), len(tc.players))
			}
			for _, record := range *log {
				if len(record.options) != 3 {
					t.Fatalf("player %d agent options = %v, want their three hand cards", record.player, record.options)
				}
			}
			for _, playerID := range tc.players {
				if !graveyardHasCardNamed(g, playerID, fmt.Sprintf("P%d Toss", playerID)) {
					t.Fatalf("player %d graveyard lacks their chosen discard", playerID)
				}
			}
			discardChoices := 0
			for _, decision := range turnLog.Choices {
				if decision.Request.Kind != game.ChoiceResolution {
					continue
				}
				discardChoices++
				request := decision.Request
				if !request.Private || request.Prompt != "Private decision" {
					t.Fatalf("logged discard request is not private: %+v", request)
				}
				for _, option := range request.Options {
					if option.Card.Exists || option.Label != "Private option" {
						t.Fatalf("logged discard option retains hand information: %+v", option)
					}
				}
			}
			if discardChoices != len(tc.players) {
				t.Fatalf("logged discard choices = %d, want %d", discardChoices, len(tc.players))
			}
			data, err := json.Marshal(GameResult{Turns: []TurnLog{*turnLog}})
			if err != nil {
				t.Fatal(err)
			}
			for _, playerID := range []game.PlayerID{game.Player1, game.Player2, game.Player3, game.Player4} {
				for _, role := range []string{"Keep", "Hold"} {
					if name := fmt.Sprintf("P%d %s", playerID, role); strings.Contains(string(data), name) {
						t.Fatalf("returned game log reveals undiscarded hand card %q", name)
					}
				}
			}
		})
	}
}

func handCardNames(g *game.Game, playerID game.PlayerID) []string {
	var names []string
	for _, cardID := range g.Players[playerID].Hand.All() {
		if card, ok := g.GetCardInstance(cardID); ok {
			names = append(names, card.Def.Name)
		}
	}
	return names
}

func graveyardHasCardNamed(g *game.Game, playerID game.PlayerID, name string) bool {
	for _, cardID := range g.Players[playerID].Graveyard.All() {
		if card, ok := g.GetCardInstance(cardID); ok && card.Def.Name == name {
			return true
		}
	}
	return false
}

func assertZoneChangeBeforeEachDiscard(t *testing.T, g *game.Game) {
	t.Helper()
	for i, event := range g.Events {
		if event.Kind != game.EventCardDiscarded {
			continue
		}
		if event.SimultaneousID == 0 {
			t.Fatalf("discard of card %d has no batch", event.CardID)
		}
		moved := false
		for _, earlier := range g.Events[:i] {
			if earlier.Kind == game.EventZoneChanged && earlier.CardID == event.CardID && earlier.SimultaneousID == event.SimultaneousID {
				moved = true
			}
		}
		if !moved {
			t.Fatalf("discard of card %d was not preceded by its zone change", event.CardID)
		}
	}
}
