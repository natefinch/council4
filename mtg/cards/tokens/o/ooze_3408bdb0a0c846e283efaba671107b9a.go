package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ooze is the card definition for Ooze.
//
// Type: Token Creature — Ooze
//
// Oracle text:
//   This creature's power is equal to the number of card types among cards in your graveyard and its toughness is equal to that number plus 1.

// OozeToken3408bdb0a0c846e283efaba671107b9a is the card definition for Ooze.
var OozeToken3408bdb0a0c846e283efaba671107b9a = newOozeToken3408bdb0a0c846e283efaba671107b9a()

func newOozeToken3408bdb0a0c846e283efaba671107b9a() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:             "Ooze",
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Ooze},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCardTypesInGraveyard}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerCardTypesInGraveyard, Offset: 1}),
			OracleText: `
			This creature's power is equal to the number of card types among cards in your graveyard and its toughness is equal to that number plus 1.
		`,
		},
	}
}
