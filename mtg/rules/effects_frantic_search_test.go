package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

type franticSearchChoiceAgent struct {
	requests []game.ChoiceRequest
}

func (*franticSearchChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (a *franticSearchChoiceAgent) ChooseChoice(obs PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	switch request.Prompt {
	case "Choose cards to discard":
		if obs.Players()[int(game.Player1)].HandSize != 3 {
			return nil
		}
		return []int{0, 1}
	case "Choose permanents to untap":
		return []int{2, 0, 1}
	default:
		return nil
	}
}

func TestFranticSearchResolvesChoicesAndUntapEventsInOrder(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Old"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn A"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Drawn B"}})
	first := addBasicLandPermanent(g, game.Player1, types.Island)
	second := addBasicLandPermanent(g, game.Player2, types.Forest)
	third := addBasicLandPermanent(g, game.Player1, types.Plains)
	unselected := addBasicLandPermanent(g, game.Player1, types.Mountain)
	nonland := addCreaturePermanent(g, game.Player1)
	for _, permanent := range []*game.Permanent{first, second, third, unselected, nonland} {
		permanent.Tapped = true
	}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{
		{Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
		{Primitive: game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(2)}},
		{Primitive: game.Untap{
			Group: game.BattlefieldGroup(game.Selection{
				RequiredTypes: []types.Card{types.Land},
			}),
			ChooseUpTo: true,
			Amount:     game.Fixed(3),
		}},
	}, nil)

	agent := &franticSearchChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(agent.requests) != 2 ||
		agent.requests[1].MinChoices != 0 ||
		agent.requests[1].MaxChoices != 3 ||
		len(agent.requests[1].Options) != 4 {
		t.Fatalf("requests = %#v, want discard then bounded land untap", agent.requests)
	}
	if first.Tapped || second.Tapped || third.Tapped {
		t.Fatal("chosen lands remained tapped")
	}
	if !unselected.Tapped || !nonland.Tapped {
		t.Fatal("unchosen or nonland permanent was untapped")
	}
	var untapped []id.ID
	for _, event := range g.Events {
		if event.Kind == game.EventPermanentUntapped {
			untapped = append(untapped, event.PermanentID)
		}
	}
	want := []id.ID{third.ObjectID, first.ObjectID, second.ObjectID}
	if !slices.Equal(untapped, want) {
		t.Fatalf("untap events = %v, want selected order %v", untapped, want)
	}
}

func TestBoundedUntapRejectsDuplicateChoiceAndFallsBackDistinct(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for _, subtype := range []types.Sub{types.Island, types.Forest, types.Plains, types.Mountain} {
		addBasicLandPermanent(g, game.Player1, subtype).Tapped = true
	}
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{Primitive: game.Untap{
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Land},
		}),
		ChooseUpTo: true,
		Amount:     game.Fixed(3),
	}}}, nil)
	agent := &duplicateUntapChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	if len(log.Choices) != 1 || !log.Choices[0].UsedFallback {
		t.Fatalf("choices = %#v, want duplicate answer rejected", log.Choices)
	}
	selected := log.Choices[0].Selected
	if len(selected) != 3 || selected[0] == selected[1] || selected[0] == selected[2] || selected[1] == selected[2] {
		t.Fatalf("selected = %v, want three distinct fallback choices", selected)
	}
}

func TestBoundedUntapMayChooseNone(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	land := addBasicLandPermanent(g, game.Player1, types.Island)
	land.Tapped = true
	addInstructionSpellToStackForController(g, game.Player1, []game.Instruction{{Primitive: game.Untap{
		Group: game.BattlefieldGroup(game.Selection{
			RequiredTypes: []types.Card{types.Land},
		}),
		ChooseUpTo: true,
		Amount:     game.Fixed(3),
	}}}, nil)
	agent := &emptyUntapChoiceAgent{}
	var agents [game.NumPlayers]PlayerAgent
	agents[game.Player1] = agent
	log := &TurnLog{}
	NewEngine(nil).resolveTopOfStackWithChoices(g, agents, log)

	if !land.Tapped {
		t.Fatal("land was untapped after choosing none")
	}
	if len(log.Choices) != 1 ||
		log.Choices[0].UsedFallback ||
		len(log.Choices[0].Selected) != 0 ||
		log.Choices[0].Request.MinChoices != 0 {
		t.Fatalf("choices = %#v, want accepted empty up-to choice", log.Choices)
	}
}

type duplicateUntapChoiceAgent struct{}

func (*duplicateUntapChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (*duplicateUntapChoiceAgent) ChooseChoice(PlayerObservation, game.ChoiceRequest) []int {
	return []int{0, 0, 1}
}

type emptyUntapChoiceAgent struct{}

func (*emptyUntapChoiceAgent) ChooseAction(PlayerObservation, []action.Action) action.Action {
	return action.Pass()
}

func (*emptyUntapChoiceAgent) ChooseChoice(PlayerObservation, game.ChoiceRequest) []int {
	return nil
}
