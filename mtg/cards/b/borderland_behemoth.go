package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BorderlandBehemoth is the card definition for Borderland Behemoth.
//
// Type: Creature — Giant Warrior
// Cost: {5}{R}{R}
//
// Oracle text:
//
//	Trample
//	This creature gets +4/+4 for each other Giant you control.
var BorderlandBehemoth = newBorderlandBehemoth

func newBorderlandBehemoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Borderland Behemoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 4,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Giant")}, Controller: game.ControllerYou, ExcludeSource: true}),
							}),
							ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 4,
								Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Giant")}, Controller: game.ControllerYou, ExcludeSource: true}),
							}),
						},
					},
				},
			},
			OracleText: `
			Trample
			This creature gets +4/+4 for each other Giant you control.
		`,
		},
	}
}
