package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Eldrazi Horror
//
// Type: Token Creature — Eldrazi Horror
//
// Oracle text:

// EldraziHorrorToken294116b158c640bb89970deda327c522 is the card definition for Eldrazi Horror.
var EldraziHorrorToken294116b158c640bb89970deda327c522 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Eldrazi Horror",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Eldrazi, types.Horror},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 2}),
	},
}
