package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FatedRetribution is the card definition for Fated Retribution.
//
// Type: Instant
// Cost: {4}{W}{W}{W}
//
// Oracle text:
//
//	Destroy all creatures and planeswalkers. If it's your turn, scry 2.
var FatedRetribution = newFatedRetribution

func newFatedRetribution() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Fated Retribution",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Destroy{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SourceControllerTurn: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Destroy all creatures and planeswalkers. If it's your turn, scry 2.
		`,
		},
	}
}
