package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelGargoyle is the card definition for Darksteel Gargoyle.
//
// Type: Artifact Creature — Gargoyle
// Cost: {7}
//
// Oracle text:
//
//	Flying
//	Indestructible (Damage and effects that say "destroy" don't destroy this creature.)
var DarksteelGargoyle = newDarksteelGargoyle

func newDarksteelGargoyle() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Gargoyle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Gargoyle},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.IndestructibleStaticBody,
			},
			OracleText: `
			Flying
			Indestructible (Damage and effects that say "destroy" don't destroy this creature.)
		`,
		},
	}
}
