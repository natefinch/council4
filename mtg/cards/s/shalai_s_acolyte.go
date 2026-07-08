package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShalaiSAcolyte is the card definition for Shalai's Acolyte.
//
// Type: Creature — Angel
// Cost: {4}{W}
//
// Oracle text:
//
//	Kicker {1}{G} (You may pay an additional {1}{G} as you cast this spell.)
//	Flying
//	If this creature was kicked, it enters with two +1/+1 counters on it.
var ShalaiSAcolyte = newShalaiSAcolyte

func newShalaiSAcolyte() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Shalai's Acolyte",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(1), cost.G}},
					},
				},
				game.FlyingStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersIfReplacement("If this creature was kicked, it enters with two +1/+1 counters on it.", &game.Condition{
					EventPermanentWasKicked: true,
				}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 2}),
			},
			OracleText: `
			Kicker {1}{G} (You may pay an additional {1}{G} as you cast this spell.)
			Flying
			If this creature was kicked, it enters with two +1/+1 counters on it.
		`,
		},
	}
}
