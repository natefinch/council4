package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// WestgateRegent is the card definition for Westgate Regent.
//
// Type: Creature — Vampire
// Cost: {3}{B}{B}
//
// Oracle text:
//
//	Flying
//	Ward—Discard a card. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player discards a card.)
//	Whenever this creature deals combat damage to a player, put that many +1/+1 counters on it.
var WestgateRegent = newWestgateRegent

func newWestgateRegent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Westgate Regent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalDiscard,
						Text:   "Discard a card",
						Amount: 1,
						Source: zone.Hand,
					},
				}),
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
								Primitive: game.AddCounter{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountEventDamage,
										Multiplier: 1,
									}),
									Object:      game.EventPermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Ward—Discard a card. (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player discards a card.)
			Whenever this creature deals combat damage to a player, put that many +1/+1 counters on it.
		`,
		},
	}
}
