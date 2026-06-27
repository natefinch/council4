package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Squeeze is the card definition for Squeeze.
//
// Type: Enchantment
// Cost: {3}{U}
//
// Oracle text:
//
//	Sorcery spells cost {3} more to cast.
var Squeeze = newSqueeze()

func newSqueeze() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Squeeze",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
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
								CardSelection:   game.Selection{RequiredTypes: []types.Card{types.Sorcery}},
								GenericIncrease: 3,
							},
						},
					},
				},
			},
			OracleText: `
			Sorcery spells cost {3} more to cast.
		`,
		},
	}
}
