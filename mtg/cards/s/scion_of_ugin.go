package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScionOfUgin is the card definition for Scion of Ugin.
//
// Type: Creature — Dragon Spirit
// Cost: {6}
//
// Oracle text:
//
//	Flying
var ScionOfUgin = newScionOfUgin

func newScionOfUgin() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scion of Ugin",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon, types.Spirit},
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
