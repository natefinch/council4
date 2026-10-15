package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LoamdraggerGiant is the card definition for Loamdragger Giant.
//
// Type: Creature — Giant Warrior
// Cost: {4}{R/G}{R/G}{R/G}
//
// Oracle text:
var LoamdraggerGiant = newLoamdraggerGiant

func newLoamdraggerGiant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Loamdragger Giant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.HybridMana(mana.R, mana.G),
				cost.HybridMana(mana.R, mana.G),
				cost.HybridMana(mana.R, mana.G),
			}),
			Colors:    []color.Color{color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Warrior},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 6}),
		},
	}
}
