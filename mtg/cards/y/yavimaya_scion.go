package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YavimayaScion is the card definition for Yavimaya Scion.
//
// Type: Creature — Treefolk
// Cost: {4}{G}
//
// Oracle text:
//
//	Protection from artifacts
var YavimayaScion = newYavimayaScion

func newYavimayaScion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Yavimaya Scion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Treefolk},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ProtectionFromTypesStaticAbility(types.Artifact),
			},
			OracleText: `
			Protection from artifacts
		`,
		},
	}
}
