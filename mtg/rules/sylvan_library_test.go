package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/cards/s"
	"github.com/natefinch/council4/mtg/game"
)

// addSylvanLibrary puts the real Sylvan Library card onto controller's
// battlefield so its "At the beginning of your draw step, you may draw two
// additional cards. If you do, choose two cards in your hand drawn this turn.
// For each of those cards, pay 4 life or put the card on top of your library."
// trigger runs through the real resolution path.
func addSylvanLibrary(g *game.Game, controller game.PlayerID) *game.Permanent {
	return addCombatPermanent(g, controller, s.SylvanLibrary())
}

// resolveSylvanDrawStep emits the beginning-of-draw-step event on controller's
// turn and resolves the resulting trigger with agents. It reports whether a
// trigger was put on the stack.
func resolveSylvanDrawStep(engine *Engine, g *game.Game, controller game.PlayerID, agents [game.NumPlayers]PlayerAgent) bool {
	g.Turn.ActivePlayer = controller
	emitBeginningOfStepEvent(g, game.StepDraw)
	if !engine.putTriggeredAbilitiesOnStack(g) {
		return false
	}
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, agents, &log)
	return true
}

// TestSylvanLibraryPaysLifeAndPutsOnTop proves the whole Sylvan Library filter:
// the controller accepts the optional draw of two, chooses both freshly drawn
// cards, pays 4 life to keep the first, and puts the second on top of their
// library.
func TestSylvanLibraryPaysLifeAndPutsOnTop(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addSylvanLibrary(g, game.Player1)

	// Library, top-first, is [payCard, topCard]: payCard is drawn first and kept
	// for life, topCard is drawn second and put back on top.
	topCard := addCardToLibraryNamed(g, game.Player1, "Top Card")
	payCard := addCardToLibraryNamed(g, game.Player1, "Pay Card")

	startingLife := g.Players[game.Player1].Life
	startingHand := g.Players[game.Player1].Hand.Size()

	// Accept the draw, choose both drawn cards, pay life for payCard, put topCard
	// on top.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0, 1}, {0}, {1}}}}
	if !resolveSylvanDrawStep(engine, g, game.Player1, agents) {
		t.Fatal("beginning-of-draw-step trigger did not fire on the controller's draw step")
	}

	player := g.Players[game.Player1]
	if got := player.Life; got != startingLife-4 {
		t.Fatalf("life = %d, want %d after paying 4 life for one card", got, startingLife-4)
	}
	if !player.Hand.Contains(payCard) {
		t.Fatal("paid-for card is not in hand; paying life should keep it")
	}
	if player.Hand.Contains(topCard) {
		t.Fatal("card chosen for the library is still in hand")
	}
	top, ok := player.Library.Top()
	if !ok || top != topCard {
		t.Fatalf("library top = %v (ok=%v), want the card put on top %v", top, ok, topCard)
	}
	if got := player.Hand.Size(); got != startingHand+1 {
		t.Fatalf("hand size = %d, want %d (drew two, put one back)", got, startingHand+1)
	}
}

// TestSylvanLibraryDeclineDrawsNothing proves the "if you do" gate: declining the
// optional draw skips the whole filter, so no cards are drawn, no life is paid,
// and the library is untouched.
func TestSylvanLibraryDeclineDrawsNothing(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addSylvanLibrary(g, game.Player1)
	addCardToLibraryNamed(g, game.Player1, "Top Card")
	addCardToLibraryNamed(g, game.Player1, "Pay Card")

	startingLife := g.Players[game.Player1].Life
	startingHand := g.Players[game.Player1].Hand.Size()
	startingLibrary := g.Players[game.Player1].Library.Size()

	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{0}}}}
	if !resolveSylvanDrawStep(engine, g, game.Player1, agents) {
		t.Fatal("beginning-of-draw-step trigger did not fire on the controller's draw step")
	}

	player := g.Players[game.Player1]
	if got := player.Life; got != startingLife {
		t.Fatalf("life = %d, want %d (declining pays nothing)", got, startingLife)
	}
	if got := player.Hand.Size(); got != startingHand {
		t.Fatalf("hand size = %d, want %d (declining draws nothing)", got, startingHand)
	}
	if got := player.Library.Size(); got != startingLibrary {
		t.Fatalf("library size = %d, want %d (declining leaves the library untouched)", got, startingLibrary)
	}
}

// TestSylvanLibraryForcesTopWhenLifeTooLow proves CR 119.4: a player who cannot
// pay 4 life must put the chosen card on top of their library instead, even if
// they would rather pay.
func TestSylvanLibraryForcesTopWhenLifeTooLow(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addSylvanLibrary(g, game.Player1)
	firstCard := addCardToLibraryNamed(g, game.Player1, "First Card")
	secondCard := addCardToLibraryNamed(g, game.Player1, "Second Card")

	g.Players[game.Player1].Life = 2
	startingHand := g.Players[game.Player1].Hand.Size()

	// Accept the draw and choose both cards, then try to pay for both. With only
	// 2 life, neither payment is possible, so both cards go on top.
	agents := [game.NumPlayers]PlayerAgent{game.Player1: &choiceOnlyAgent{choices: [][]int{{1}, {0, 1}, {0}, {0}}}}
	if !resolveSylvanDrawStep(engine, g, game.Player1, agents) {
		t.Fatal("beginning-of-draw-step trigger did not fire on the controller's draw step")
	}

	player := g.Players[game.Player1]
	if got := player.Life; got != 2 {
		t.Fatalf("life = %d, want 2 (a player who can't pay 4 life pays nothing)", got)
	}
	if player.Hand.Contains(firstCard) || player.Hand.Contains(secondCard) {
		t.Fatal("a drawn card stayed in hand even though its owner could not pay life")
	}
	if got := player.Hand.Size(); got != startingHand {
		t.Fatalf("hand size = %d, want %d (drew two, put both back)", got, startingHand)
	}
	if got := player.Library.Size(); got != 2 {
		t.Fatalf("library size = %d, want 2 (both cards returned to the top)", got)
	}
}
