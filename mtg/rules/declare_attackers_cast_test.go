package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestDeclareAttackersCastRestriction(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	card := &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Instant},
		StaticAbilities: []game.StaticAbility{{
			CastOnlyAfterAttackedThisStep: true,
		}},
	}}
	g.Turn.Step = game.StepDeclareAttackers
	g.Combat = &game.CombatState{PlayersAttacked: map[game.PlayerID]bool{}}
	if canCastAtCurrentTiming(g, game.Player2, card) {
		t.Fatal("spell was cast before its controller was attacked")
	}
	g.Combat.PlayersAttacked[game.Player2] = true
	if !canCastAtCurrentTiming(g, game.Player2, card) {
		t.Fatal("spell was not cast after its controller was attacked")
	}
	g.Turn.Step = game.StepDeclareBlockers
	if canCastAtCurrentTiming(g, game.Player2, card) {
		t.Fatal("spell was cast outside the declare attackers step")
	}
	clone := g.Clone()
	g.Combat.PlayersAttacked[game.Player2] = false
	if !clone.Combat.PlayersAttacked[game.Player2] {
		t.Fatal("combat attack history was not cloned independently")
	}
}
