package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Tombstalker is the card definition for Tombstalker.
//
// Type: Creature — Demon
// Cost: {6}{B}{B}
//
// Oracle text:
//
//	Flying
//	Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
var Tombstalker = newTombstalker

func newTombstalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Tombstalker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Demon},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DelveStaticBody,
			},
			OracleText: `
			Flying
			Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
		`,
		},
	}
}
