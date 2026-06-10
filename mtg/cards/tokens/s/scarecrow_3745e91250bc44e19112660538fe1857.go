package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Scarecrow
//
// Type: Token Artifact Creature — Scarecrow
//
// Oracle text:

// ScarecrowToken3745e91250bc44e19112660538fe1857 is the card definition for Scarecrow.
var ScarecrowToken3745e91250bc44e19112660538fe1857 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Scarecrow",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Scarecrow},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
