package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HeadlessSpecter is the card definition for Headless Specter.
//
// Type: Creature — Specter
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Flying
//	Hellbent — Whenever this creature deals combat damage to a player, if you have no cards in hand, that player discards a card at random.
var HeadlessSpecter = newHeadlessSpecter

func newHeadlessSpecter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Headless Specter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Specter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
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
						InterveningIf: "if you have no cards in hand",
						InterveningCondition: opt.Val(game.Condition{
							ControllerHandEmpty: true,
						}),
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
			Flying
			Hellbent — Whenever this creature deals combat damage to a player, if you have no cards in hand, that player discards a card at random.
		`,
		},
	}
}
