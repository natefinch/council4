package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ArdenvalePaladin is the card definition for Ardenvale Paladin.
//
// Type: Creature — Human Knight
// Cost: {3}{W}
//
// Oracle text:
//
//	Adamant — If at least three white mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
var ArdenvalePaladin = newArdenvalePaladin()

func newArdenvalePaladin() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ardenvale Paladin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("Adamant — If at least three white mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.", &game.Condition{
					SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.White, Count: 3},
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Adamant — If at least three white mana was spent to cast this spell, this creature enters with a +1/+1 counter on it.
		`,
		},
	}
}
