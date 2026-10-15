package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GiantSolifuge is the card definition for Giant Solifuge.
//
// Type: Creature — Insect
// Cost: {2}{R/G}{R/G}
//
// Oracle text:
//
//	({R/G} can be paid with either {R} or {G}.)
//	Trample; haste; shroud (This creature can't be the target of spells or abilities.)
var GiantSolifuge = newGiantSolifuge

func newGiantSolifuge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Giant Solifuge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.R, mana.G),
				cost.HybridMana(mana.R, mana.G),
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HasteStaticBody,
				game.ShroudStaticBody,
			},
			OracleText: `
			({R/G} can be paid with either {R} or {G}.)
			Trample; haste; shroud (This creature can't be the target of spells or abilities.)
		`,
		},
	}
}
