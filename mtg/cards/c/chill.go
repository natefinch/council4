package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Chill is the card definition for Chill.
//
// Type: Enchantment
// Cost: {1}{U}
//
// Oracle text:
//
//	Red spells cost {2} more to cast.
var Chill = newChill()

func newChill() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Chill",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind: game.RuleEffectCostModifier,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								CardSelection:   game.Selection{ColorsAny: []color.Color{color.Red}},
								GenericIncrease: 2,
							},
						},
					},
				},
			},
			OracleText: `
			Red spells cost {2} more to cast.
		`,
		},
	}
}
