package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmberethPaladin is the card definition for Embereth Paladin.
//
// Type: Creature — Human Knight
// Cost: {3}{R}
//
// Oracle text:
//
//	Haste
//	Adamant — If at least three red mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
var EmberethPaladin = newEmberethPaladin

func newEmberethPaladin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Embereth Paladin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Adamant — If at least three red mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.", &game.Condition{
					SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Red, Count: 3},
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Haste
			Adamant — If at least three red mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
		`,
		},
	}
}
