package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Insect
//
// Type: Token Creature — Phyrexian Insect
//
// Oracle text:
//   Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)

// PhyrexianInsectTokend7e7f7f2d2e148ce97d055c43c360d88 is the card definition for Phyrexian Insect.
var PhyrexianInsectTokend7e7f7f2d2e148ce97d055c43c360d88 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Green),
	CardFace: game.CardFace{
		Name:      "Phyrexian Insect",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Insect},
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
