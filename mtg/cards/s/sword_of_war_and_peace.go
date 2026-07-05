package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SwordOfWarAndPeace is the card definition for Sword of War and Peace.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Equipped creature gets +2/+2 and has protection from red and from white.
//	Whenever equipped creature deals combat damage to a player, this Equipment deals damage to that player equal to the number of cards in their hand and you gain 1 life for each card in your hand.
//	Equip {2}
var SwordOfWarAndPeace = newSwordOfWarAndPeace()

func newSwordOfWarAndPeace() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sword of War and Peace",
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
								new(game.ProtectionFromColorsStaticAbility(color.Red, color.White)),
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
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.EventPlayerReference(); return &ref }(),
										CardZone:   zone.Hand,
										Selection:  &game.Selection{},
									}),
									Recipient:    game.PlayerDamageRecipient(game.EventPlayerReference()),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountCountCardsInZone,
										Multiplier: 1,
										Player:     func() *game.PlayerReference { ref := game.ControllerReference(); return &ref }(),
										CardZone:   zone.Hand,
										Selection:  &game.Selection{},
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Equipped creature gets +2/+2 and has protection from red and from white.
			Whenever equipped creature deals combat damage to a player, this Equipment deals damage to that player equal to the number of cards in their hand and you gain 1 life for each card in your hand.
			Equip {2}
		`,
		},
	}
}
