package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Spellbinder is the card definition for Spellbinder.
//
// Type: Artifact — Equipment
// Cost: {3}
//
// Oracle text:
//
//	Imprint — When this Equipment enters, you may exile an instant card from your hand.
//	Whenever equipped creature deals combat damage to a player, you may copy the exiled card. If you do, you may cast the copy without paying its mana cost.
//	Equip {4}
var Spellbinder = newSpellbinder

func newSpellbinder() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Spellbinder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Equipment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EquipActivatedAbility(cost.Mana{cost.O(4)}),
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
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseFromZone{
									Player:     game.ControllerReference(),
									SourceZone: zone.Hand,
									Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Instant}},
									Quantity:   game.Fixed(1),
									Destination: game.ChooseDestination{
										Zone: zone.Exile,
									},
									Riders: game.ChooseRiders{
										PublishLinked:       game.LinkedKey("imprint"),
										PublishObjectScoped: true,
									},
									Prompt: "Choose a card to exile",
								},
							},
						},
					}.Ability(),
				},
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
								Primitive: game.CopyCard{
									Player: game.ControllerReference(),
									LinkID: "imprint",
								},
								Optional:      true,
								PublishResult: game.ResultKey("imprint-copy-made"),
							},
							{
								Primitive: game.PlayLinkedExiledCard{
									Player:                game.ControllerReference(),
									LinkID:                "imprint",
									Copy:                  true,
									WithoutPayingManaCost: true,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "imprint-copy-made",
									Succeeded: game.TriTrue,
								}),
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Imprint — When this Equipment enters, you may exile an instant card from your hand.
			Whenever equipped creature deals combat damage to a player, you may copy the exiled card. If you do, you may cast the copy without paying its mana cost.
			Equip {4}
		`,
		},
	}
}
