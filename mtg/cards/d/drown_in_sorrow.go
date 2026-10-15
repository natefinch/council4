package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DrownInSorrow is the card definition for Drown in Sorrow.
//
// Type: Sorcery
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	All creatures get -2/-2 until end of turn. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
var DrownInSorrow = newDrownInSorrow

func newDrownInSorrow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Drown in Sorrow",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     -2,
									ToughnessDelta: -2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Scry{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			All creatures get -2/-2 until end of turn. Scry 1. (Look at the top card of your library. You may put that card on the bottom.)
		`,
		},
	}
}
