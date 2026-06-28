package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FlamecacheGecko is the card definition for Flamecache Gecko.
//
// Type: Creature — Lizard Warlock
// Cost: {1}{R}
//
// Oracle text:
//
//	When this creature enters, if an opponent lost life this turn, add {B}{R}.
//	{1}{R}, Discard a card: Draw a card.
var FlamecacheGecko = newFlamecacheGecko()

func newFlamecacheGecko() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Flamecache Gecko",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard, types.Warlock},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}{R}, Discard a card: Draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(1), cost.R}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalDiscard,
							Text:   "Discard a card",
							Amount: 1,
							Source: zone.Hand,
						},
					},
					ZoneOfFunction: zone.Battlefield,
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if an opponent lost life this turn",
						InterveningCondition: opt.Val(game.Condition{
							EventHistory: opt.Val(game.EventHistoryCondition{Pattern: game.TriggerPattern{
								Event:  game.EventLifeLost,
								Player: game.TriggerPlayerOpponent,
							}, Window: game.EventHistoryCurrentTurn}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.B,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.R,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, if an opponent lost life this turn, add {B}{R}.
			{1}{R}, Discard a card: Draw a card.
		`,
		},
	}
}
