package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Kraken
//
// Type: Token Creature — Kraken
//
// Oracle text:
//   Hexproof (This creature can't be the target of spells or abilities your opponents control.)

// KrakenTokenfcd2cb28287f41598d77853293372438 is the card definition for Kraken.
var KrakenTokenfcd2cb28287f41598d77853293372438 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue),
	CardFace: game.CardFace{
		Name:      "Kraken",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Kraken},
		Power:     opt.Val(game.PT{Value: 8}),
		Toughness: opt.Val(game.PT{Value: 8}),
		StaticAbilities: []game.StaticAbility{
			game.HexproofStaticBody,
		},
		OracleText: `
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
	},
}
