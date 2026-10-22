package rules

import (
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func assertPrivateScryLog(t *testing.T, log *TurnLog) {
	t.Helper()
	check := func(decision game.ChoiceDecision) {
		t.Helper()
		request := decision.Request
		if request.Kind != game.ChoiceScry && request.Kind != game.ChoiceOrder {
			return
		}
		if !request.Private || request.Subject.Exists || request.Prompt != "Private decision" {
			t.Errorf("scry log retained a private subject or prompt: %+v", request)
		}
		for _, option := range request.Options {
			if option.Card.Exists || len(option.Targets) != 0 || option.Label != "Private option" {
				t.Errorf("scry log retained private option information: %+v", option)
			}
		}
	}
	for _, decision := range log.Choices {
		check(decision)
	}
	for _, entry := range log.Entries {
		if entry.Kind == TurnLogEntryChoice {
			check(entry.Choice)
		}
	}
	data, err := json.Marshal(GameResult{Turns: []TurnLog{*log}})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "Scry secret") {
		t.Fatal("returned game log contains a private library name")
	}
}

func TestScryPrivateChoicesAndReplay(t *testing.T) {
	for _, tc := range []struct {
		name    string
		size    int
		amount  int
		answers [][]int
		want    []int
	}{
		{"single top", 3, 1, [][]int{{0}}, []int{0, 1, 2}},
		{"single bottom", 3, 1, [][]int{{1}}, []int{1, 2, 0}},
		{"short library", 1, 4, [][]int{{1}}, []int{0}},
		{"reverse top", 5, 4, [][]int{nil, {3, 2, 1, 0}}, []int{3, 2, 1, 0, 4}},
		{"reverse bottom", 5, 4, [][]int{{0, 1, 2, 3}, {3, 2, 1, 0}}, []int{4, 3, 2, 1, 0}},
		{"both orders", 5, 4, [][]int{{0, 2}, {1, 0}, {1, 0}}, []int{3, 1, 4, 2, 0}},
		{"empty library", 0, 3, nil, nil},
		{"scry zero", 2, 0, nil, []int{0, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cards := make([]id.ID, tc.size)
			for i := tc.size - 1; i >= 0; i-- {
				cards[i] = addCardToLibrary(g, game.Player1, vanillaCreature(fmt.Sprintf("Scry secret %d", i), 2, 3))
			}
			addEffectSpellToStack(g, game.Player1, game.Scry{Player: game.ControllerReference(), Amount: game.Fixed(tc.amount)}, nil)
			replay := g.Clone()
			seen := 0
			agent := &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
				if obs.Player != game.Player1 || obs.DecisionMaker != game.Player1 {
					t.Fatal("scry was sent to the wrong actor or deciding agent")
				}
				if len(NewObservation(g, game.Player2).LibraryLookable(game.Player1)) != 0 {
					t.Fatal("opponent can see the private scry")
				}
				if tc.amount == 1 || tc.size == 1 {
					if !request.Subject.Exists || request.Subject.Val.CardID != cards[0] ||
						request.Subject.Val.Name != "Scry secret 0" {
						t.Fatal("chooser lost the actual private single-card subject")
					}
				} else {
					for _, option := range request.Options {
						if !option.Card.Exists || option.Card.Val.CardID == 0 ||
							!strings.HasPrefix(option.Label, "Scry secret ") ||
							!slices.Equal(option.Card.Val.Types, []types.Card{types.Creature}) {
							t.Fatal("chooser lost private partition/order identities")
						}
					}
				}
				if seen >= len(tc.answers) {
					t.Fatal("unexpected extra scry decision")
				}
				answer := tc.answers[seen]
				seen++
				return answer
			}}
			engine := NewEngine(nil)
			log := &TurnLog{}
			engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{agent}, log)
			if seen != len(tc.answers) || len(log.Choices) != len(tc.answers) {
				t.Fatalf("scry decisions = %d/%d, want %d", seen, len(log.Choices), len(tc.answers))
			}
			want := make([]id.ID, len(tc.want))
			for i, index := range tc.want {
				want[i] = cards[index]
			}
			if !slices.Equal(g.Players[game.Player1].Library.All(), want) {
				t.Fatalf("library = %v, want %v", g.Players[game.Player1].Library.All(), want)
			}
			assertPrivateScryLog(t, log)
			var answers [][]int
			for i, decision := range log.Choices {
				if decision.Request.ID != i+1 || decision.UsedFallback ||
					!slices.Equal(decision.Selected, tc.answers[i]) {
					t.Fatal("privacy sanitization changed replay indices")
				}
				answers = append(answers, slices.Clone(decision.Selected))
			}
			replayLog := &TurnLog{}
			engine.resolveTopOfStackWithChoices(replay, [game.NumPlayers]PlayerAgent{&choiceOnlyAgent{choices: answers}}, replayLog)
			if !slices.Equal(replay.Players[game.Player1].Library.All(), want) ||
				!reflect.DeepEqual(log, replayLog) {
				t.Fatal("sanitized replay choices did not reproduce the cloned game")
			}
			if len(cards) > 1 {
				replay.Players[game.Player1].Library.Remove(cards[0])
				if !g.Players[game.Player1].Library.Contains(cards[0]) {
					t.Fatal("scry clone shares its mutable library")
				}
			}
			wantEvents := 1
			if tc.amount == 0 {
				wantEvents = 0
			}
			if countEventsOfKind(g.Events, game.EventScry) != wantEvents ||
				countEventsOfKind(g.Events, game.EventCardRevealed) != 0 {
				t.Fatal("scry event or private in-library reorder semantics changed")
			}
			for _, event := range g.Events {
				if event.Kind == game.EventZoneChanged && slices.Contains(cards, event.CardID) {
					t.Fatal("scry created a zone change for a reordered library card")
				}
			}
		})
	}
}

