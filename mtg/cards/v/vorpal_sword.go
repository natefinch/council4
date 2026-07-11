package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// VorpalSword is the card definition for Vorpal Sword.
//
// Type: Artifact — Equipment
// Cost: {B}
//
// Oracle text:
//
//	Equipped creature gets +2/+0 and has deathtouch.
//	{5}{B}{B}{B}: Until end of turn, this Equipment gains "Whenever equipped creature deals combat damage to a player, that player loses the game."
//	Equip {B}{B}
var VorpalSword = newVorpalSword

func newVorpalSword() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vorpal Sword",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:   []color.Color{color.Black},
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:      game.LayerPowerToughnessModify,
							Group:      game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta: 2,
						},
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddKeywords: []game.Keyword{
								game.Deathtouch,
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{5}{B}{B}{B}: Until end of turn, this Equipment gains \"Whenever equipped creature deals combat damage to a player, that player loses the game.\"",
					ManaCost:       opt.Val(cost.Mana{cost.O(5), cost.B, cost.B, cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.SourceCardPermanentReference()),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddAbilities: []game.Ability{
												new(game.TriggeredAbility{
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
																Primitive: game.PlayerLosesGame{
																	Player: game.EventPlayerReference(),
																},
															},
														},
													}.Ability(),
												}),
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.EquipActivatedAbility(cost.Mana{cost.B, cost.B}),
			},
			OracleText: `
			Equipped creature gets +2/+0 and has deathtouch.
			{5}{B}{B}{B}: Until end of turn, this Equipment gains "Whenever equipped creature deals combat damage to a player, that player loses the game."
			Equip {B}{B}
		`,
		},
	}
}
