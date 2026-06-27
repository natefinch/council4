package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestDigPutsChosenCardOntoBattlefieldTapped verifies the typed optional
// dig-to-battlefield: with a land filter and a tapped entry, only matching
// looked-at cards are offered, the chosen card enters the battlefield tapped,
// and every other looked-at card goes to the bottom of the library.
func TestDigPutsChosenCardOntoBattlefieldTapped(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Add bottom-to-top: c1 deepest, c3 on top. Seen order is [c3, c2, c1].
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bolt", Types: []types.Card{types.Instant}}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island", Types: []types.Card{types.Land}}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:       game.ControllerReference(),
		Look:         game.Fixed(3),
		Take:         game.Fixed(1),
		Remainder:    game.DigRemainderLibraryBottom,
		Filter:       opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
		TakeUpTo:     true,
		Destination:  zone.Battlefield,
		EntersTapped: true,
	}, nil)
	log := TurnLog{}
	// Eligible lands in seen order are [c3, c1]; choosing index 0 takes c3.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	permanent := permanentForCard(g, c3)
	if permanent == nil {
		t.Fatal("dig did not put the chosen card onto the battlefield")
	}
	if !permanent.Tapped {
		t.Fatal("dig put the chosen card onto the battlefield untapped, want tapped")
	}
	if permanent.Controller != game.Player1 {
		t.Fatalf("permanent controller = %v, want %v", permanent.Controller, game.Player1)
	}
	if player.Hand.Contains(c3) || player.Library.Contains(c3) {
		t.Fatal("dig left the chosen card off the battlefield")
	}
	if !player.Library.Contains(c1) || !player.Library.Contains(c2) {
		t.Fatal("dig did not bottom the unchosen looked-at cards")
	}
	if player.Graveyard.Contains(c1) || player.Graveyard.Contains(c2) {
		t.Fatal("dig sent a remainder card to the graveyard instead of the library bottom")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c3 && event.FromZone == zone.Library && event.ToZone == zone.Battlefield
	})
}

// TestDigBattlefieldMayPutNone verifies the "you may" semantics for the
// battlefield destination: when the digging player declines, no permanent is
// created and every looked-at card returns to the bottom of the library.
func TestDigBattlefieldMayPutNone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Forest", Types: []types.Card{types.Land}}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Island", Types: []types.Card{types.Land}}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:       game.ControllerReference(),
		Look:         game.Fixed(2),
		Take:         game.Fixed(1),
		Remainder:    game.DigRemainderLibraryBottom,
		Filter:       opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
		TakeUpTo:     true,
		Destination:  zone.Battlefield,
		EntersTapped: true,
	}, nil)
	log := TurnLog{}
	// An empty choice declines the optional put.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if len(g.Battlefield) != 0 {
		t.Fatalf("battlefield permanents = %d, want 0 (declined put)", len(g.Battlefield))
	}
	if !player.Library.Contains(c1) || !player.Library.Contains(c2) {
		t.Fatal("declined dig did not return the looked-at cards to the library")
	}
}
