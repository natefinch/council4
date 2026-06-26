package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TestDigTakesChosenCardAndGraveyardsRest verifies a Dig effect lets the player
// take their chosen card from the top of the library into hand and sends the
// remaining looked-at cards to the graveyard, using the choice agent.
func TestDigTakesChosenCardAndGraveyardsRest(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Add bottom-to-top: c1 is added first (deepest), c3 last (top). peekLibrary
	// returns top-first, so the seen order is c3, c2, c1.
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(3),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderGraveyard,
	}, nil)
	log := TurnLog{}
	// Seen order is [c3, c2, c1]; choosing index 1 takes the middle card c2.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c2) {
		t.Fatal("dig did not put the chosen card into hand")
	}
	if player.Library.Contains(c1) || player.Library.Contains(c2) || player.Library.Contains(c3) {
		t.Fatal("dig left a looked-at card in the library")
	}
	if !player.Graveyard.Contains(c1) || !player.Graveyard.Contains(c3) {
		t.Fatal("dig did not send the unchosen cards to the graveyard")
	}
	if player.Graveyard.Contains(c2) {
		t.Fatal("dig sent the chosen card to the graveyard")
	}
	if len(log.Choices) != 1 || log.Choices[0].Request.Kind != game.ChoiceDig || log.Choices[0].UsedFallback {
		t.Fatalf("choices = %+v, want non-fallback dig choice", log.Choices)
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c2 && event.FromZone == zone.Library && event.ToZone == zone.Hand
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c1 && event.FromZone == zone.Library && event.ToZone == zone.Graveyard
	})
}

// TestDigBottomsRestOfLibrary verifies a Dig with the library-bottom remainder
// takes the chosen card into hand and moves the unchosen looked-at cards to the
// bottom of the library rather than the graveyard.
func TestDigBottomsRestOfLibrary(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Add bottom-to-top: bottom is added first (deepest), c3 last (top).
	// peekLibrary returns top-first, so the seen order is c3, c2, c1.
	bottom := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bottom"}})
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(3),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderLibraryBottom,
	}, nil)
	log := TurnLog{}
	// Seen order is [c3, c2, c1]; choosing index 1 takes the middle card c2.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c2) {
		t.Fatal("dig did not put the chosen card into hand")
	}
	if player.Graveyard.Contains(c1) || player.Graveyard.Contains(c3) {
		t.Fatal("dig sent unchosen cards to the graveyard instead of the library bottom")
	}
	// The unchosen looked-at cards return to the library, beneath the untouched
	// bottom card, so they are now the deepest cards.
	if !player.Library.Contains(c1) || !player.Library.Contains(c3) {
		t.Fatal("dig did not return the unchosen cards to the library")
	}
	if player.Library.Contains(c2) {
		t.Fatal("dig left the chosen card in the library")
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c1 && event.FromZone == zone.Library && event.ToZone == zone.Library
	})
	// The card that was already on the bottom stays in the library.
	if !player.Library.Contains(bottom) {
		t.Fatal("dig disturbed the untouched bottom card")
	}
}

// TestDigFallsBackToTopCards verifies that, without a choosing agent, a Dig
// takes the topmost looked-at cards deterministically and graveyards the rest.
func TestDigFallsBackToTopCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(3),
		Take:      game.Fixed(2),
		Remainder: game.DigRemainderGraveyard,
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	player := g.Players[game.Player1]
	// Seen order is [c3, c2, c1]; the fallback takes the first two (c3, c2).
	if !player.Hand.Contains(c3) || !player.Hand.Contains(c2) {
		t.Fatal("dig fallback did not take the top two looked-at cards")
	}
	if !player.Graveyard.Contains(c1) {
		t.Fatal("dig fallback did not graveyard the remaining looked-at card")
	}
	if player.Hand.Contains(c1) {
		t.Fatal("dig fallback took more cards than requested")
	}
}

// TestDigFilterRevealTakesOnlyMatchingCard verifies the typed optional-reveal
// dig: with a creature-only filter, only the matching looked-at cards are
// offered, the chosen card is revealed as it enters the hand, and every other
// looked-at card (matching or not) goes to the bottom of the library.
func TestDigFilterRevealTakesOnlyMatchingCard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Add bottom-to-top: c1 deepest, c3 on top. Seen order is [c3, c2, c1].
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bolt", Types: []types.Card{types.Instant}}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Elk", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(3),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderLibraryBottom,
		Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		TakeUpTo:  true,
		Reveal:    true,
	}, nil)
	log := TurnLog{}
	// Eligible creatures in seen order are [c3, c1]; choosing index 0 takes c3.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c3) {
		t.Fatal("dig did not put the chosen creature into hand")
	}
	if player.Hand.Contains(c1) || player.Hand.Contains(c2) {
		t.Fatal("dig put an unchosen card into hand")
	}
	if !player.Library.Contains(c1) || !player.Library.Contains(c2) {
		t.Fatal("dig did not bottom the unchosen looked-at cards")
	}
	if player.Graveyard.Contains(c1) || player.Graveyard.Contains(c2) {
		t.Fatal("dig sent a remainder card to the graveyard instead of the library bottom")
	}
	assertEvent(t, g.Events, game.EventCardRevealed, func(event game.Event) bool {
		return event.CardID == c3
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c3 && event.FromZone == zone.Library && event.ToZone == zone.Hand
	})
}

// TestDigFilterRevealMayTakeNone verifies the "you may" semantics: when the
// digging player declines (an empty choice), no card is revealed or taken and
// every looked-at card returns to the bottom of the library.
func TestDigFilterRevealMayTakeNone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Bear", Types: []types.Card{types.Creature}}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Elk", Types: []types.Card{types.Creature}}})
	addEffectSpellToStack(g, game.Player1, game.Dig{
		Player:    game.ControllerReference(),
		Look:      game.Fixed(2),
		Take:      game.Fixed(1),
		Remainder: game.DigRemainderLibraryBottom,
		Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
		TakeUpTo:  true,
		Reveal:    true,
	}, nil)
	log := TurnLog{}
	// An empty choice declines the optional reveal.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{}}}}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if player.Hand.Contains(c1) || player.Hand.Contains(c2) {
		t.Fatal("declined dig still put a card into hand")
	}
	if !player.Library.Contains(c1) || !player.Library.Contains(c2) {
		t.Fatal("declined dig did not return the looked-at cards to the library")
	}
	for _, event := range g.Events {
		if event.Kind == game.EventCardRevealed {
			t.Fatal("declined dig revealed a card")
		}
	}
}
