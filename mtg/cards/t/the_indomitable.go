package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheIndomitable is the card definition for The Indomitable.
//
// Type: Legendary Artifact — Vehicle
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	Trample
//	Whenever a creature you control deals combat damage to a player, draw a card.
//	Crew 3
//	You may cast this card from your graveyard as long as you control three or more tapped Pirates and/or Vehicles.
var TheIndomitable = newTheIndomitable

func newTheIndomitable() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "The Indomitable",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			Subtypes:   []types.Sub{types.Vehicle},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					ZoneOfFunction: zone.Graveyard,
					Condition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Pirate"), types.Sub("Vehicle")}, Tapped: game.TriTrue},
							MinCount:  3,
						}),
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCastFromZone,
							AffectedSource: true,
							AffectedPlayer: game.PlayerYou,
							CastFromZone:   zone.Graveyard,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CrewActivatedAbility(3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
			Trample
			Whenever a creature you control deals combat damage to a player, draw a card.
			Crew 3
			You may cast this card from your graveyard as long as you control three or more tapped Pirates and/or Vehicles.
		`,
		},
	}
}
