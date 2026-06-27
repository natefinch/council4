package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlimeMolding is the card definition for Slime Molding.
//
// Type: Sorcery
// Cost: {X}{G}
//
// Oracle text:
//
//	Create an X/X green Ooze creature token.
var SlimeMolding = newSlimeMolding()

func newSlimeMolding() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Slime Molding",
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
							Amount: game.Fixed(1),
							Source: game.TokenDef(slimeMoldingToken),
							Power: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							})),
							Toughness: opt.Val(game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							})),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create an X/X green Ooze creature token.
		`,
		},
	}
}

var slimeMoldingToken = newSlimeMoldingToken()

func newSlimeMoldingToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Ooze",
			Colors:   []color.Color{color.Green},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Ooze},
		},
	}
}
