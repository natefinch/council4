package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BuilderSBlessing is the card definition for Builder's Blessing.
//
// Type: Enchantment
// Cost: {3}{W}
//
// Oracle text:
//
//	Untapped creatures you control get +0/+2.
var BuilderSBlessing = newBuilderSBlessing

func newBuilderSBlessing() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Builder's Blessing",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Untapped creatures you control get +0/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriFalse}),
							Layer:          game.LayerPowerToughnessModify,
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			Untapped creatures you control get +0/+2.
		`,
		},
	}
}
