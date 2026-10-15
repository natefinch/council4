package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CallOfTheConclave is the card definition for Call of the Conclave.
//
// Type: Sorcery
// Cost: {G}{W}
//
// Oracle text:
//
//	Create a 3/3 green Centaur creature token.
var CallOfTheConclave = newCallOfTheConclave

func newCallOfTheConclave() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Call of the Conclave",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.W,
			}),
			Colors: []color.Color{color.Green, color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(callOfTheConclaveToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 3/3 green Centaur creature token.
		`,
		},
	}
}

var callOfTheConclaveToken = newCallOfTheConclaveToken()

func newCallOfTheConclaveToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Centaur",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Centaur},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
