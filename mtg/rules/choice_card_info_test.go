package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// recordingChoiceAgent records every ChoiceRequest it is asked about and
// answers with a scripted selection.
type recordingChoiceAgent struct {
	requests []game.ChoiceRequest
	answer   []int
}

func (*recordingChoiceAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return actionBuild.pass()
}

func (a *recordingChoiceAgent) ChooseChoice(_ PlayerObservation, request game.ChoiceRequest) []int {
	a.requests = append(a.requests, request)
	return a.answer
}

func TestCardChoiceInfoPopulatesPublicFields(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	cardID := g.IDGen.Next()
	g.CardInstances[cardID] = &game.CardInstance{ID: cardID, Owner: game.Player1, Def: &game.CardDef{CardFace: game.CardFace{
		Name:  "Llanowar Elves",
		Types: []types.Card{types.Creature},
	}}}

	info := cardChoiceInfo(g, cardID)
	if !info.Exists {
		t.Fatal("cardChoiceInfo returned unset for a known card")
	}
	if info.Val.CardID != cardID || info.Val.Name != "Llanowar Elves" {
		t.Errorf("info = %+v, want Llanowar Elves with the card id", info.Val)
	}
	if len(info.Val.Types) != 1 || info.Val.Types[0] != types.Creature {
		t.Errorf("info types = %v, want [Creature]", info.Val.Types)
	}
}

func TestCardChoiceInfoUnknownCardUnset(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	if info := cardChoiceInfo(g, g.IDGen.Next()); info.Exists {
		t.Errorf("cardChoiceInfo for an unknown card = %+v, want unset", info)
	}
}

func TestScrySubjectIdentifiesCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Forest",
		Types: []types.Card{types.Land},
	}})
	addEffectSpellToStack(g, game.Player1, game.Scry{Amount: game.Fixed(1), Player: game.ControllerReference()}, nil)

	agent := &recordingChoiceAgent{answer: []int{0}}
	agents := [game.NumPlayers]PlayerAgent{game.Player1: agent}
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if len(agent.requests) != 1 {
		t.Fatalf("recorded %d scry requests, want 1", len(agent.requests))
	}
	subject := agent.requests[0].Subject
	if !subject.Exists {
		t.Fatal("scry request Subject is unset; the agent cannot tell which card it is scrying")
	}
	if subject.Val.Name != "Forest" {
		t.Errorf("scry Subject name = %q, want Forest (the top card)", subject.Val.Name)
	}
}
