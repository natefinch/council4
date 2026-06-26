package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HengeWalker is the card definition for Henge Walker.
//
// Type: Artifact Creature — Golem
// Cost: {3}
//
// Oracle text:
//
//	Adamant — If at least three mana of the same color was spent to cast this spell, this creature enters with a +1/+1 counter on it.
var HengeWalker = newHengeWalker()

func newHengeWalker() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Henge Walker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Adamant — If at least three mana of the same color was spent to cast this spell, this creature enters with a +1/+1 counter on it.", &game.Condition{
					SpellSameColorManaSpentAtLeast: 3,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Adamant — If at least three mana of the same color was spent to cast this spell, this creature enters with a +1/+1 counter on it.
		`,
		},
	}
}
