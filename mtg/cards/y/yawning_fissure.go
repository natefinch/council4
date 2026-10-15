package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YawningFissure is the card definition for Yawning Fissure.
//
// Type: Sorcery
// Cost: {4}{R}
//
// Oracle text:
//
//	Each opponent sacrifices a land of their choice.
var YawningFissure = newYawningFissure

func newYawningFissure() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Yawning Fissure",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.SacrificePermanents{
							PlayerGroup: game.OpponentsReference(),
							Amount:      game.Fixed(1),
							Selection:   game.Selection{RequiredTypes: []types.Card{types.Land}},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Each opponent sacrifices a land of their choice.
		`,
		},
	}
}
