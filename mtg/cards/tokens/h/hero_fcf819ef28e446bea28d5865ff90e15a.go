package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hero
//
// Type: Token Creature — Hero
//
// Oracle text:

// HeroTokenfcf819ef28e446bea28d5865ff90e15a is the card definition for Hero.
var HeroTokenfcf819ef28e446bea28d5865ff90e15a = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Hero",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Hero},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}
