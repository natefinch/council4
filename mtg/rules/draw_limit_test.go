package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// drawLimitPermanent gives controller a battlefield permanent whose continuous
// static ability caps how many cards the affected players may draw each turn
// (Narset, Parter of Veils; Spirit of the Labyrinth).
func drawLimitPermanent(g *game.Game, controller game.PlayerID, affected game.PlayerRelation, limit int) *game.Permanent {
	return addCombatPermanent(g, controller, &game.CardDef{CardFace: game.CardFace{
		Name:  "Drawcap",
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{
			RuleEffects: []game.RuleEffect{{
				Kind:             game.RuleEffectDrawLimitPerTurn,
				AffectedPlayer:   affected,
				DrawLimitPerTurn: limit,
			}},
		}},
	}})
}

func TestDrawLimitReplacesOpponentSecondDraw(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 4 {
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	drawLimitPermanent(g, game.Player1, game.PlayerOpponent, 1)

	log := TurnLog{}
	if !engine.drawCardWithReplacements(g, game.Player2, [game.NumPlayers]PlayerAgent{}, &log, false) {
		t.Fatal("first opponent draw should succeed")
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("hand size after first draw = %d, want 1", got)
	}
	if engine.drawCardWithReplacements(g, game.Player2, [game.NumPlayers]PlayerAgent{}, &log, false) {
		t.Fatal("second opponent draw should be replaced by drawing nothing")
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("hand size after over-limit draw = %d, want 1 (no extra card)", got)
	}
}

func TestDrawLimitDoesNotAffectController(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 4 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	drawLimitPermanent(g, game.Player1, game.PlayerOpponent, 1)

	log := TurnLog{}
	engine.drawCards(g, game.Player1, 3, [game.NumPlayers]PlayerAgent{}, &log)
	if got := g.Players[game.Player1].Hand.Size(); got != 3 {
		t.Fatalf("controller hand size = %d, want 3 (an opponent-only limit never restricts the controller)", got)
	}
}

func TestDrawLimitEachPlayerRestrictsEveryone(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	engine := NewEngine(nil)
	for range 4 {
		addCardToLibrary(g, game.Player1, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
		addCardToLibrary(g, game.Player2, &game.CardDef{CardFace: game.CardFace{Name: "Card"}})
	}
	drawLimitPermanent(g, game.Player1, game.PlayerAny, 1)

	log := TurnLog{}
	engine.drawCards(g, game.Player1, 3, [game.NumPlayers]PlayerAgent{}, &log)
	engine.drawCards(g, game.Player2, 3, [game.NumPlayers]PlayerAgent{}, &log)
	if got := g.Players[game.Player1].Hand.Size(); got != 1 {
		t.Fatalf("controller hand size = %d, want 1 (each player limited to one)", got)
	}
	if got := g.Players[game.Player2].Hand.Size(); got != 1 {
		t.Fatalf("opponent hand size = %d, want 1 (each player limited to one)", got)
	}
}
