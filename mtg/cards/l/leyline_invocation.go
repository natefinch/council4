package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeylineInvocation is the card definition for Leyline Invocation.
//
// Type: Sorcery
// Cost: {5}{G}
//
// Oracle text:
//
//	Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it, where X is the number of lands you control.
var LeylineInvocation = newLeylineInvocation

func newLeylineInvocation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Leyline Invocation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:        game.Fixed(1),
							Source:        game.TokenDef(leylineInvocationToken),
							PublishLinked: game.LinkedKey("created-token"),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountCountSelector,
								Multiplier: 1,
								Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
							}),
							Object:      game.LinkedObjectReference("created-token"),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it, where X is the number of lands you control.
		`,
		},
	}
}

var leylineInvocationToken = newLeylineInvocationToken()

func newLeylineInvocationToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Fractal",
			Colors:    []color.Color{color.Green, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fractal},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
		},
	}
}
