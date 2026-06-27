package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WildHypothesis is the card definition for Wild Hypothesis.
//
// Type: Sorcery
// Cost: {X}{G}
//
// Oracle text:
//
//	Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it.
//	Surveil 2. (Look at the top two cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)
var WildHypothesis = newWildHypothesis()

func newWildHypothesis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wild Hypothesis",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount:        game.Fixed(1),
							Source:        game.TokenDef(wildHypothesisToken),
							PublishLinked: game.LinkedKey("created-token"),
						},
					},
					{
						Primitive: game.AddCounter{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Object:      game.LinkedObjectReference("created-token"),
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					{
						Primitive: game.Surveil{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 0/0 green and blue Fractal creature token. Put X +1/+1 counters on it.
			Surveil 2. (Look at the top two cards of your library, then put any number of them into your graveyard and the rest on top of your library in any order.)
		`,
		},
	}
}

var wildHypothesisToken = newWildHypothesisToken()

func newWildHypothesisToken() *game.CardDef {
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
