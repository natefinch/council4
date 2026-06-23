package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func spellCostIncrease(g *game.Game, caster game.PlayerID, card *game.CardDef) int {
	increase := 0
	for _, modifier := range staticCostModifiersForContext(g, caster, card, zone.Hand, nil) {
		increase += modifier.GenericIncrease
	}
	return increase
}

// TestOneShotSpellCostModifierIncreasesOpponentNoncreatureSpells proves the
// duration-bounded resolved cost modifier produced by Elspeth Conquers Death
// chapter II ("Noncreature spells your opponents cast cost {2} more to cast
// until your next turn.") raises an opponent's noncreature spell cost by {2},
// exempts creature spells, never touches the controller's own spells, and lapses
// once its until-your-next-turn duration expires.
func TestOneShotSpellCostModifierIncreasesOpponentNoncreatureSpells(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	g.RuleEffects = append(g.RuleEffects, game.RuleEffect{
		ID:                 g.IDGen.Next(),
		Kind:               game.RuleEffectCostModifier,
		Controller:         game.Player1,
		AffectedPlayer:     game.PlayerOpponent,
		ExcludedSpellTypes: []types.Card{types.Creature},
		CostModifier: game.CostModifier{
			Kind:            game.CostModifierSpell,
			GenericIncrease: 2,
		},
		Duration:    game.DurationUntilYourNextTurn,
		ExpiresFor:  game.Player1,
		CreatedTurn: g.Turn.TurnNumber,
	})

	instant := &game.CardDef{CardFace: game.CardFace{Name: "Test Bolt", Types: []types.Card{types.Instant}}}
	creature := &game.CardDef{CardFace: game.CardFace{Name: "Test Bear", Types: []types.Card{types.Creature}}}

	if got := spellCostIncrease(g, game.Player2, instant); got != 2 {
		t.Fatalf("opponent noncreature spell cost increase = {%d}, want {2}", got)
	}
	if got := spellCostIncrease(g, game.Player2, creature); got != 0 {
		t.Fatalf("opponent creature spell cost increase = {%d}, want {0} (creature spells are exempt)", got)
	}
	if got := spellCostIncrease(g, game.Player1, instant); got != 0 {
		t.Fatalf("controller's own noncreature spell cost increase = {%d}, want {0}", got)
	}

	// The effect is bound to its controller's next turn. Advancing to that turn
	// and running the turn-start cleanup must remove it so opponents cast at the
	// normal cost again.
	g.Turn.TurnNumber++
	g.Turn.ActivePlayer = game.Player1
	expireTurnStartDurations(g)
	if got := spellCostIncrease(g, game.Player2, instant); got != 0 {
		t.Fatalf("cost increase after expiry = {%d}, want {0} (the effect must lapse at the controller's next turn)", got)
	}
}