func TestScryPrivateFallbackAndControlledDecision(t *testing.T) {
	for _, scenario := range []string{"no agent", "invalid answer", "controlled"} {
		t.Run(scenario, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToLibrary(g, game.Player1, vanillaCreature("Scry secret bottom", 2, 2))
			first := addCardToLibrary(g, game.Player1, vanillaCreature("Scry secret top", 2, 2))
			addEffectSpellToStack(g, game.Player1, game.Scry{Player: game.ControllerReference(), Amount: game.Fixed(1)}, nil)
			agents := [game.NumPlayers]PlayerAgent{}
			if scenario == "invalid answer" {
				agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{99}}}
			}
			if scenario == "controlled" {
				g.Turn.TurnNumber = 1
				g.PlayerControlEffects = []game.PlayerControlEffect{{
					ID: g.IDGen.Next(), Player: game.Player1, Controller: game.Player3, Turn: 1,
				}}
				agents[game.Player1] = &controlScript{choice: func(PlayerObservation, game.ChoiceRequest) []int {
					t.Fatal("controlled actor received its own decision")
					return nil
				}}
				agents[game.Player3] = &controlScript{choice: func(obs PlayerObservation, request game.ChoiceRequest) []int {
					if obs.Player != game.Player1 || obs.DecisionMaker != game.Player3 ||
						!request.Subject.Exists || request.Subject.Val.CardID != first {
						t.Fatal("controller did not receive actor's private scry subject")
					}
					return []int{1}
				}}
			}
			log := &TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)
			if len(log.Choices) != 1 || log.Choices[0].UsedFallback != (scenario != "controlled") {
				t.Fatal("scry fallback accounting changed")
			}
			wantDecider := game.Player1
			if scenario == "controlled" {
				wantDecider = game.Player3
			}
			if log.Choices[0].DecisionMaker != wantDecider {
				t.Fatal("logged scry decision maker is incorrect")
			}
			assertPrivateScryLog(t, log)
		})
	}
}

func TestRevealedLibraryOrderRemainsPublic(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCardToLibrary(g, game.Player1, vanillaCreature("Public first", 2, 2))
	second := addCardToLibrary(g, game.Player1, vanillaCreature("Public second", 2, 2))
	log := &TurnLog{}
	addEffectSpellToStack(g, game.Player1, game.RevealUntil{
		Player: game.ControllerReference(), Until: game.Selection{RequiredTypes: []types.Card{types.Land}},
		MatchToDestinationRestChosenBottom: true,
	}, nil)
	NewEngine(nil).resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{
		&choiceOnlyAgent{choices: [][]int{{1, 0}}},
	}, log)
	if len(log.Choices) != 1 || log.Choices[0].Request.Private ||
		!log.Choices[0].Request.Options[0].Card.Exists ||
		log.Choices[0].Request.Options[0].Label != "Public second" ||
		!slices.Equal(g.Players[game.Player1].Library.All(), []id.ID{first, second}) ||
		countEventsOfKind(g.Events, game.EventCardRevealed) != 2 {
		t.Fatal("public revealed order was privatized or failed to reorder")
	}
}

func TestScryPrivateMultiCardFallback(t *testing.T) {
	for _, invalid := range []bool{false, true} {
		t.Run(fmt.Sprint(invalid), func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			for range 3 {
				addCardToLibrary(g, game.Player1, vanillaCreature("Scry secret fallback", 2, 2))
			}
			want := g.Players[game.Player1].Library.All()
			addEffectSpellToStack(g, game.Player1, game.Scry{Player: game.ControllerReference(), Amount: game.Fixed(3)}, nil)
			agents := [game.NumPlayers]PlayerAgent{}
			if invalid {
				agents[game.Player1] = &choiceOnlyAgent{choices: [][]int{{99}, {0, 0, 0}}}
			}
			log := &TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)
			if !slices.Equal(g.Players[game.Player1].Library.All(), want) ||
				len(log.Choices) != 2 || !log.Choices[0].UsedFallback || !log.Choices[1].UsedFallback ||
				len(log.Choices[0].Selected) != 0 || !slices.Equal(log.Choices[1].Selected, []int{0, 1, 2}) ||
				g.Execution.ChoiceFallbacks != 2 || (invalid && g.Execution.InvalidChoices != 2) {
				t.Fatal("partition/order fallback did not preserve replay and library order")
			}
			assertPrivateScryLog(t, log)
		})
	}
}
