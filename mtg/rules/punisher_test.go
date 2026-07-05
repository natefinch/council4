package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
)

func punisherStackObject(g *game.Game) *game.StackObject {
	source := addCreaturePermanent(g, game.Player1)
	return &game.StackObject{
		ID:         g.IDGen.Next(),
		Controller: game.Player1,
		SourceID:   source.ObjectID,
	}
}

// TestPunisherEachLoseLifeNoAlternativeLosesLife proves that when the affected
// opponent can neither sacrifice nor discard, the punisher choice falls through
// to the life loss, as Hag of Ceaseless Torment relies on.
func TestPunisherEachLoseLifeNoAlternativeLosesLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := punisherStackObject(g)

	resolveInstruction(engine, g, obj, game.PunisherEachLoseLife{
		PlayerGroup:        game.OpponentsReference(),
		Amount:             game.Fixed(3),
		AllowSacrifice:     true,
		SacrificeSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
		AllowDiscard:       true,
	}, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 37 {
		t.Fatalf("Player2 life = %d, want 37 (lost 3)", got)
	}
}

// TestPunisherEachLoseLifeDiscardCountDiscardsMultipleCards proves that a
// punisher whose discard alternative requires more than one card (Court of
// Ambition's monarch escalation, "unless they discard two cards") makes the
// affected player discard exactly that many cards to avoid the life loss.
func TestPunisherEachLoseLifeDiscardCountDiscardsMultipleCards(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := punisherStackObject(g)
	first := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Card A"}})
	second := addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Card B"}})

	agents := [game.NumPlayers]PlayerAgent{
		// First choice: pick the discard alternative (index 1, after "Lose
		// life"). Second choice: discard both required cards.
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}, {0, 1}}},
	}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.PunisherEachLoseLife{
			PlayerGroup:  game.OpponentsReference(),
			Amount:       game.Fixed(6),
			AllowDiscard: true,
			DiscardCount: 2,
		},
	}, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("Player2 life = %d, want 40 (discarded two cards instead)", got)
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 0 {
		t.Fatalf("Player2 hand size = %d, want 0 (both cards discarded)", got)
	}
	for _, cardID := range []id.ID{first, second} {
		if g.Players[game.Player2].Hand.Contains(cardID) {
			t.Fatalf("card %v still in hand, want discarded", cardID)
		}
	}
}

// TestPunisherEachLoseLifeDiscardCountInsufficientHandLosesLife proves that a
// player who cannot discard the full required count is not offered the discard
// alternative and therefore loses the life.
func TestPunisherEachLoseLifeDiscardCountInsufficientHandLosesLife(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := punisherStackObject(g)
	addCardToHand(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Only Card"}})

	resolveInstruction(engine, g, obj, game.PunisherEachLoseLife{
		PlayerGroup:  game.OpponentsReference(),
		Amount:       game.Fixed(6),
		AllowDiscard: true,
		DiscardCount: 2,
	}, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 34 {
		t.Fatalf("Player2 life = %d, want 34 (lost 6, could not discard two)", got)
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("Player2 hand size = %d, want 1 (no cards discarded)", got)
	}
}

// chooses to sacrifice a permanent avoids the life loss, and the chosen
// permanent leaves the battlefield.
func TestPunisherEachLoseLifeSacrificeAvoidsLifeLoss(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	obj := punisherStackObject(g)
	victim := addCreaturePermanent(g, game.Player2)

	agents := [game.NumPlayers]PlayerAgent{
		game.Player2: &choiceOnlyAgent{choices: [][]int{{1}}},
	}
	engine.resolveInstructionWithChoices(g, obj, &game.Instruction{
		Primitive: game.PunisherEachLoseLife{
			PlayerGroup:        game.OpponentsReference(),
			Amount:             game.Fixed(3),
			AllowSacrifice:     true,
			SacrificeSelection: game.Selection{ExcludedTypes: []types.Card{types.Land}},
			AllowDiscard:       true,
		},
	}, agents, &TurnLog{})

	if got := g.Players[game.Player2].Life; got != 40 {
		t.Fatalf("Player2 life = %d, want 40 (sacrificed instead)", got)
	}
	for _, permanent := range g.Battlefield {
		if permanent.ObjectID == victim.ObjectID {
			t.Fatal("victim permanent still on battlefield, want sacrificed")
		}
	}
}
