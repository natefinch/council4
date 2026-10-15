package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cadet is the card definition for Cadet.
//
// Type: Token Creature — Wizard Soldier
//
// Oracle text:

// CadetToken2d32116165cd43ed8285551e3d4989e4 is the card definition for Cadet.
var CadetToken2d32116165cd43ed8285551e3d4989e4 = newCadetToken2d32116165cd43ed8285551e3d4989e4()

func newCadetToken2d32116165cd43ed8285551e3d4989e4() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cadet",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wizard, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
