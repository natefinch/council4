package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NipGwyllion is the card definition for Nip Gwyllion.
//
// Type: Creature — Hag
// Cost: {W/B}
//
// Oracle text:
//
//	Lifelink (Damage dealt by this creature also causes you to gain that much life.)
var NipGwyllion = newNipGwyllion

func newNipGwyllion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Nip Gwyllion",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.W, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hag},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.LifelinkStaticBody,
			},
			OracleText: `
			Lifelink (Damage dealt by this creature also causes you to gain that much life.)
		`,
		},
	}
}
