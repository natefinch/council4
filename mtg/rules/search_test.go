package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/types"
)

// stubSearchAgent decides by exercising the SearchContext: it determinizes the
// state, enumerates legal actions with the Simulator, and returns a play-land
// action if one applies cleanly on the determinized world.
type stubSearchAgent struct {
	searched bool
}

func (*stubSearchAgent) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return action.Pass()
}

func (s *stubSearchAgent) ChooseActionBySearch(ctx SearchContext, _ []action.Action) action.Action {
	s.searched = true
	world := ctx.Determinize()
	sim := ctx.Simulator()
	for _, act := range sim.LegalActions(world, ctx.Player()) {
		if act.Kind != action.ActionPlayLand {
			continue
		}
		if next, ok := sim.Apply(world, ctx.Player(), act, simPassPolicies()); ok &&
			len(next.Battlefield) == len(world.Battlefield)+1 {
			return act
		}
	}
	return action.Pass()
}

func TestEngineRoutesPriorityThroughSearchAgent(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	setMainPhasePriority(g, game.Player1)

	stub := &stubSearchAgent{}
	legal := e.Simulator().LegalActions(g, game.Player1)
	chosen := e.decidePriorityAction(g, stub, game.Player1, legal)

	if !stub.searched {
		t.Fatal("engine did not route the priority decision through ChooseActionBySearch")
	}
	if chosen.Kind != action.ActionPlayLand {
		t.Fatalf("search agent chose %v, want the play-land action it found via the Simulator", chosen.Kind)
	}
	if len(g.Battlefield) != 0 {
		t.Fatal("searching mutated the live game state")
	}
}

func TestNonSearchAgentUsesObservationPath(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	setMainPhasePriority(g, game.Player1)

	// simPassPolicy is a plain PlayerAgent (not a SearchAgent), so the engine must
	// use the ordinary observation path and honor its choice (pass).
	legal := e.Simulator().LegalActions(g, game.Player1)
	if chosen := e.decidePriorityAction(g, simPassPolicy{}, game.Player1, legal); chosen.Kind != action.ActionPass {
		t.Fatalf("plain agent decision = %v, want pass", chosen.Kind)
	}
}

func TestDeterminizeIsIndependentClone(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addBasicLandPermanent(g, game.Player1, types.Forest)
	ctx := e.newSearchContext(g, game.Player1)

	world := ctx.Determinize()
	world.Battlefield = world.Battlefield[:0]
	if len(g.Battlefield) != 1 {
		t.Fatal("mutating the determinized world changed the live game")
	}
}
