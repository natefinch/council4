package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YotianSoldier is the card definition for Yotian Soldier.
//
// Type: Artifact Creature — Soldier
// Cost: {3}
//
// Oracle text:
//
//	Vigilance
var YotianSoldier = newYotianSoldier

func newYotianSoldier() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Yotian Soldier",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			OracleText: `
			Vigilance
		`,
		},
	}
}
