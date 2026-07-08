package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PonybackBrigade is the card definition for Ponyback Brigade.
//
// Type: Creature — Goblin Warrior
// Cost: {3}{R}{W}{B}
//
// Oracle text:
//
//	When this creature enters or is turned face up, create three 1/1 red Goblin creature tokens.
//	Morph {2}{R}{W}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var PonybackBrigade = newPonybackBrigade

func newPonybackBrigade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Ponyback Brigade",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{Cost: cost.Mana{cost.O(2), cost.R, cost.W, cost.B}},
					},
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
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(3),
									Source: game.TokenDef(ponybackBrigadeToken),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentTurnedFaceUp,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(3),
									Source: game.TokenDef(ponybackBrigadeToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters or is turned face up, create three 1/1 red Goblin creature tokens.
			Morph {2}{R}{W}{B} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}

var ponybackBrigadeToken = newPonybackBrigadeToken()

func newPonybackBrigadeToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Goblin",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
