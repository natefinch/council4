package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ValgavothHarrowerOfSouls is the card definition for Valgavoth, Harrower of Souls.
//
// Type: Legendary Creature — Elder Demon
// Cost: {2}{B}{R}
//
// Oracle text:
//
//	Flying
//	Ward—Pay 2 life.
//	Whenever an opponent loses life for the first time during each of their turns, put a +1/+1 counter on Valgavoth and draw a card.
var ValgavothHarrowerOfSouls = newValgavothHarrowerOfSouls()

func newValgavothHarrowerOfSouls() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Valgavoth, Harrower of Souls",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elder, types.Demon},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.WardStaticAbilityWithCosts(cost.Mana{}, []cost.Additional{
					{
						Kind:   cost.AdditionalPayLife,
						Text:   "Pay 2 life",
						Amount: 2,
					},
				}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventLifeLost,
							Player:                     game.TriggerPlayerOpponent,
							CastDuringTurn:             game.TriggerTurnNotYours,
							PlayerEventOrdinalThisTurn: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
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
			Flying
			Ward—Pay 2 life.
			Whenever an opponent loses life for the first time during each of their turns, put a +1/+1 counter on Valgavoth and draw a card.
		`,
		},
	}
}
