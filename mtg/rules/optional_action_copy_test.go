package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func TestOptionalGroupCopiedResolutionMakesOwnDecision(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	source := addCreaturePermanent(g, game.Player1)
	content := compiledScopedResultContent(t, "You may draw a card and gain 2 life.")
	for range 2 {
		addCardToLibrary(g, game.Player1, evidenceCard("Drawn", 1))
	}
	trigger := game.TriggeredAbility{Content: content}
	original := &game.StackObject{
		ID: g.IDGen.Next(), Kind: game.StackTriggeredAbility, SourceID: source.ObjectID,
		SourceCardID: source.CardInstanceID, Controller: game.Player1, InlineTrigger: &trigger,
	}
	g.Stack.Push(original)
	addEffectSpellToStack(g, game.Player1,
		game.CopyStackObject{Object: game.TargetStackObjectReference(0)},
		[]game.Target{game.StackObjectTarget(original.ID)})
	engine.resolveTopOfStack(g, &TurnLog{})
	copy, ok := g.Stack.Peek()
	if !ok || !copy.Copy || copy.ID == original.ID || copy.SourceID != original.SourceID {
		t.Fatal("did not produce an independent copy of the ability")
	}
	agent := &scopedMayAgent{accept: []bool{true, false}}
	agents := [game.NumPlayers]PlayerAgent{}
	agents[game.Player1] = agent
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})
	if agent.next != 2 || g.Players[game.Player1].Life != 42 || g.Players[game.Player1].Hand.Size() != 1 {
		t.Fatal("original inherited the copy's accepted decision")
	}
}
