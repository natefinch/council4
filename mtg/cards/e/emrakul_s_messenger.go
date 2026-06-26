package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmrakulSMessenger is the card definition for Emrakul's Messenger.
//
// Type: Creature — Eldrazi Faerie Rogue
// Cost: {1}{U}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Flying
//	Whenever you draw your second card each turn, create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
var EmrakulSMessenger = newEmrakulSMessenger()

func newEmrakulSMessenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Emrakul's Messenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Faerie, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventCardDrawn,
							Player:                     game.TriggerPlayerYou,
							PlayerEventOrdinalThisTurn: 2,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(emrakulSMessengerToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			Flying
			Whenever you draw your second card each turn, create a 0/1 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add {C}."
		`,
		},
	}
}

var emrakulSMessengerToken = newEmrakulSMessengerToken()

func newEmrakulSMessengerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Eldrazi Spawn",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Spawn},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
							Amount: 1,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
