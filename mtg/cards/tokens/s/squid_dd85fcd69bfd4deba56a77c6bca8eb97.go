package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Squid is the card definition for Squid.
//
// Type: Token Creature — Squid
//
// Oracle text:
//   Islandwalk (This creature can't be blocked as long as defending player controls an Island.)

// SquidTokendd85fcd69bfd4deba56a77c6bca8eb97 is the card definition for Squid.
var SquidTokendd85fcd69bfd4deba56a77c6bca8eb97 = newSquidTokendd85fcd69bfd4deba56a77c6bca8eb97()

func newSquidTokendd85fcd69bfd4deba56a77c6bca8eb97() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Squid",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Squid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.IslandwalkStaticBody,
			},
			OracleText: `
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
		`,
		},
	}
}
