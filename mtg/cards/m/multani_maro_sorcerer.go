package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MultaniMaroSorcerer is the card definition for Multani, Maro-Sorcerer.
//
// Type: Legendary Creature — Elemental Sorcerer
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	Shroud (This creature can't be the target of spells or abilities.)
//	Multani's power and toughness are each equal to the total number of cards in all players' hands.
var MultaniMaroSorcerer = newMultaniMaroSorcerer

func newMultaniMaroSorcerer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Multani, Maro-Sorcerer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Supertypes:       []types.Super{types.Legendary},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Elemental, types.Sorcerer},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueAllPlayersHandSize}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueAllPlayersHandSize}),
			StaticAbilities: []game.StaticAbility{
				game.ShroudStaticBody,
			},
			OracleText: `
			Shroud (This creature can't be the target of spells or abilities.)
			Multani's power and toughness are each equal to the total number of cards in all players' hands.
		`,
		},
	}
}
