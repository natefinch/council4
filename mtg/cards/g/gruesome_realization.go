package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GruesomeRealization is the card definition for Gruesome Realization.
//
// Type: Sorcery
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Choose one —
//	• You draw two cards and you lose 2 life.
//	• Creatures your opponents control get -1/-1 until end of turn.
var GruesomeRealization = newGruesomeRealization

func newGruesomeRealization() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gruesome Realization",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "You draw two cards and you lose 2 life.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Creatures your opponents control get -1/-1 until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Duration:       game.DurationPermanent,
											Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
											Layer:          game.LayerPowerToughnessModify,
											PowerDelta:     -1,
											ToughnessDelta: -1,
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
			• You draw two cards and you lose 2 life.
			• Creatures your opponents control get -1/-1 until end of turn.
		`,
		},
	}
}
