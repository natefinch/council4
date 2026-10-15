package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InkwellLeviathan is the card definition for Inkwell Leviathan.
//
// Type: Artifact Creature — Leviathan
// Cost: {7}{U}{U}
//
// Oracle text:
//
//	Trample
//	Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
//	Shroud (This creature can't be the target of spells or abilities.)
var InkwellLeviathan = newInkwellLeviathan

func newInkwellLeviathan() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Inkwell Leviathan",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Leviathan},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 11}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.IslandwalkStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			Trample
			Islandwalk (This creature can't be blocked as long as defending player controls an Island.)
			Shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
