package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
)

func drawDoublingPermanent(g *game.Game, controller game.PlayerID, multiplier int, exceptFirst bool) *game.Permanent {
	def := &game.CardDef{
		CardFace: game.CardFace{
			Name: "Doubler",
			ReplacementAbilities: []game.ReplacementAbility{
				game.DrawCardMultiplierReplacement("doubler", multiplier, exceptFirst),
			},
		},
	}
	permanent := addCombatPermanent(g, controller, def)
	registerPermanentReplacementEffects(g, permanent)
	return permanent
}

func TestDrawCardMultiplierDoublesSpellDraws(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 4 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	drawDoublingPermanent(g, game.Player1, 2, false)

	log := TurnLog{}
	engine.drawCards(g, game.Player1, 1, [game.NumPlayers]PlayerAgent{}, &log)

	if got := g.Players[game.Player1].Hand.Size(); got != 2 {
		t.Fatalf("hand size = %d, want 2 (one draw doubled)", got)
	}
}

func TestDrawCardMultiplierOnlyHelpsController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 4 {
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	drawDoublingPermanent(g, game.Player1, 2, false)

	log := TurnLog{}
	engine.drawCards(g, game.Player2, 1, [game.NumPlayers]PlayerAgent{}, &log)

	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("opponent hand size = %d, want 1 (controller-only replacement)", got)
	}
}

func TestDrawCardMultiplierExemptsFirstDrawStepDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 4 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	drawDoublingPermanent(g, game.Player1, 2, true)

	if got := drawCardMultiplier(g, game.Player1, true); got != 1 {
		t.Fatalf("first draw-step multiplier = %d, want 1 (exempt)", got)
	}
	if got := drawCardMultiplier(g, game.Player1, false); got != 2 {
		t.Fatalf("non-draw-step multiplier = %d, want 2", got)
	}
}

func TestDrawCardMultiplierPlainDoublesDrawStepDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	for range 4 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	drawDoublingPermanent(g, game.Player1, 2, false)

	if got := drawCardMultiplier(g, game.Player1, true); got != 2 {
		t.Fatalf("first draw-step multiplier = %d, want 2 (plain doubler is not exempt)", got)
	}
}
