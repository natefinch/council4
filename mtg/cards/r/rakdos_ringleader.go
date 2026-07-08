package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RakdosRingleader is the card definition for Rakdos Ringleader.
//
// Type: Creature — Skeleton Warrior
// Cost: {4}{B}{R}
//
// Oracle text:
//
//	First strike
//	Whenever this creature deals combat damage to a player, that player discards a card at random.
//	{B}: Regenerate this creature.
var RakdosRingleader = newRakdosRingleader

func newRakdosRingleader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Rakdos Ringleader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.R,
			}),
			Colors:    []color.Color{color.Black, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Skeleton, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{B}: Regenerate this creature.",
					ManaCost:       opt.Val(cost.Mana{cost.B}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Regenerate{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount:   game.Fixed(1),
									Player:   game.EventPlayerReference(),
									AtRandom: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			First strike
			Whenever this creature deals combat damage to a player, that player discards a card at random.
			{B}: Regenerate this creature.
		`,
		},
	}
}
