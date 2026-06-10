package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Nightmare Horror
//
// Type: Token Creature — Nightmare Horror
//
// Oracle text:

// NightmareHorrorTokeneabcfcf54f5142c18eb838810d6ebd47 is the card definition for Nightmare Horror.
var NightmareHorrorTokeneabcfcf54f5142c18eb838810d6ebd47 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Nightmare Horror",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Nightmare, types.Horror},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
