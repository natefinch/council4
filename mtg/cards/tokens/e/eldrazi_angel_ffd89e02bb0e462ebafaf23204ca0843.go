package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Eldrazi Angel
//
// Type: Token Creature — Eldrazi Angel
//
// Oracle text:
//   Flying, vigilance

// EldraziAngelTokenffd89e02bb0e462ebafaf23204ca0843 is the card definition for Eldrazi Angel.
var EldraziAngelTokenffd89e02bb0e462ebafaf23204ca0843 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Eldrazi Angel",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Eldrazi, types.Angel},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.VigilanceStaticBody,
		},
		OracleText: `
			Flying, vigilance
		`,
	},
}
