package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HitMonkey is the card definition for Hit-Monkey.
//
// Type: Legendary Creature — Monkey Assassin
// Cost: {3}{G}
//
// Oracle text:
//
//	This spell can't be countered.
//	Reach, vigilance, deathtouch, hexproof, haste
var HitMonkey = newHitMonkey

func newHitMonkey() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Hit-Monkey",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Monkey, types.Assassin},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CantBeCounteredStaticBody,
				game.ReachStaticBody,
				game.VigilanceStaticBody,
				game.DeathtouchStaticBody,
				game.HexproofStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			This spell can't be countered.
			Reach, vigilance, deathtouch, hexproof, haste
		`,
		},
	}
}
