package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GrimFeast is the card definition for Grim Feast.
//
// Type: Enchantment
// Cost: {1}{B}{G}
//
// Oracle text:
//
//	At the beginning of your upkeep, this enchantment deals 1 damage to you.
//	Whenever a creature is put into an opponent's graveyard from the battlefield, you gain life equal to its toughness.
var GrimFeast = newGrimFeast()

func newGrimFeast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Grim Feast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.PlayerDamageRecipient(game.ControllerReference()),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Player:           game.TriggerPlayerOpponent,
							MatchFromZone:    true,
							FromZone:         zone.Battlefield,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
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
			At the beginning of your upkeep, this enchantment deals 1 damage to you.
			Whenever a creature is put into an opponent's graveyard from the battlefield, you gain life equal to its toughness.
		`,
		},
	}
}
