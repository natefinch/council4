package rules

import (
	"testing"

	cardw "github.com/natefinch/council4/mtg/cards/w"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// wellOfLostDreamsScratchCard is a vanilla artifact used to seed a player's
// library with drawable cards for the Well of Lost Dreams draw tests.
func wellOfLostDreamsScratchCard() *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{
		Name:  "Well Scratch Card",
		Types: []types.Card{types.Artifact},
	}}
}

// fireWellOfLostDreams stages the real Well of Lost Dreams card on Player1's
// battlefield, gives Player1 landCount Islands to fund the optional {1} payments
// and libraryCount cards to draw, emits a life-gain event of the given amount so
// the "whenever you gain life" trigger records it, resolves the trigger, and
// returns Player1's resulting hand size. The default agent pays the optional
// resolution cost greedily, so the number of {1} payments is min(landCount,
// lifeGained) — the amount of life gained bounds it via the PayRepeatedly
// MaxCount, and the mana available bounds it otherwise.
func fireWellOfLostDreams(t *testing.T, lifeGained, landCount, libraryCount int) int {
	t.Helper()
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	if issues := game.ValidateCardDef(cardw.WellOfLostDreams()); len(issues) != 0 {
		t.Fatalf("carddef invalid: %v", issues)
	}
	addCombatPermanent(g, game.Player1, cardw.WellOfLostDreams())
	for range landCount {
		addBasicLandPermanent(g, game.Player1, types.Island)
	}
	for range libraryCount {
		addLibraryCard(g, game.Player1, wellOfLostDreamsScratchCard())
	}

	g.Turn.ActivePlayer = game.Player1
	emitEvent(g, game.Event{Kind: game.EventLifeGained, Player: game.Player1, Amount: lifeGained})
	if !engine.putTriggeredAbilitiesOnStack(g) {
		t.Fatal("Well of Lost Dreams life-gain trigger was not put on the stack")
	}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &TurnLog{})
	return g.Players[game.Player1].Hand.Size()
}

// TestWellOfLostDreamsBoundsDrawByLifeGained proves the amount of life gained
// caps the draw: gaining 3 life with ample mana (six Islands) and library still
// draws exactly three cards because the PayRepeatedly MaxCount reads the
// triggering life-change quantity, so the controller cannot pay a fourth {1}.
func TestWellOfLostDreamsBoundsDrawByLifeGained(t *testing.T) {
	if got := fireWellOfLostDreams(t, 3, 6, 6); got != 3 {
		t.Fatalf("hand size = %d, want 3 (draw bounded by the 3 life gained)", got)
	}
}

// TestWellOfLostDreamsBoundedByAvailableMana proves the paid X — and thus the
// draw — is also bounded below the life gained by the mana the controller can
// spend: gaining 5 life with only two Islands funds two {1} payments and draws
// two cards.
func TestWellOfLostDreamsBoundedByAvailableMana(t *testing.T) {
	if got := fireWellOfLostDreams(t, 5, 2, 6); got != 2 {
		t.Fatalf("hand size = %d, want 2 (draw bounded by two available mana)", got)
	}
}

// TestWellOfLostDreamsDrawsNothingWithoutMana proves the affirmative gate is
// fail-closed: with no mana the controller pays nothing, so the payment-succeeded
// gate skips the draw and the hand stays empty.
func TestWellOfLostDreamsDrawsNothingWithoutMana(t *testing.T) {
	if got := fireWellOfLostDreams(t, 3, 0, 6); got != 0 {
		t.Fatalf("hand size = %d, want 0 (no mana paid, no draw)", got)
	}
}
