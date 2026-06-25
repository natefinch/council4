package rules

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/id"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const milledCardsTestLink = "milled-cards"

// millPaidReturnInstructions models the resolved Ripples of Undeath ability:
// mill three cards (publishing exactly those cards as linked objects), then
// optionally pay {1} and 3 life, and if paid, return one of those milled cards
// from the graveyard to hand. The return is restricted to the linked milled
// cards and gated on the optional payment having succeeded.
func millPaidReturnInstructions() []game.Instruction {
	paid := opt.Val(game.InstructionResultGate{Key: "controller-paid", Succeeded: game.TriTrue})
	return []game.Instruction{
		{
			Primitive: game.Mill{
				Amount:        game.Fixed(3),
				Player:        game.ControllerReference(),
				PublishLinked: game.LinkedKey(milledCardsTestLink),
			},
		},
		{
			Primitive: game.Pay{Payment: game.ResolutionPayment{
				Prompt:   "Pay {1} and 3 life?",
				ManaCost: opt.Val(cost.Mana{cost.O(1)}),
				AdditionalCosts: []cost.Additional{
					{Kind: cost.AdditionalPayLife, Text: "3 life", Amount: 3},
				},
			}},
			PublishResult: "controller-paid",
		},
		{
			Primitive: game.ReturnFromGraveyardChoice(
				game.ControllerReference(),
				game.Selection{},
				game.Fixed(1),
				zone.Hand,
				false,
				opt.V[int]{},
				false,
				game.LinkedKey(milledCardsTestLink),
			),
			ResultGate: paid,
		},
	}
}

func milledCreatureDef(name string) *game.CardDef {
	return &game.CardDef{CardFace: game.CardFace{Name: name, Types: []types.Card{types.Creature}}}
}

// TestMillPaidReturnReturnsOnlyMilledCardWhenPaid proves that paying the
// optional combined mana+life cost returns one of exactly the three milled cards
// to hand, spends the mana and life, and never returns an unrelated graveyard
// card.
func TestMillPaidReturnReturnsOnlyMilledCardWhenPaid(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	oldGrave := addCardToGraveyard(g, game.Player1, milledCreatureDef("Old Grave"))
	milled := []id.ID{
		addCardToLibrary(g, game.Player1, milledCreatureDef("Milled One")),
		addCardToLibrary(g, game.Player1, milledCreatureDef("Milled Two")),
		addCardToLibrary(g, game.Player1, milledCreatureDef("Milled Three")),
	}
	addInstructionSpellToStack(g, millPaidReturnInstructions())

	engine.resolveTopOfStack(g, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 37 {
		t.Fatalf("controller life = %d, want 37 (paid 3 life)", got)
	}
	hand := g.Players[game.Player1].Hand
	if hand.Size() != 1 {
		t.Fatalf("hand size = %d, want 1 returned milled card", hand.Size())
	}
	returned := hand.All()[0]
	if !slices.Contains(milled, returned) {
		t.Fatalf("returned card %v was not one of the milled cards %v", returned, milled)
	}
	if hand.Contains(oldGrave) {
		t.Fatal("unrelated graveyard card was returned to hand")
	}
	if !g.Players[game.Player1].Graveyard.Contains(oldGrave) {
		t.Fatal("unrelated graveyard card left the graveyard")
	}
	graveMilled := 0
	for _, cardID := range milled {
		if g.Players[game.Player1].Graveyard.Contains(cardID) {
			graveMilled++
		}
	}
	if graveMilled != 2 {
		t.Fatalf("milled cards still in graveyard = %d, want 2 (one returned)", graveMilled)
	}
}

// TestMillPaidReturnDeclinedKeepsMilledCardsInGraveyard proves that declining
// the optional payment still mills three cards but returns none, leaving life
// and the graveyard untouched by the gated return.
func TestMillPaidReturnDeclinedKeepsMilledCardsInGraveyard(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	addBasicLandPermanent(g, game.Player1, types.Swamp)
	milled := []id.ID{
		addCardToLibrary(g, game.Player1, milledCreatureDef("Milled One")),
		addCardToLibrary(g, game.Player1, milledCreatureDef("Milled Two")),
		addCardToLibrary(g, game.Player1, milledCreatureDef("Milled Three")),
	}
	addInstructionSpellToStack(g, millPaidReturnInstructions())
	agents := [game.NumPlayers]PlayerAgent{game.Player1: optionalMayAgent{accept: false}}

	engine.resolveTopOfStackWithChoices(g, agents, &TurnLog{})

	if got := g.Players[game.Player1].Life; got != 40 {
		t.Fatalf("controller life = %d, want 40 (declined payment)", got)
	}
	if got := g.Players[game.Player1].Hand.Size(); got != 0 {
		t.Fatalf("hand size = %d, want 0 (declined return)", got)
	}
	for _, cardID := range milled {
		if !g.Players[game.Player1].Graveyard.Contains(cardID) {
			t.Fatalf("milled card %v missing from graveyard after declined payment", cardID)
		}
	}
}
