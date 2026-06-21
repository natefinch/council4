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

// TestOneShotCantCastNoncreatureSpells proves the noncreature filter produced by
// Ranger-Captain of Eos ("Your opponents can't cast noncreature spells this
// turn.") stops opponents from casting noncreature spells while still letting
// them cast creature spells, and lapses when the this-turn duration expires.
func TestOneShotCantCastNoncreatureSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:                 g.IDGen.Next(),
		Kind:               game.RuleEffectCantCastSpells,
		Controller:         game.Player1,
		AffectedPlayer:     game.PlayerOpponent,
		ExcludedSpellTypes: []types.Card{types.Creature},
		Duration:           game.DurationThisTurn,
		CreatedTurn:        g.Turn.TurnNumber,
	})
	instant := &game.CardDef{CardFace: game.CardFace{Name: "Test Bolt", Types: []types.Card{types.Instant}}}
	creature := &game.CardDef{CardFace: game.CardFace{Name: "Test Bear", Types: []types.Card{types.Creature}}}
	artifactCreature := &game.CardDef{CardFace: game.CardFace{Name: "Test Myr", Types: []types.Card{types.Artifact, types.Creature}}}

	g.Turn.ActivePlayer = game.Player2
	if !spellCastProhibited(g, game.Player2, instant) {
		t.Fatal("opponent should be unable to cast a noncreature spell")
	}
	if spellCastProhibited(g, game.Player2, creature) {
		t.Fatal("opponent should still be able to cast a creature spell")
	}
	if spellCastProhibited(g, game.Player2, artifactCreature) {
		t.Fatal("opponent should still be able to cast an artifact creature spell")
	}
	if spellCastProhibited(g, game.Player1, instant) {
		t.Fatal("the caster is never restricted by its own prohibition")
	}

	// The restriction is a this-turn effect; clearing it (as end-of-turn cleanup
	// does) lets opponents cast noncreature spells again next turn.
	g.RuleEffects = g.RuleEffects[:0]
	if spellCastProhibited(g, game.Player2, instant) {
		t.Fatal("the prohibition must lapse once its this-turn duration expires")
	}
}

// TestOneShotCantCastCreatureSpells proves the positive creature filter ("can't
// cast creature spells this turn.") restricts only creature spells, leaving
// noncreature spells castable.
func TestOneShotCantCastCreatureSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:             g.IDGen.Next(),
		Kind:           game.RuleEffectCantCastSpells,
		Controller:     game.Player1,
		AffectedPlayer: game.PlayerOpponent,
		SpellTypes:     []types.Card{types.Creature},
		Duration:       game.DurationThisTurn,
		CreatedTurn:    g.Turn.TurnNumber,
	})
	instant := &game.CardDef{CardFace: game.CardFace{Name: "Test Bolt", Types: []types.Card{types.Instant}}}
	creature := &game.CardDef{CardFace: game.CardFace{Name: "Test Bear", Types: []types.Card{types.Creature}}}

	g.Turn.ActivePlayer = game.Player2
	if !spellCastProhibited(g, game.Player2, creature) {
		t.Fatal("opponent should be unable to cast a creature spell")
	}
	if spellCastProhibited(g, game.Player2, instant) {
		t.Fatal("opponent should still be able to cast a noncreature spell")
	}
}
