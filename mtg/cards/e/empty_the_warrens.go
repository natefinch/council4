package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EmptyTheWarrens is the card definition for Empty the Warrens.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	Create two 1/1 red Goblin creature tokens.
//	Storm (When you cast this spell, copy it for each spell cast before it this turn.)
var EmptyTheWarrens = newEmptyTheWarrens

func newEmptyTheWarrens() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Empty the Warrens",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StormStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(emptyTheWarrensToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create two 1/1 red Goblin creature tokens.
			Storm (When you cast this spell, copy it for each spell cast before it this turn.)
		`,
		},
	}
}

var emptyTheWarrensToken = newEmptyTheWarrensToken()

func newEmptyTheWarrensToken() *game.CardDef {
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
