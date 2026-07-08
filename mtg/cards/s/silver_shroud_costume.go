package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SilverShroudCostume is the card definition for Silver Shroud Costume.
//
// Type: Artifact — Equipment
// Cost: {2}
//
// Oracle text:
//
//	Flash
//	When this Equipment enters, attach it to target creature you control. That creature gains shroud until end of turn. (It can't be the target of spells or abilities.)
//	Equipped creature can't be blocked.
//	Equip {3}
var SilverShroudCostume = newSilverShroudCostume

func newSilverShroudCostume() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Silver Shroud Costume",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectCantBeBlocked,
							AffectedAttached: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(3)}),
			},
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
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Attach{
									Attachment: game.EventPermanentReference(),
									Target:     game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Shroud,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			When this Equipment enters, attach it to target creature you control. That creature gains shroud until end of turn. (It can't be the target of spells or abilities.)
			Equipped creature can't be blocked.
			Equip {3}
		`,
		},
	}
}
