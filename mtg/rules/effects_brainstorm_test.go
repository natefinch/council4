package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

type orderedHandChoiceAgent struct {
	order        []string
	observedHand []string
	requests     []game.ChoiceRequest
	answer       []int
}

func (*orderedHandChoiceAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *orderedHandChoiceAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	for _, card := range obs.Hand() {
		a.observedHand = append(a.observedHand, card.Name)
	}
	if a.answer != nil {
		return append([]int(nil), a.answer...)
	}
	selected := make([]int, 0, len(a.order))
	for _, name := range a.order {
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.Name == name {
				selected = append(selected, option.Index)
				break
			}
		}
	}
	return selected
}

func brainstormInstructions() []game.Instruction {
	return []game.Instruction{
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(3)}},
		{Primitive: game.MoveCard{
			Player:      game.ControllerReference(),
			Amount:      game.Fixed(2),
			FromZone:    zone.Hand,
			Destination: zone.Library,
		}},
	}
}

func TestBrainstormDrawsThenChoosesIncludingDrawnCardsInTopOrder(t *testing.T) {
	t.Parallel()
	for _, order := range [][]string{
		{"Drawn C", "Old A"},
		{"Old A", "Drawn C"},
	} {
		order := append([]string(nil), order...)
		t.Run(order[0]+" first", func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old A"}})
			addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old B"}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Anchor"}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn A"}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn B"}})
			addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn C"}})
			addInstructionSpellToStack(g, brainstormInstructions())

			agent := &orderedHandChoiceAgent{order: order}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = agent
			log := &TurnLog{}
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

			if !slices.Contains(agent.observedHand, "Drawn C") {
				t.Fatalf("choice observation hand = %v, want newly drawn card", agent.observedHand)
			}
			if len(agent.requests) != 1 {
				t.Fatalf("choice requests = %d, want 1", len(agent.requests))
			}
			request := agent.requests[0]
			if request.Player != game.Player1 || request.MinChoices != 2 || request.MaxChoices != 2 {
				t.Fatalf("request = %#v, want controller exact two", request)
			}
			if len(log.Choices) != 1 || len(log.Choices[0].Selected) != 2 ||
				log.Choices[0].Selected[0] == log.Choices[0].Selected[1] {
				t.Fatalf("choice log = %#v, want two distinct ordered selections", log.Choices)
			}
			assertLibraryTopNames(t, g, game.Player1, order)
			assertBrainstormZoneEvents(t, g.Events, 2)
		})
	}
}

func TestBrainstormInvalidChoiceFallsBackToExactDistinctCards(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "A"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "B"}})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "C"}})
	addInstructionSpellToStack(g, brainstormInstructions()[1:])

	agent := &orderedHandChoiceAgent{answer: []int{0, 0}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	if len(log.Choices) != 1 || !log.Choices[0].UsedFallback {
		t.Fatalf("choices = %#v, want fallback", log.Choices)
	}
	selected := log.Choices[0].Selected
	if len(selected) != 2 || selected[0] == selected[1] {
		t.Fatalf("fallback selected = %v, want exact distinct pair", selected)
	}
}

func TestBrainstormInsufficientHandMovesAllAvailable(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	only := addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Only"}})
	addInstructionSpellToStack(g, brainstormInstructions()[1:])

	agent := &orderedHandChoiceAgent{order: []string{"Only"}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	if g.Players[game.Player1].Hand.Size() != 0 {
		t.Fatalf("hand size = %d, want 0", g.Players[game.Player1].Hand.Size())
	}
	top, ok := g.Players[game.Player1].Library.Top()
	if !ok || top != only {
		t.Fatalf("library top = %v/%v, want only card %v", top, ok, only)
	}
	if len(log.Choices) != 1 ||
		log.Choices[0].Request.MinChoices != 1 ||
		log.Choices[0].Request.MaxChoices != 1 {
		t.Fatalf("choices = %#v, want exact one available card", log.Choices)
	}
	assertBrainstormZoneEvents(t, g.Events, 1)
}

func assertLibraryTopNames(t *testing.T, g *game.Game, playerID game.PlayerID, want []string) {
	t.Helper()
	ids := g.Players[playerID].Library.All()
	if len(ids) < len(want) {
		t.Fatalf("library = %v, want at least %d cards", ids, len(want))
	}
	for i, name := range want {
		card, ok := g.GetCardInstance(ids[i])
		if !ok || card.Def.Name != name {
			t.Fatalf("library[%d] = %#v, want %q", i, card, name)
		}
	}
}

func assertBrainstormZoneEvents(t *testing.T, events []game.Event, want int) {
	t.Helper()
	var moves []game.Event
	for _, event := range events {
		if event.Kind == game.EventZoneChanged &&
			event.FromZone == zone.Hand &&
			event.ToZone == zone.Library {
			moves = append(moves, event)
		}
	}
	if len(moves) != want {
		t.Fatalf("hand-to-library events = %#v, want %d", moves, want)
	}
	var simultaneousID id.ID
	for i, event := range moves {
		if event.CardID == 0 {
			t.Fatalf("move event %d has no card id: %#v", i, event)
		}
		if i == 0 {
			simultaneousID = event.SimultaneousID
		}
		if event.SimultaneousID == 0 || event.SimultaneousID != simultaneousID {
			t.Fatalf("move events not one simultaneous batch: %#v", moves)
		}
	}
}
