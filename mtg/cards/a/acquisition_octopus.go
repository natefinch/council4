package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AcquisitionOctopus is the card definition for Acquisition Octopus.
//
// Type: Artifact Creature — Equipment Octopus
// Cost: {2}{U}
//
// Oracle text:
//
//	Whenever this creature or equipped creature deals combat damage to a player, draw a card.
//	Reconfigure {2} ({2}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
var AcquisitionOctopus = newAcquisitionOctopus

func newAcquisitionOctopus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Acquisition Octopus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Equipment, types.Octopus},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ReconfigureActivatedAbility(cost.Mana{cost.O(2)}),
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
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature or equipped creature deals combat damage to a player, draw a card.
			Reconfigure {2} ({2}: Attach to target creature you control; or unattach from a creature. Reconfigure only as a sorcery. While attached, this isn't a creature.)
		`,
		},
	}
}
