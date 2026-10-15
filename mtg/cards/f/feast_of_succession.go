package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FeastOfSuccession is the card definition for Feast of Succession.
//
// Type: Sorcery
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	All creatures get -4/-4 until end of turn. You become the monarch.
var FeastOfSuccession = newFeastOfSuccession

func newFeastOfSuccession() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Feast of Succession",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
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
									PowerDelta:     -4,
									ToughnessDelta: -4,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.BecomeMonarch{
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			All creatures get -4/-4 until end of turn. You become the monarch.
		`,
		},
	}
}
