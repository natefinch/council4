package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HypnoticSpecter is the card definition for Hypnotic Specter.
//
// Type: Creature — Specter
// Cost: {1}{B}{B}
//
// Oracle text:
//
//	Flying
//	Whenever this creature deals damage to an opponent, that player discards a card at random.
var HypnoticSpecter = newHypnoticSpecter

func newHypnoticSpecter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Hypnotic Specter",
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
							Event:           game.EventDamageDealt,
							Source:          game.TriggerSourceSelf,
							Subject:         game.TriggerSubjectDamageSource,
							Player:          game.TriggerPlayerOpponent,
							DamageRecipient: game.DamageRecipientPlayer,
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
			Flying
			Whenever this creature deals damage to an opponent, that player discards a card at random.
		`,
		},
	}
}
