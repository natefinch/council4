package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PincerSpider is the card definition for Pincer Spider.
//
// Type: Creature — Spider
// Cost: {2}{G}
//
// Oracle text:
//
//	Kicker {3} (You may pay an additional {3} as you cast this spell.)
//	Reach (This creature can block creatures with flying.)
//	If this creature was kicked, it enters with a +1/+1 counter on it.
var PincerSpider = newPincerSpider

func newPincerSpider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Pincer Spider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(3)}},
					},
				},
				game.ReachStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with a +1/+1 counter on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Kicker {3} (You may pay an additional {3} as you cast this spell.)
			Reach (This creature can block creatures with flying.)
			If this creature was kicked, it enters with a +1/+1 counter on it.
		`,
		},
	}
}
