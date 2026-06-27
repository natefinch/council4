package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IkraShidiqiTheUsurper is the card definition for Ikra Shidiqi, the Usurper.
//
// Type: Legendary Creature — Snake Wizard
// Cost: {3}{B}{G}
//
// Oracle text:
//
//	Menace
//	Whenever a creature you control deals combat damage to a player, you gain life equal to that creature's toughness.
//	Partner (You can have two commanders if both have partner.)
var IkraShidiqiTheUsurper = newIkraShidiqiTheUsurper()

func newIkraShidiqiTheUsurper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Ikra Shidiqi, the Usurper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Snake, types.Wizard},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.PartnerStaticBody,
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
								Primitive: game.GainLife{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectToughness,
										Multiplier: 1,
										Object:     game.EventPermanentReference(),
									}),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Menace
			Whenever a creature you control deals combat damage to a player, you gain life equal to that creature's toughness.
			Partner (You can have two commanders if both have partner.)
		`,
		},
	}
}
