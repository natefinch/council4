package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HuntingPack is the card definition for Hunting Pack.
//
// Type: Instant
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Create a 4/4 green Beast creature token.
//	Storm (When you cast this spell, copy it for each spell cast before it this turn.)
var HuntingPack = newHuntingPack

func newHuntingPack() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hunting Pack",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StormStaticBody,
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(huntingPackToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Create a 4/4 green Beast creature token.
			Storm (When you cast this spell, copy it for each spell cast before it this turn.)
		`,
		},
	}
}

var huntingPackToken = newHuntingPackToken()

func newHuntingPackToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Beast",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
