package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

func TestGoadedOpponentCreaturesCantBlock(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCombatPermanent(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Types: []types.Card{types.Creature},
		StaticAbilities: []game.StaticAbility{{RuleEffects: []game.RuleEffect{{
			Kind:               game.RuleEffectCantBlock,
			AffectedController: game.ControllerOpponent,
			PermanentTypes:     []types.Card{types.Creature},
			AffectedSelection:  game.Selection{MatchGoaded: true},
		}}}},
	}})
	opponent := addCombatCreaturePermanent(g, game.Player2)
	controller := addCombatCreaturePermanent(g, game.Player1)
	if ruleEffectProhibitsBlock(g, opponent) {
		t.Fatal("ungoaded opponent creature could not block")
	}
	goadPermanent(g, opponent, game.Player1, false)
	goadPermanent(g, controller, game.Player2, false)
	if !ruleEffectProhibitsBlock(g, opponent) {
		t.Fatal("goaded opponent creature could block")
	}
	if ruleEffectProhibitsBlock(g, controller) {
		t.Fatal("goaded controller creature could not block")
	}
}
