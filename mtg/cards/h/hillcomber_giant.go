package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HillcomberGiant is the card definition for Hillcomber Giant.
//
// Type: Creature — Giant Scout
// Cost: {2}{W}{W}
//
// Oracle text:
//
//	Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
var HillcomberGiant = newHillcomberGiant

func newHillcomberGiant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Hillcomber Giant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Scout},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MountainwalkStaticBody,
			},
			OracleText: `
			Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
		`,
		},
	}
}
