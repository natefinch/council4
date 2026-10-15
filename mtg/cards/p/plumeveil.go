package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Plumeveil is the card definition for Plumeveil.
//
// Type: Creature — Elemental
// Cost: {W/U}{W/U}{W/U}
//
// Oracle text:
//
//	Flash
//	Defender, flying
var Plumeveil = newPlumeveil

func newPlumeveil() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Plumeveil",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.W, mana.U),
				cost.HybridMana(mana.W, mana.U),
				cost.HybridMana(mana.W, mana.U),
			}),
			Colors:    []color.Color{color.Blue, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.DefenderStaticBody,
				game.FlyingStaticBody,
			},
			OracleText: `
			Flash
			Defender, flying
		`,
		},
	}
}
