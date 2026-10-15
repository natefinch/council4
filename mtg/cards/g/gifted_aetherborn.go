package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GiftedAetherborn is the card definition for Gifted Aetherborn.
//
// Type: Creature — Aetherborn Vampire
// Cost: {B}{B}
//
// Oracle text:
//
//	Deathtouch, lifelink
var GiftedAetherborn = newGiftedAetherborn

func newGiftedAetherborn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Gifted Aetherborn",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Aetherborn, types.Vampire},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
				game.LifelinkStaticBody,
			},
			OracleText: `
			Deathtouch, lifelink
		`,
		},
	}
}
