package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ReturnOfTheWildspeaker is the card definition for Return of the Wildspeaker.
//
// Type: Instant
// Cost: {4}{G}
//
// Oracle text:
//
//	Choose one —
//	• Draw cards equal to the greatest power among non-Human creatures you control.
//	• Non-Human creatures you control get +3/+3 until end of turn.
var ReturnOfTheWildspeaker = newReturnOfTheWildspeaker

func newReturnOfTheWildspeaker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Return of the Wildspeaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Draw cards equal to the greatest power among non-Human creatures you control.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountGreatestPowerInGroup,
										Multiplier: 1,
										Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Human"), Controller: game.ControllerYou}),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Non-Human creatures you control get +3/+3 until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:       game.DurationPermanent,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Human"), Controller: game.ControllerYou}),
											Layer:          game.LayerPowerToughnessModify,
											PowerDelta:     3,
											ToughnessDelta: 3,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Draw cards equal to the greatest power among non-Human creatures you control.
			• Non-Human creatures you control get +3/+3 until end of turn.
		`,
		},
	}
}
