package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KavuAggressor is the card definition for Kavu Aggressor.
//
// Type: Creature — Kavu
// Cost: {2}{R}
//
// Oracle text:
//
//	Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	This creature can't block.
//	If this creature was kicked, it enters with a +1/+1 counter on it.
var KavuAggressor = newKavuAggressor

func newKavuAggressor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Kavu Aggressor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kavu},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(4)}},
					},
				},
				game.CantBlockStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with a +1/+1 counter on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Kicker {4} (You may pay an additional {4} as you cast this spell.)
			This creature can't block.
			If this creature was kicked, it enters with a +1/+1 counter on it.
		`,
		},
	}
}
