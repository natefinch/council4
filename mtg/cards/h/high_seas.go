package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HighSeas is the card definition for High Seas.
//
// Type: Enchantment
// Cost: {2}{U}
//
// Oracle text:
//
//	Red creature spells and green creature spells cost {1} more to cast.
var HighSeas = newHighSeas

func newHighSeas() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "High Seas",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Red creature spells and green creature spells cost {1} more to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:     game.RuleEffectCostModifier,
							Duration: game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								GenericIncrease: 1,
								CardSelection:   game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Red, color.Green}},
							},
						},
					},
				},
			},
			OracleText: `
			Red creature spells and green creature spells cost {1} more to cast.
		`,
		},
	}
}
