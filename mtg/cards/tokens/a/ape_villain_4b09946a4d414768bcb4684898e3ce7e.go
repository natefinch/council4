package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ApeVillain is the card definition for Ape Villain.
//
// Type: Token Creature — Ape Villain
//
// Oracle text:
//   Haste

// ApeVillainToken4b09946a4d414768bcb4684898e3ce7e is the card definition for Ape Villain.
var ApeVillainToken4b09946a4d414768bcb4684898e3ce7e = newApeVillainToken4b09946a4d414768bcb4684898e3ce7e()

func newApeVillainToken4b09946a4d414768bcb4684898e3ce7e() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:      "Ape Villain",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ape, types.Villain},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			OracleText: `
			Haste
		`,
		},
	}
}
