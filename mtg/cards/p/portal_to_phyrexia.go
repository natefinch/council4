package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PortalToPhyrexia is the card definition for Portal to Phyrexia.
//
// Type: Artifact
// Cost: {9}
//
// Oracle text:
//
//	When this artifact enters, each opponent sacrifices three creatures of their choice.
//	At the beginning of your upkeep, put target creature card from a graveyard onto the battlefield under your control. It's a Phyrexian in addition to its other types.
var PortalToPhyrexia = newPortalToPhyrexia

func newPortalToPhyrexia() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Portal to Phyrexia",
			ManaCost: opt.Val(cost.Mana{
				cost.O(9),
			}),
			Types: []types.Card{types.Artifact},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:      game.Fixed(3),
									PlayerGroup: game.OpponentsReference(),
									Selection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card from a graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									Recipient:     opt.Val(game.ControllerReference()),
									PublishLinked: game.LinkedKey("leave-bf-exile-1"),
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.LinkedObjectReference("leave-bf-exile-1")),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:       game.LayerType,
											AddSubtypes: []types.Sub{types.Sub("Phyrexian")},
										},
									},
									Duration: game.DurationPermanent,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this artifact enters, each opponent sacrifices three creatures of their choice.
			At the beginning of your upkeep, put target creature card from a graveyard onto the battlefield under your control. It's a Phyrexian in addition to its other types.
		`,
		},
	}
}
