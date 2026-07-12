package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestSourceAttachedCombatCounterpartCondition(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	equipment := addCombatPermanent(g, game.Player1, &game.CardDef{})
	equipped := addCombatCreaturePermanent(g, game.Player1)
	goblin := addCombatPermanent(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Types:    []types.Card{types.Creature},
		Subtypes: []types.Sub{types.Goblin},
	}})
	equipment.AttachedTo = opt.Val(equipped.ObjectID)
	g.Combat = &game.CombatState{Blockers: []game.BlockDeclaration{{
		Blocker:  equipped.ObjectID,
		Blocking: goblin.ObjectID,
	}}}
	condition := opt.Val(game.Condition{
		SourceAttachedCombatCounterpartSubtypes: [2]types.Sub{types.Goblin, types.Orc},
	})
	if !conditionSatisfied(g, conditionContext{controller: game.Player1, source: equipment}, condition) {
		t.Fatal("condition did not match equipped creature blocking a Goblin")
	}
	card, _ := g.GetCardInstance(goblin.CardInstanceID)
	card.Def.Subtypes = nil
	if conditionSatisfied(g, conditionContext{controller: game.Player1, source: equipment}, condition) {
		t.Fatal("condition matched a blocker without a named subtype")
	}
}
