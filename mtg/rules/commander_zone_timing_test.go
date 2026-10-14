package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func TestCommanderOwnerMayDeclineSBAOncePerIncarnation(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	commander := addCommanderPermanent(g, game.Player1)
	commander.Controller = game.Player2
	owner := &choiceOnlyAgent{choices: [][]int{{0}, {1}}}
	controller := &choiceOnlyAgent{choices: [][]int{{1}}}
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{
		game.Player1: owner, game.Player2: controller,
	}, &TurnLog{})
	destroyPermanent(g, commander.ObjectID)
	engine.applyStateBasedActions(g)
	if !g.Players[game.Player1].Graveyard.Contains(commander.CardInstanceID) || owner.next != 1 || controller.next != 0 {
		t.Fatal("SBA did not honor the owner's decision to leave the commander in graveyard")
	}
	engine.applyStateBasedActions(g)
	if owner.next != 1 {
		t.Fatal("declined SBA was offered again without a new incarnation")
	}
	clone := g.Clone()
	engine.applyStateBasedActions(clone)
	if !clone.Players[game.Player1].Graveyard.Contains(commander.CardInstanceID) ||
		clone.CommanderEventCursor != g.CommanderEventCursor {
		t.Fatal("clone lost the declined-incarnation checkpoint")
	}
	moveCardBetweenZones(g, game.Player1, commander.CardInstanceID, zone.Graveyard, zone.Exile)
	engine.applyStateBasedActions(g)
	if owner.next != 2 || !g.Players[game.Player1].CommandZone.Contains(commander.CardInstanceID) {
		t.Fatal("a new exile incarnation did not offer the owner a new choice")
	}
}

func TestCommanderHandLibraryReplacementIsOptional(t *testing.T) {
	t.Parallel()
	for _, destination := range []zone.Type{zone.Hand, zone.Library} {
		g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
		engine := NewEngine(nil)
		commander := addCommanderPermanent(g, game.Player1)
		engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{
			game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}},
		}, &TurnLog{})
		movePermanentToZone(g, commander, destination)
		engine.applyStateBasedActions(g)
		cards, _ := destinationZone(g, game.Player1, destination)
		if !cards.Contains(commander.CardInstanceID) || g.Players[game.Player1].CommandZone.Contains(commander.CardInstanceID) {
			t.Fatal("declined hand/library replacement was applied anyway")
		}
	}
}

func TestCommanderSBAIgnoresIncarnationThatLeftDuringResolution(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := addCommanderPermanent(g, game.Player1)
	destroyPermanent(g, commander.ObjectID)
	agent := &choiceOnlyAgent{choices: [][]int{{0}}}
	engine := NewEngine(nil)
	engine.setReplacementChoiceContext(g, [game.NumPlayers]PlayerAgent{game.Player1: agent}, &TurnLog{})
	moveCardBetweenZones(g, game.Player1, commander.CardInstanceID, zone.Graveyard, zone.Hand)
	engine.applyStateBasedActions(g)
	if agent.next != 1 || !g.Players[game.Player1].Hand.Contains(commander.CardInstanceID) {
		t.Fatal("a stale graveyard-arrival event prompted a second command-zone decision")
	}
}

func TestCommanderSimulatorUsesOwnerChoice(t *testing.T) {
	t.Parallel()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	commander := addCommanderPermanent(g, game.Player1)
	destroyPermanent(g, commander.ObjectID)
	setMainPhasePriority(g, game.Player1)
	owner := &choiceOnlyAgent{choices: [][]int{{0}}}
	policies := simPassPolicies()
	policies[game.Player1] = owner
	resolved := NewEngine(nil).Simulator().ResolvePriority(g, policies)
	if owner.next != 1 || !resolved.Players[game.Player1].Graveyard.Contains(commander.CardInstanceID) {
		t.Fatal("Simulator did not route the command-zone SBA choice to the owner")
	}
	if g.CommanderEventCursor != 0 || !g.Players[game.Player1].Graveyard.Contains(commander.CardInstanceID) {
		t.Fatal("Simulator mutated the original commander checkpoint")
	}
}
