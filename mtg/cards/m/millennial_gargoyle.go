package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MillennialGargoyle is the card definition for Millennial Gargoyle.
//
// Type: Artifact Creature — Gargoyle
// Cost: {4}
//
// Oracle text:
//
//	Flying
var MillennialGargoyle = newMillennialGargoyle

func newMillennialGargoyle() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Millennial Gargoyle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
		`,
		},
	}
}
