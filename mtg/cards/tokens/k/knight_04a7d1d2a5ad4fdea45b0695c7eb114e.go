package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Knight is the card definition for Knight.
//
// Type: Token Creature — Knight
//
// Oracle text:
//   Protection from white, haste
//   Flanking (Whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.)

// KnightToken04a7d1d2a5ad4fdea45b0695c7eb114e is the card definition for Knight.
var KnightToken04a7d1d2a5ad4fdea45b0695c7eb114e = newKnightToken04a7d1d2a5ad4fdea45b0695c7eb114e()

func newKnightToken04a7d1d2a5ad4fdea45b0695c7eb114e() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Knight",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Knight},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromColorsStaticAbility(color.White),
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.FlankingTriggeredBody,
			},
			OracleText: `
			Protection from white, haste
			Flanking (Whenever a creature without flanking blocks this creature, the blocking creature gets -1/-1 until end of turn.)
		`,
		},
	}
}
