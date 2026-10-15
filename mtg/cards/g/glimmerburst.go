package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Glimmerburst is the card definition for Glimmerburst.
//
// Type: Instant
// Cost: {3}{U}
//
// Oracle text:
//
//	Draw two cards. Create a 1/1 white Glimmer enchantment creature token.
var Glimmerburst = newGlimmerburst

func newGlimmerburst() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Glimmerburst",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(glimmerburstToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards. Create a 1/1 white Glimmer enchantment creature token.
		`,
		},
	}
}

var glimmerburstToken = newGlimmerburstToken()

func newGlimmerburstToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Glimmer",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Glimmer},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
