package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RallyForTheThrone is the card definition for Rally for the Throne.
//
// Type: Instant
// Cost: {2}{W}
//
// Oracle text:
//
//	Create two 1/1 white Human creature tokens.
//	Adamant — If at least three white mana was spent to cast this spell, you gain 1 life for each creature you control.
var RallyForTheThrone = newRallyForTheThrone

func newRallyForTheThrone() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rally for the Throne",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(rallyForTheThroneToken),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							}),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.White, Count: 3},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 white Human creature tokens.
			Adamant — If at least three white mana was spent to cast this spell, you gain 1 life for each creature you control.
		`,
		},
	}
}

var rallyForTheThroneToken = newRallyForTheThroneToken()

func newRallyForTheThroneToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Human",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
