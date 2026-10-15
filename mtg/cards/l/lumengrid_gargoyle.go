package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LumengridGargoyle is the card definition for Lumengrid Gargoyle.
//
// Type: Artifact Creature — Gargoyle
// Cost: {6}
//
// Oracle text:
//
//	Flying
var LumengridGargoyle = newLumengridGargoyle

func newLumengridGargoyle() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Lumengrid Gargoyle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
		`,
		},
	}
}
