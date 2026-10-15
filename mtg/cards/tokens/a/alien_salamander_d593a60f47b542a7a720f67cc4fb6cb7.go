package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AlienSalamander is the card definition for Alien Salamander.
//
// Type: Token Creature — Alien Salamander
//
// Oracle text:
//   Islandwalk

// AlienSalamanderTokend593a60f47b542a7a720f67cc4fb6cb7 is the card definition for Alien Salamander.
var AlienSalamanderTokend593a60f47b542a7a720f67cc4fb6cb7 = newAlienSalamanderTokend593a60f47b542a7a720f67cc4fb6cb7()

func newAlienSalamanderTokend593a60f47b542a7a720f67cc4fb6cb7() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:      "Alien Salamander",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Alien, types.Salamander},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.IslandwalkStaticBody,
			},
			OracleText: `
			Islandwalk
		`,
		},
	}
}
