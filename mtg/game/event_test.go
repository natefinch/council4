package game

import (
	"testing"

	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestEventsForTurnReturnsCopy(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.Turn.TurnNumber = 1
	g.EventTurnStarts = []int{0}
	g.AppendEvent(Event{
		Kind:           EventSpellCast,
		CardTypes:      []types.Card{types.Instant},
		CardSupertypes: []types.Super{types.Legendary},
		CardSubtypes:   []types.Sub{types.Arcane},
		Colors:         []color.Color{color.Blue},
		TriggeredAbilities: []EventTriggeredAbility{{
			Controller: Player1,
		}},
	})

	events := g.EventsForTurn(1)
	events[0].Kind = EventDamageDealt
	events[0].CardTypes[0] = types.Creature
	events[0].CardSupertypes[0] = types.Basic
	events[0].CardSubtypes[0] = types.Spirit
	events[0].Colors[0] = color.Red
	events[0].TriggeredAbilities[0].Controller = Player2

	if g.Events[0].Kind != EventSpellCast {
		t.Fatalf("event kind mutated through EventsForTurn copy: %v", g.Events[0].Kind)
	}
	if g.Events[0].CardTypes[0] != types.Instant {
		t.Fatalf("event card types mutated through EventsForTurn copy: %v", g.Events[0].CardTypes)
	}
	if g.Events[0].CardSupertypes[0] != types.Legendary {
		t.Fatalf("event card supertypes mutated through EventsForTurn copy: %v", g.Events[0].CardSupertypes)
	}
	if g.Events[0].CardSubtypes[0] != types.Arcane {
		t.Fatalf("event card subtypes mutated through EventsForTurn copy: %v", g.Events[0].CardSubtypes)
	}
	if g.Events[0].Colors[0] != color.Blue {
		t.Fatalf("event colors mutated through EventsForTurn copy: %v", g.Events[0].Colors)
	}
	if g.Events[0].TriggeredAbilities[0].Controller != Player1 {
		t.Fatalf("event trigger snapshots mutated through EventsForTurn copy: %v", g.Events[0].TriggeredAbilities)
	}
}

func TestEventsThisTurnAndPreviousTurnReturnCopies(t *testing.T) {
	g := NewGame([NumPlayers]PlayerConfig{})
	g.Turn.TurnNumber = 2
	g.EventTurnStarts = []int{0, 1}
	g.AppendEvent(Event{Kind: EventLifeGained, Amount: 1})
	g.AppendEvent(Event{Kind: EventLifeLost, Amount: 2})

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
	cardSupertypes := []types.Super{types.Legendary}
	cardSubtypes := []types.Sub{types.Arcane}
	colors := []color.Color{color.Green}
	triggers := []EventTriggeredAbility{{Controller: Player1}}

	g.AppendEvent(Event{
		Kind:               EventSpellCast,
		CardTypes:          cardTypes,
		CardSupertypes:     cardSupertypes,
		CardSubtypes:       cardSubtypes,
		Colors:             colors,
		TriggeredAbilities: triggers,
	})
	cardTypes[0] = types.Artifact
	cardSupertypes[0] = types.Basic
	cardSubtypes[0] = types.Spirit
	colors[0] = color.Black
	triggers[0].Controller = Player2

	if g.Events[0].CardTypes[0] != types.Sorcery {
		t.Fatalf("AppendEvent aliased caller card types: %v", g.Events[0].CardTypes)
	}
	if g.Events[0].CardSupertypes[0] != types.Legendary {
		t.Fatalf("AppendEvent aliased caller card supertypes: %v", g.Events[0].CardSupertypes)
	}
	if g.Events[0].CardSubtypes[0] != types.Arcane {
		t.Fatalf("AppendEvent aliased caller card subtypes: %v", g.Events[0].CardSubtypes)
	}
	if g.Events[0].Colors[0] != color.Green {
		t.Fatalf("AppendEvent aliased caller colors: %v", g.Events[0].Colors)
	}
	if g.Events[0].TriggeredAbilities[0].Controller != Player1 {
		t.Fatalf("AppendEvent aliased caller trigger snapshots: %v", g.Events[0].TriggeredAbilities)
	}
}
