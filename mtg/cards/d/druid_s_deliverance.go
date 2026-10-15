package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DruidSDeliverance is the card definition for Druid's Deliverance.
//
// Type: Instant
// Cost: {1}{G}
//
// Oracle text:
//
//	Prevent all combat damage that would be dealt to you this turn. Populate. (Create a token that's a copy of a creature token you control.)
var DruidSDeliverance = newDruidSDeliverance

func newDruidSDeliverance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Druid's Deliverance",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.PreventDamage{
							Player:     game.ControllerReference(),
							All:        true,
							CombatOnly: true,
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source: game.TokenCopySourceChosenControlledCreatureToken,
							}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Prevent all combat damage that would be dealt to you this turn. Populate. (Create a token that's a copy of a creature token you control.)
		`,
		},
	}
}
