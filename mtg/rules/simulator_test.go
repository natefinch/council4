package rules

import (
	"math/rand/v2"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/action"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// simPassPolicy is a rollout/choice policy that always passes priority, used to
// drive the Simulator API deterministically in tests.
type simPassPolicy struct{}

func (simPassPolicy) ChooseAction(_ PlayerObservation, _ []action.Action) action.Action {
	return action.Pass()
}

func simPassPolicies() [game.NumPlayers]PlayerAgent {
	var policies [game.NumPlayers]PlayerAgent
	for i := range policies {
		policies[i] = simPassPolicy{}
	}
	return policies
}

func newSimEngine() *Engine {
	return NewEngine(rand.New(rand.NewPCG(1, 2)))
}

func zeroCostCreature(name string) *game.CardDef {
	pt := game.PT{Value: 2}
	return &game.CardDef{CardFace: game.CardFace{
		Name:      name,
		ManaCost:  opt.Val(cost.Mana{cost.O(0)}),
		Types:     []types.Card{types.Creature},
		Power:     opt.Val(pt),
		Toughness: opt.Val(pt),
	}}
}

func simHasKind(actions []action.Action, kind action.Kind) bool {
	_, ok := simFirstOfKind(actions, kind)
	return ok
}

func simFirstOfKind(actions []action.Action, kind action.Kind) (action.Action, bool) {
	for i := range actions {
		if actions[i].Kind == kind {
			return actions[i], true
		}
	}
	return action.Action{}, false
}

func simOnBattlefield(g *game.Game, cardID id.ID) bool {
	for _, permanent := range g.Battlefield {
		if permanent.CardInstanceID == cardID {
			return true
		}
	}
	return false
}

func TestSimulatorLegalActionsMatchesEngine(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	setMainPhasePriority(g, game.Player1)

	legal := e.LegalActions(g, game.Player1)
	if !simHasKind(legal, action.ActionPlayLand) {
		t.Fatalf("LegalActions = %v, want a play-land action", summarizeLegalActions(g, legal))
	}
}

func TestSimulateActionBranchesWithoutMutatingOriginal(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	setMainPhasePriority(g, game.Player1)

	landAction, ok := simFirstOfKind(e.LegalActions(g, game.Player1), action.ActionPlayLand)
	if !ok {
		t.Fatal("no play-land action offered")
	}
	before := len(g.Battlefield)

	next, ok := e.SimulateAction(g, game.Player1, landAction, simPassPolicies())
	if !ok {
		t.Fatal("SimulateAction reported the land play illegal")
	}
	if len(g.Battlefield) != before {
		t.Fatalf("original battlefield mutated: had %d permanents, now %d", before, len(g.Battlefield))
	}
	if len(next.Battlefield) != before+1 {
		t.Fatalf("clone battlefield = %d permanents, want %d", len(next.Battlefield), before+1)
	}
}

func TestSimulateActionRejectsIllegalAction(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	setMainPhasePriority(g, game.Player1)

	// Playing a land that is not in hand is not a legal action.
	bogus := action.PlayLandFace(id.ID(999999), game.FaceFront)
	if _, ok := e.SimulateAction(g, game.Player1, bogus, simPassPolicies()); ok {
		t.Fatal("SimulateAction accepted an illegal land play")
	}
}

func TestResolvePriorityResolvesStackedSpell(t *testing.T) {
	e := newSimEngine()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	creatureID := addCardToHand(g, game.Player1, zeroCostCreature("Bear"))
	setMainPhasePriority(g, game.Player1)

	castAction, ok := simFirstOfKind(e.LegalActions(g, game.Player1), action.ActionCastSpell)
	if !ok {
		t.Fatal("no cast action offered for the zero-cost creature")
	}
	afterCast, ok := e.SimulateAction(g, game.Player1, castAction, simPassPolicies())
	if !ok {
		t.Fatal("SimulateAction reported the cast illegal")
	}
	if afterCast.Stack.IsEmpty() {
		t.Fatal("expected the creature spell on the stack after SimulateAction")
	}

	resolved := e.ResolvePriority(afterCast, simPassPolicies())
	if !resolved.Stack.IsEmpty() {
		t.Fatal("ResolvePriority left the stack non-empty")
	}
	if !simOnBattlefield(resolved, creatureID) {
		t.Fatal("creature did not resolve onto the battlefield")
	}
	// The source state is untouched: its stack still holds the spell.
	if afterCast.Stack.IsEmpty() {
		t.Fatal("ResolvePriority mutated the source state's stack")
	}
}
