package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Bedlam is the card definition for Bedlam.
//
// Type: Enchantment
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	Creatures can't block.
var Bedlam = newBedlam

func newBedlam() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Bedlam",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures can't block.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							Duration:       game.DurationPermanent,
							PermanentTypes: []types.Card{types.Creature},
						},
					},
				},
			},
			OracleText: `
			Creatures can't block.
		`,
		},
	}
}
