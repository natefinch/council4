package rules

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/zone"
)

type libraryOrderChoiceAgent struct {
	order    []id.ID
	shuffle  bool
	requests []game.ChoiceRequest
}

func (*libraryOrderChoiceAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *libraryOrderChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	if request.Kind == game.ChoiceMay {
		if a.shuffle {
			return []int{1}
		}
		return []int{0}
	}
	selected := make([]int, 0, len(a.order))
	for _, cardID := range a.order {
		for _, option := range request.Options {
			if option.Card.Exists && option.Card.Val.CardID == cardID {
				selected = append(selected, option.Index)
				break
			}
		}
	}
	return selected
}

func TestPonderDeclinedShuffleKeepsOrderThenDrawsNormally(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bottom := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	first := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	third := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Third"}})
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.ReorderLibraryTop{Player: game.ControllerReference(), Amount: game.Fixed(3)}},
		{Primitive: game.ShuffleLibrary{Player: game.ControllerReference()}, Optional: true},
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
	})

	agent := &libraryOrderChoiceAgent{order: []id.ID{first, third, second}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if !g.Players[game.Player1].Hand.Contains(first) {
		t.Fatalf("hand = %v, want chosen top card %v", g.Players[game.Player1].Hand.All(), first)
	}
	if got, want := g.Players[game.Player1].Library.All(), []id.ID{third, second, bottom}; !slices.Equal(got, want) {
		t.Fatalf("library = %v, want declined-shuffle order %v", got, want)
	}
	var draws []game.Event
	for _, event := range g.Events {
		if event.Kind == game.EventCardDrawn {
			draws = append(draws, event)
		}
	}
	if len(draws) != 1 || draws[0].CardID != first || draws[0].Player != game.Player1 {
		t.Fatalf("draw events = %#v, want normal draw of %v", draws, first)
	}
}

func TestPonderAcceptedShuffleUsesNormalLibraryRandomizationThenDraws(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	fourth := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Fourth"}})
	first := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	third := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Third"}})
	chosen := []id.ID{first, third, second}
	addInstructionSpellToStack(g, []game.Instruction{
		{Primitive: game.ReorderLibraryTop{Player: game.ControllerReference(), Amount: game.Fixed(3)}},
		{Primitive: game.ShuffleLibrary{Player: game.ControllerReference()}, Optional: true},
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
	})

	agent := &libraryOrderChoiceAgent{order: chosen, shuffle: true}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	expected := zone.New(zone.Library)
	for i := len(chosen); i >= 0; i-- {
		cardID := fourth
		if i < len(chosen) {
			cardID = chosen[i]
		}
		expected.Add(cardID)
	}
	expected.Shuffle(rand.New(rand.NewPCG(1, 2)))
	drawn, ok := expected.Top()
	if !ok {
		t.Fatal("expected shuffled library unexpectedly empty")
	}
	expected.Remove(drawn)
	if !g.Players[game.Player1].Hand.Contains(drawn) {
		t.Fatalf("hand = %v, want shuffled top card %v", g.Players[game.Player1].Hand.All(), drawn)
	}
	if got, want := g.Players[game.Player1].Library.All(), expected.All(); !slices.Equal(got, want) {
		t.Fatalf("library = %v, want normal helper shuffle result %v", got, want)
	}
}

func TestReorderLibraryTopPreservesChosenExactOrder(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	bottom := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	firstCopy := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Copy"}})
	middle := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	secondCopy := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Copy"}})
	addInstructionSpellToStack(g, []game.Instruction{{Primitive: game.ReorderLibraryTop{
		Player: game.ControllerReference(),
		Amount: game.Fixed(3),
	}}})

	agent := &libraryOrderChoiceAgent{order: []id.ID{firstCopy, secondCopy, middle}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	got := g.Players[game.Player1].Library.All()
	want := []id.ID{firstCopy, secondCopy, middle, bottom}
	if !slices.Equal(got, want) {
		t.Fatalf("library = %v, want exact chosen order %v", got, want)
	}
	if len(agent.requests) != 1 {
		t.Fatalf("requests = %#v, want one", agent.requests)
	}
	request := agent.requests[0]
	if request.Kind != game.ChoiceOrder || request.Player != game.Player1 ||
		request.MinChoices != 3 || request.MaxChoices != 3 {
		t.Fatalf("request = %#v, want controller ordering three cards", request)
	}
}

func TestPonderHandlesEmptyAndShortLibraries(t *testing.T) {
	t.Parallel()
	for _, size := range []int{0, 1, 2} {
		t.Run(fmt.Sprintf("%d cards", size), func(t *testing.T) {
			t.Parallel()
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			var cards []id.ID
			for i := range size {
				cards = append(cards, addCardToLibrary(g, game.Player1, &game.CardDef{
					CardFace: game.CardFace{Name: string(rune('A' + i))},
				}))
			}
			order := slices.Clone(cards)
			slices.Reverse(order)
			addInstructionSpellToStack(g, []game.Instruction{
				{Primitive: game.ReorderLibraryTop{Player: game.ControllerReference(), Amount: game.Fixed(3)}},
				{Primitive: game.ShuffleLibrary{Player: game.ControllerReference()}, Optional: true},
				{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)}},
			})
			agent := &libraryOrderChoiceAgent{order: order}
			var agents [game.NumPlayers]PlayerAgent
			agents[game.Player1] = agent
			NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

			orderRequests := 0
			for _, request := range agent.requests {
				if request.Kind == game.ChoiceOrder {
					orderRequests++
					if len(request.Options) != size {
						t.Fatalf("order options = %d, want available cards %d", len(request.Options), size)
					}
					if request.Player != game.Player1 {
						t.Fatalf("order request player = %v, want library owner", request.Player)
					}
					for i, option := range request.Options {
						if !option.Card.Exists || option.Card.Val.CardID != order[i] {
							t.Fatalf("order option %d = %#v, want looked-at card %v", i, option, order[i])
						}
					}
				}
			}
			wantOrderRequests := 0
			if size > 0 {
				wantOrderRequests = 1
			}
			if orderRequests != wantOrderRequests {
				t.Fatalf("order requests = %d, want %d", orderRequests, wantOrderRequests)
			}
			if size == 0 {
				if !g.FailedDraws[game.Player1] {
					t.Fatal("empty-library draw did not record failure")
				}
				return
			}
			if !g.Players[game.Player1].Hand.Contains(order[0]) {
				t.Fatalf("hand = %v, want available chosen top %v", g.Players[game.Player1].Hand.All(), order[0])
			}
		})
	}
}

func TestReorderedLibraryStateClonesIndependently(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	first := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "First"}})
	second := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Second"}})
	addInstructionSpellToStack(g, []game.Instruction{{Primitive: game.ReorderLibraryTop{
		Player: game.ControllerReference(),
		Amount: game.Fixed(2),
	}}})
	agent := &libraryOrderChoiceAgent{order: []id.ID{first, second}}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	clone := g.Clone()
	if got, want := clone.Players[game.Player1].Library.All(), g.Players[game.Player1].Library.All(); !slices.Equal(got, want) {
		t.Fatalf("clone library = %v, want %v", got, want)
	}
	clone.Players[game.Player1].Library.Remove(first)
	if !g.Players[game.Player1].Library.Contains(first) {
		t.Fatal("mutating clone library changed original reordered library")
	}
}
