package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OonaSGatewarden is the card definition for Oona's Gatewarden.
//
// Type: Creature — Faerie Soldier
// Cost: {U/B}
//
// Oracle text:
//
//	Defender, flying
//	Wither (This deals damage to creatures in the form of -1/-1 counters.)
var OonaSGatewarden = newOonaSGatewarden

func newOonaSGatewarden() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Oona's Gatewarden",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.U, mana.B),
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.FlyingStaticBody,
				game.WitherStaticBody,
			},
			OracleText: `
			Defender, flying
			Wither (This deals damage to creatures in the form of -1/-1 counters.)
		`,
		},
	}
}
