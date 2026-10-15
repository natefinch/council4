package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KodamaOfTheNorthTree is the card definition for Kodama of the North Tree.
//
// Type: Legendary Creature — Spirit
// Cost: {2}{G}{G}{G}
//
// Oracle text:
//
//	Trample
//	Shroud (This creature can't be the target of spells or abilities.)
var KodamaOfTheNorthTree = newKodamaOfTheNorthTree

func newKodamaOfTheNorthTree() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Kodama of the North Tree",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit},
			Power:      opt.Val(game.PT{Value: 6}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			Trample
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
