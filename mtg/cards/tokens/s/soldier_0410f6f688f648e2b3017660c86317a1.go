package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Soldier
//
// Type: Token Artifact Creature — Soldier
//
// Oracle text:

// SoldierToken0410f6f688f648e2b3017660c86317a1 is the card definition for Soldier.
var SoldierToken0410f6f688f648e2b3017660c86317a1 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Soldier",
		Types:     []types.Card{types.Artifact, types.Creature},
		Subtypes:  []types.Sub{types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
