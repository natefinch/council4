package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LocthwainPaladin is the card definition for Locthwain Paladin.
//
// Type: Creature — Human Knight
// Cost: {3}{B}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	Adamant — If at least three black mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
var LocthwainPaladin = newLocthwainPaladin

func newLocthwainPaladin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Locthwain Paladin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Adamant — If at least three black mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.", &game.Condition{
					SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Black, Count: 3},
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			Adamant — If at least three black mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
		`,
		},
	}
}
