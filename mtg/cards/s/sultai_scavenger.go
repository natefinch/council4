package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SultaiScavenger is the card definition for Sultai Scavenger.
//
// Type: Creature — Bird Warrior
// Cost: {5}{B}
//
// Oracle text:
//
//	Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
//	Flying
var SultaiScavenger = newSultaiScavenger

func newSultaiScavenger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Sultai Scavenger",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DelveStaticBody,
				game.FlyingStaticBody,
			},
			OracleText: `
			Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
			Flying
		`,
		},
	}
}
