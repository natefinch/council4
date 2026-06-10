package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Horror
//
// Type: Token Artifact Creature — Horror
//
// Oracle text:

// HorrorToken7f03946be6024624bf1eed240b883fca is the card definition for Horror.
var HorrorToken7f03946be6024624bf1eed240b883fca = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Horror",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Horror},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
	},
}
