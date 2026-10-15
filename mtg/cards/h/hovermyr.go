package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hovermyr is the card definition for Hovermyr.
//
// Type: Artifact Creature — Myr
// Cost: {2}
//
// Oracle text:
//
//	Flying, vigilance
var Hovermyr = newHovermyr

func newHovermyr() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Hovermyr",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.VigilanceStaticBody,
			},
			OracleText: `
			Flying, vigilance
		`,
		},
	}
}
