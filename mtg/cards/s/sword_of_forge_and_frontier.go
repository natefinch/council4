package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SwordOfForgeAndFrontier is the card definition for Sword of Forge and Frontier.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature gets +2/+2 and has protection from red and from green.
//	Whenever equipped creature deals combat damage to a player, exile the top two cards of your library. You may play those cards this turn. You may play an additional land this turn.
//	Equip {2}
var SwordOfForgeAndFrontier = newSwordOfForgeAndFrontier

func newSwordOfForgeAndFrontier() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sword of Forge and Frontier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ProtectionFromColorsStaticAbility(color.Red, color.Green)),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(2)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Source:                game.TriggerSourceAttachedPermanent,
							Subject:               game.TriggerSubjectDamageSource,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ImpulseExile{
									Player:   game.ControllerReference(),
									Amount:   game.Fixed(2),
									Duration: game.DurationThisTurn,
								},
							},
							{
								Primitive: game.ApplyRule{
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind:                game.RuleEffectAdditionalLandPlays,
											AffectedPlayer:      game.PlayerYou,
											AdditionalLandPlays: 1,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Equipped creature gets +2/+2 and has protection from red and from green.
			Whenever equipped creature deals combat damage to a player, exile the top two cards of your library. You may play those cards this turn. You may play an additional land this turn.
			Equip {2}
		`,
		},
	}
}
