package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CarnageTyrant is the card definition for Carnage Tyrant.
//
// Type: Creature — Dinosaur
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Trample, hexproof
var CarnageTyrant = newCarnageTyrant

func newCarnageTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Carnage Tyrant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dinosaur},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.TrampleStaticBody,
				game.HexproofStaticBody,
			},
			OracleText: `
			This spell can't be countered.
			Trample, hexproof
		`,
		},
	}
}
