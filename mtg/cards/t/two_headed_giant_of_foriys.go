package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TwoHeadedGiantOfForiys is the card definition for Two-Headed Giant of Foriys.
//
// Type: Creature — Giant
// Cost: {4}{R}
//
// Oracle text:
//
//	Trample
//	This creature can block an additional creature each combat.
var TwoHeadedGiantOfForiys = newTwoHeadedGiantOfForiys

func newTwoHeadedGiantOfForiys() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Two-Headed Giant of Foriys",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedSource:       true,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			OracleText: `
			Trample
			This creature can block an additional creature each combat.
		`,
		},
	}
}
