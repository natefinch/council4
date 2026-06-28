package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestRevealTopPartitionGraveyardRemainder verifies the Mulch shape: the
// controller reveals the top cards, the matching (land) cards go to hand, and the
// remaining revealed cards go to the graveyard. The partition is deterministic,
// so no player choice is requested.
func TestRevealTopPartitionGraveyardRemainder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	land := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Forest",
		Types: []types.Card{types.Land},
	}})
	spell := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bolt",
		Types: []types.Card{types.Instant},
	}})
	creature := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bear",
		Types: []types.Card{types.Creature},
	}})
	addEffectSpellToStack(g, game.Player1, game.RevealTopPartition{
		Player:    game.ControllerReference(),
		Amount:    game.Fixed(3),
		Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
		Remainder: game.DigRemainderGraveyard,
	}, nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(land) {
		t.Fatal("reveal-top-partition did not put the land into hand")
	}
	if !player.Graveyard.Contains(spell) || !player.Graveyard.Contains(creature) {
		t.Fatal("reveal-top-partition did not send the non-land cards to the graveyard")
	}
	if player.Library.Contains(land) || player.Library.Contains(spell) || player.Library.Contains(creature) {
		t.Fatal("reveal-top-partition left a revealed card in the library")
	}
	if len(log.Choices) != 0 {
		t.Fatalf("choices = %+v, want no player choice", log.Choices)
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == land && event.FromZone == zone.Library && event.ToZone == zone.Hand
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == spell && event.FromZone == zone.Library && event.ToZone == zone.Graveyard
	})
}

// TestRevealTopPartitionLibraryBottomRemainder verifies the Goblin Ringleader
// shape: matching cards go to hand and the rest go to the bottom of the library
// instead of the graveyard.
func TestRevealTopPartitionLibraryBottomRemainder(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// bottom is added first (deepest); the top three are revealed.
	bottom := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "Test Bottom",
		Types: []types.Card{types.Creature},
	}})
	goblin := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Goblin",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}})
	elfA := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Elf A",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}})
	elfB := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:     "Test Elf B",
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Elf},
	}})
	addEffectSpellToStack(g, game.Player1, game.RevealTopPartition{
		Player:    game.ControllerReference(),
		Amount:    game.Fixed(3),
		Selection: game.Selection{SubtypesAny: []types.Sub{types.Goblin}},
		Remainder: game.DigRemainderLibraryBottom,
	}, nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(goblin) {
		t.Fatal("reveal-top-partition did not put the goblin into hand")
	}
	if player.Graveyard.Contains(elfA) || player.Graveyard.Contains(elfB) {
		t.Fatal("reveal-top-partition sent non-matching cards to the graveyard instead of the library bottom")
	}
	if !player.Library.Contains(elfA) || !player.Library.Contains(elfB) {
		t.Fatal("reveal-top-partition did not return the non-matching cards to the library")
	}
	// The untouched bottom card stays in the library.
	if !player.Library.Contains(bottom) {
		t.Fatal("reveal-top-partition disturbed the untouched bottom card")
	}
}
