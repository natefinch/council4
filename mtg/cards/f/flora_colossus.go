package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FloraColossus is the card definition for Flora Colossus.
//
// Type: Creature — Treefolk
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
//	Flora Colossus's power and toughness are each equal to the number of lands you control.
var FloraColossus = newFloraColossus

func newFloraColossus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Flora Colossus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:           []color.Color{color.Green},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Treefolk},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerLandCount}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			OracleText: `
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
			Flora Colossus's power and toughness are each equal to the number of lands you control.
		`,
		},
	}
}
