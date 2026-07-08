package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OgreHeadHelm is the card definition for Ogre-Head Helm.
//
// Type: Artifact Creature — Equipment Ogre
// Cost: {1}{R}
//
// Oracle text:
//
//	Equipped creature gets +2/+2.
//	Whenever this creature or equipped creature deals combat damage to a player, you may sacrifice it. If you do, discard your hand, then draw three cards.
//	Reconfigure {3} ({3}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
var OgreHeadHelm = newOgreHeadHelm

func newOgreHeadHelm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ogre-Head Helm",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Equipment, types.Ogre},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.AttachedObjectGroup(game.SourcePermanentReference()),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ReconfigureActivatedAbility(cost.Mana{cost.O(3)}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                       game.EventDamageDealt,
							Source:                      game.TriggerSourceAttachedPermanent,
							Subject:                     game.TriggerSubjectDamageSource,
							RequireCombatDamage:         true,
							DamageRecipient:             game.DamageRecipientPlayer,
							DamageSourceSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}},
							DamageSourceSelectionOrSelf: true,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.EventPermanentReference(),
								},
								Optional:      true,
								PublishResult: game.ResultKey("if-you-do"),
							},
							{
								Primitive: game.Discard{
									EntireHand: true,
									Player:     game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "if-you-do",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Equipped creature gets +2/+2.
			Whenever this creature or equipped creature deals combat damage to a player, you may sacrifice it. If you do, discard your hand, then draw three cards.
			Reconfigure {3} ({3}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
		`,
		},
	}
}
