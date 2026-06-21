package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

// TestSeaGateRestorationDrawsHandSizePlusOne resolves the Sea Gate Restoration
// draw clause: a Draw whose amount is the controller's hand size plus one. The
// controller starts with three cards in hand, so resolution must draw four.
func TestSeaGateRestorationDrawsHandSizePlusOne(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for i := range 3 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}})
	}
	for i := range 10 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('a' + i))}})
	}
	player := game.ControllerReference()
	addEffectSpellToStack(g, game.Player1, game.Draw{
		Amount: game.Dynamic(game.DynamicAmount{
			Kind:       game.DynamicAmountCountCardsInZone,
			Multiplier: 1,
			Addend:     1,
			Player:     &player,
			CardZone:   zone.Hand,
			Selection:  &game.Selection{},
		}),
		Player: game.ControllerReference(),
	}, nil)

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Hand.Size(); got != 7 {
		t.Fatalf("hand size = %d, want 7 (drew hand size 3 plus one = 4)", got)
	}
}

// TestSeaGateRestorationRemovesMaximumHandSize resolves the Sea Gate Restoration
// rest-of-game clause as a permanent ApplyRule that removes the controller's
// maximum hand size, then confirms the cleanup step no longer forces a discard.
func TestSeaGateRestorationRemovesMaximumHandSize(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := &game.StackObject{Controller: game.Player1}
	resolveInstruction(engine, g, obj, game.ApplyRule{
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectNoMaximumHandSize,
			AffectedPlayer: game.PlayerYou,
		}},
		Duration: game.DurationPermanent,
	}, &TurnLog{})

	if !playerHasNoMaximumHandSize(g, game.Player1) {
		t.Fatal("controller should have no maximum hand size after resolution")
	}

	for i := range 10 {
		addCardToHand(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: string(rune('A' + i))}})
	}

	engine.runEndingPhase(g, [game.NumPlayers]PlayerAgent{})

	if got := g.Players[game.Player1].Hand.Size(); got != 10 {
		t.Fatalf("hand size = %d, want 10 (no discard)", got)
	}
	if got := g.Players[game.Player1].Graveyard.Size(); got != 0 {
		t.Fatalf("graveyard size = %d, want 0", got)
	}
}
