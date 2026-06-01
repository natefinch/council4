package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/types"
)

func TestEventsForTurnReturnsCopy(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.Turn.TurnNumber = 1
	g.EventTurnStarts = []int{0}
	g.AppendEvent(GameEvent{
		Kind:      EventSpellCast,
		CardTypes: []types.Card{types.Instant},
	})

	events := g.EventsForTurn(1)
	events[0].Kind = EventDamageDealt
	events[0].CardTypes[0] = types.Creature

	if g.Events[0].Kind != EventSpellCast {
		t.Fatalf("event kind mutated through EventsForTurn copy: %v", g.Events[0].Kind)
	}
	if g.Events[0].CardTypes[0] != types.Instant {
		t.Fatalf("event card types mutated through EventsForTurn copy: %v", g.Events[0].CardTypes)
	}
}

func TestEventsThisTurnAndPreviousTurnReturnCopies(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.Turn.TurnNumber = 2
	g.EventTurnStarts = []int{0, 1}
	g.AppendEvent(GameEvent{Kind: EventLifeGained, Amount: 1})
	g.AppendEvent(GameEvent{Kind: EventLifeLost, Amount: 2})

	previous := g.EventsPreviousTurn()
	current := g.EventsThisTurn()
	previous[0].Amount = 100
	current[0].Amount = 200

	if g.Events[0].Amount != 1 {
		t.Fatalf("previous turn event mutated through accessor copy: %+v", g.Events[0])
	}
	if g.Events[1].Amount != 2 {
		t.Fatalf("this turn event mutated through accessor copy: %+v", g.Events[1])
	}
}

func TestAppendEventCopiesEventSlices(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	cardTypes := []types.Card{types.Sorcery}

	g.AppendEvent(GameEvent{Kind: EventSpellCast, CardTypes: cardTypes})
	cardTypes[0] = types.Artifact

	if g.Events[0].CardTypes[0] != types.Sorcery {
		t.Fatalf("AppendEvent aliased caller card types: %v", g.Events[0].CardTypes)
	}
}
