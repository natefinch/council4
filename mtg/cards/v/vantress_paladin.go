package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VantressPaladin is the card definition for Vantress Paladin.
//
// Type: Creature — Human Knight
// Cost: {3}{U}
//
// Oracle text:
//
//	Flying
//	Adamant — If at least three blue mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
var VantressPaladin = newVantressPaladin()

func newVantressPaladin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vantress Paladin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Adamant — If at least three blue mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.", &game.Condition{
					SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Blue, Count: 3},
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Flying
			Adamant — If at least three blue mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
		`,
		},
	}
}
