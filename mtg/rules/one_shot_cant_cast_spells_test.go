package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// TestOneShotCantCastSpellsProhibitsOpponents proves the one-shot, turn-scoped
// cast prohibition produced by Silence ("Your opponents can't cast spells this
// turn.") stops the caster's opponents from casting spells while leaving the
// caster unaffected, on any player's turn (the restriction is not scoped to the
// controller's turn; its this-turn duration handles expiry).
func TestOneShotCantCastSpellsProhibitsOpponents(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCantCastSpells,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerOpponent,
		Duration:       game.DurationThisTurn,
		CreatedTurn:    g.Turn.TurnNumber,
	})
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Test Bolt", Types: []types.Card{types.Instant}}}

	g.Turn.ActivePlayer = game.Player2
	if !spellCastProhibited(g, game.Player2, spell) {
		t.Fatal("opponent should be unable to cast spells after Silence resolves")
	}
	if spellCastProhibited(g, game.Player1, spell) {
		t.Fatal("the caster of Silence is never restricted by it")
	}
}

// TestOneShotCantCastSpellsAllPlayers proves the all-players form ("Players
// can't cast spells this turn.") restricts every player, including the caster.
func TestOneShotCantCastSpellsAllPlayers(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCantCastSpells,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerAny,
		Duration:       game.DurationThisTurn,
		CreatedTurn:    g.Turn.TurnNumber,
	})
	spell := &game.CardDef{CardFace: game.CardFace{Name: "Test Bolt", Types: []types.Card{types.Instant}}}

	if !spellCastProhibited(g, game.Player1, spell) {
		t.Fatal("the all-players prohibition must restrict the caster too")
	}
	if !spellCastProhibited(g, game.Player2, spell) {
		t.Fatal("the all-players prohibition must restrict opponents too")
	}
}
