package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/id"
)

// resolveReturnExiledCardsWithCounter resolves a ReturnExiledCardsWithCounter
// ability controlled by Player1. The primitive needs no resolution-time choice,
// so no agents supply input.
func resolveReturnExiledCardsWithCounter(t *testing.T, g *game.Game, prim game.ReturnExiledCardsWithCounter) *TurnLog {
	t.Helper()
	engine := NewEngine(nil)
	addEffectSpellToStack(g, game.Player1, prim, nil)
	log := TurnLog{}
	engine.resolveTopOfStackWithChoices(g, [game.NumPlayers]PlayerAgent{}, &log)
	return &log
}

func exileCardWithCounter(g *game.Game, owner game.PlayerID, name string, kind counter.Kind) id.ID {
	cardID := addCardToExile(g, owner, &game.CardDef{CardFace: game.CardFace{Name: name}})
	g.AddExileCounter(cardID, kind, 1)
	return cardID
}

// TestReturnExiledCardsWithCounterReturnsOwnedCountered verifies the mass return
// moves every card the controller owns in exile that bears the named marker
// counter to the controller's hand, while cards the controller owns without that
// counter stay in exile (Flamewar, Brash Veteran: "Put all exiled cards you own
// with intel counters on them into your hand.").
func TestReturnExiledCardsWithCounterReturnsOwnedCountered(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	intelA := exileCardWithCounter(g, game.Player1, "Own Intel A", counter.Intel)
	intelB := exileCardWithCounter(g, game.Player1, "Own Intel B", counter.Intel)
	// A card the controller owns with no intel counter stays in exile.
	plain := addCardToExile(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Own Plain"}})

	resolveReturnExiledCardsWithCounter(t, g, game.ReturnExiledCardsWithCounter{
		Player:  game.ControllerReference(),
		Counter: counter.Intel,
	})

	for _, cardID := range []id.ID{intelA, intelB} {
		if !g.Players[game.Player1].Hand.Contains(cardID) {
			t.Fatalf("intel-countered card %v was not returned to the owner's hand", cardID)
		}
		if g.Players[game.Player1].Exile.Contains(cardID) {
			t.Fatalf("intel-countered card %v remained in exile after return", cardID)
		}
		if g.HasExileCounter(cardID, counter.Intel) {
			t.Fatalf("intel-countered card %v kept its exile counter after leaving exile", cardID)
		}
	}
	if !g.Players[game.Player1].Exile.Contains(plain) {
		t.Fatal("card without an intel counter must remain in exile")
	}
	if g.Players[game.Player1].Hand.Contains(plain) {
		t.Fatal("card without an intel counter must not be returned to hand")
	}
}

// TestReturnExiledCardsWithCounterIgnoresOtherOwners verifies the mass return
// only touches cards the controller owns: an intel-countered card another player
// owns stays in that player's exile, matching "cards you own."
func TestReturnExiledCardsWithCounterIgnoresOtherOwners(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	ownIntel := exileCardWithCounter(g, game.Player1, "Own Intel", counter.Intel)
	// Player2 owns an intel-countered card in their own exile zone.
	oppIntel := exileCardWithCounter(g, game.Player2, "Opp Intel", counter.Intel)

	resolveReturnExiledCardsWithCounter(t, g, game.ReturnExiledCardsWithCounter{
		Player:  game.ControllerReference(),
		Counter: counter.Intel,
	})

	if !g.Players[game.Player1].Hand.Contains(ownIntel) {
		t.Fatal("controller-owned intel-countered card should be returned to hand")
	}
	if !g.Players[game.Player2].Exile.Contains(oppIntel) {
		t.Fatal("another player's intel-countered card must stay in their exile")
	}
	if g.Players[game.Player1].Hand.Contains(oppIntel) {
		t.Fatal("another player's card must not be returned to the controller's hand")
	}
}

// TestReturnExiledCardsWithCounterMatchesOnlyNamedCounter verifies the filter is
// specific to the primitive's counter kind: an exiled card the controller owns
// bearing a different named marker counter is left in exile.
func TestReturnExiledCardsWithCounterMatchesOnlyNamedCounter(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})

	intel := exileCardWithCounter(g, game.Player1, "Own Intel", counter.Intel)
	// Same owner, but a void counter rather than intel.
	void := exileCardWithCounter(g, game.Player1, "Own Void", counter.Void)

	resolveReturnExiledCardsWithCounter(t, g, game.ReturnExiledCardsWithCounter{
		Player:  game.ControllerReference(),
		Counter: counter.Intel,
	})

	if !g.Players[game.Player1].Hand.Contains(intel) {
		t.Fatal("intel-countered card should be returned to hand")
	}
	if !g.Players[game.Player1].Exile.Contains(void) {
		t.Fatal("void-countered card must not be returned by an intel-counter filter")
	}
}
