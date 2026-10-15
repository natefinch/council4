package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TransdimensionalBovine is the card definition for Transdimensional Bovine.
//
// Type: Creature — Ox Avatar
// Cost: {2}{G}
//
// Oracle text:
//
//	Flying
//	{T}: Add two mana of any one color.
var TransdimensionalBovine = newTransdimensionalBovine

func newTransdimensionalBovine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Transdimensional Bovine",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ox, types.Avatar},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceCountAbility("{T}: Add two mana of any one color.", 2, mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			Flying
			{T}: Add two mana of any one color.
		`,
		},
	}
}
