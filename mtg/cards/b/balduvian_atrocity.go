package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BalduvianAtrocity is the card definition for Balduvian Atrocity.
//
// Type: Creature — Phyrexian Berserker
// Cost: {2}{B}
//
// Oracle text:
//
//	Kicker {R} (You may pay an additional {R} as you cast this spell.)
//	Menace
//	When this creature enters, if it was kicked, return target creature card with mana value 3 or less from your graveyard to the battlefield. It gains haste. Sacrifice it at the beginning of the next end step.
var BalduvianAtrocity = newBalduvianAtrocity

func newBalduvianAtrocity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Balduvian Atrocity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Berserker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.R}},
					},
				},
				game.MenaceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf:                        "if it was kicked",
						InterveningIfEventPermanentWasKicked: true,
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card with mana value 3 or less from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou, ManaValue: opt.Val(compare.Int{Op: compare.LessOrEqual, Value: 3})}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.PutOnBattlefield{
									Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
									PublishLinked: game.LinkedKey("gain-keyword-1"),
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.LinkedObjectReference("gain-keyword-1")),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Haste,
											},
										},
									},
									Duration:      game.DurationPermanent,
									PublishLinked: game.LinkedKey("delayed-sacrifice-2"),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing:         game.DelayedAtBeginningOfNextEndStep,
										CapturedObject: opt.Val(game.LinkedObjectReference("delayed-sacrifice-2")),
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Sacrifice{
														Object: game.CapturedObjectReference(),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Kicker {R} (You may pay an additional {R} as you cast this spell.)
			Menace
			When this creature enters, if it was kicked, return target creature card with mana value 3 or less from your graveyard to the battlefield. It gains haste. Sacrifice it at the beginning of the next end step.
		`,
		},
	}
}
