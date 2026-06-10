package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Insect
//
// Type: Token Creature — Insect
//
// Oracle text:
//   Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)

// InsectToken5769718d67814919bc9b245e9e0cb507 is the card definition for Insect.
var InsectToken5769718d67814919bc9b245e9e0cb507 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Insect",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Insect},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.InfectStaticBody,
		},
		OracleText: `
			Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)
		`,
	},
}
