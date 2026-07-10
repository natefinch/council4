package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KolaghanTheStormSFury is the card definition for Kolaghan, the Storm's Fury.
//
// Type: Legendary Creature — Dragon
// Cost: {3}{B}{R}
//
// Oracle text:
//
//	Flying
//	Whenever a Dragon you control attacks, creatures you control get +1/+0 until end of turn.
//	Dash {3}{B}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
var KolaghanTheStormSFury = newKolaghanTheStormSFury

func newKolaghanTheStormSFury() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Kolaghan, the Storm's Fury",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Dragon},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventAttackerDeclared,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Dragon")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
											PowerDelta: 1,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.DashTriggeredAbility(),
			},
			AlternativeCosts: []cost.Alternative{
				cost.Alternative{
					Label:    "Dash",
					ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B, cost.R}),
					Mechanic: cost.AlternativeMechanicDash,
				},
			},
			OracleText: `
			Flying
			Whenever a Dragon you control attacks, creatures you control get +1/+0 until end of turn.
			Dash {3}{B}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)
		`,
		},
	}
}
