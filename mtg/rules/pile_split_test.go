package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestPileSplitOpponentSeparatesControllerChooses exercises the Fact or Fiction
// shape: an opponent separates the revealed cards into two piles and the
// controller chooses which pile is put into hand; the other pile is graveyarded.
func TestPileSplitOpponentSeparatesControllerChooses(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	// Add bottom-to-top: c1 deepest, c3 top. peekLibrary is top-first: [c3,c2,c1].
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.PileSplit{
		Player:            game.ControllerReference(),
		Amount:            game.Fixed(3),
		SeparatorOpponent: true,
		ChooserOpponent:   false,
		Kept:              zone.Hand,
		Other:             zone.Graveyard,
	}, nil)
	log := TurnLog{}
	agents := [game.NumPlayers]PlayerAgent{
		// Opponent (Player2) separates: first pile is index 0 => {c3}.
		game.Player2: &choiceOnlyAgent{choices: [][]int{{0}}},
		// Controller (Player1) keeps the second pile (index 1) => {c2,c1}.
		game.Player1: &choiceOnlyAgent{choices: [][]int{{1}}},
	}

	engine.resolveTopOfStackWithChoices(g, agents, &log)

	player := g.Players[game.Player1]
	if !player.Hand.Contains(c2) || !player.Hand.Contains(c1) {
		t.Fatal("pile split did not put the kept pile into hand")
	}
	if player.Hand.Contains(c3) {
		t.Fatal("pile split put a card from the other pile into hand")
	}
	if !player.Graveyard.Contains(c3) {
		t.Fatal("pile split did not graveyard the other pile")
	}
	if player.Graveyard.Contains(c1) || player.Graveyard.Contains(c2) {
		t.Fatal("pile split graveyarded a kept card")
	}
	if len(log.Choices) != 2 {
		t.Fatalf("choices = %+v, want a separate and a choose decision", log.Choices)
	}
	if log.Choices[0].Request.Kind != game.ChoicePileSeparate || log.Choices[0].Request.Player != game.Player2 {
		t.Fatalf("first choice = %+v, want opponent separation", log.Choices[0])
	}
	if log.Choices[1].Request.Kind != game.ChoicePileChoose || log.Choices[1].Request.Player != game.Player1 {
		t.Fatalf("second choice = %+v, want controller choice", log.Choices[1])
	}
	if log.Choices[0].UsedFallback || log.Choices[1].UsedFallback {
		t.Fatalf("choices unexpectedly used fallback: %+v", log.Choices)
	}
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c1 && event.FromZone == zone.Library && event.ToZone == zone.Hand
	})
	assertEvent(t, g.Events, game.EventZoneChanged, func(event game.Event) bool {
		return event.CardID == c3 && event.FromZone == zone.Library && event.ToZone == zone.Graveyard
	})
}

// TestPileSplitFallbackSplitsEvenlyAndKeepsLargerPile verifies the deterministic
// fallback: with no answering agent the revealed cards split as evenly as
// possible and the larger pile is kept.
func TestPileSplitFallbackSplitsEvenlyAndKeepsLargerPile(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	c1 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Deep"}})
	c2 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Middle"}})
	c3 := addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Top"}})
	addEffectSpellToStack(g, game.Player1, game.PileSplit{
		Player:            game.ControllerReference(),
		Amount:            game.Fixed(3),
		SeparatorOpponent: true,
		ChooserOpponent:   false,
		Kept:              zone.Hand,
		Other:             zone.Graveyard,
	}, nil)
	log := TurnLog{}

	engine.resolveTopOfStack(g, &log)

	player := g.Players[game.Player1]
	// Seen [c3,c2,c1]; default first pile is the first len/2 = 1 card {c3};
	// second pile {c2,c1} is larger, so it is kept.
	if !player.Hand.Contains(c2) || !player.Hand.Contains(c1) {
		t.Fatal("pile split fallback did not keep the larger pile")
	}
	if !player.Graveyard.Contains(c3) {
		t.Fatal("pile split fallback did not graveyard the smaller pile")
	}
	if len(log.Choices) != 2 || !log.Choices[0].UsedFallback || !log.Choices[1].UsedFallback {
		t.Fatalf("choices = %+v, want two fallback decisions", log.Choices)
	}
}
