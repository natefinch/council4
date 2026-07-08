package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IvoryCharm is the card definition for Ivory Charm.
//
// Type: Instant
// Cost: {W}
//
// Oracle text:
//
//	Choose one —
//	• All creatures get -2/-0 until end of turn.
//	• Tap target creature.
//	• Prevent the next 1 damage that would be dealt to any target this turn.
var IvoryCharm = newIvoryCharm

func newIvoryCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Ivory Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "All creatures get -2/-0 until end of turn.",
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:      game.LayerPowerToughnessModify,
											Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
											PowerDelta: -2,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Tap target creature.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Prevent the next 1 damage that would be dealt to any target this turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "any target",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(1),
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
			• All creatures get -2/-0 until end of turn.
			• Tap target creature.
			• Prevent the next 1 damage that would be dealt to any target this turn.
		`,
		},
	}
}
