package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YargleGoliathOfOtaria is the card definition for Yargle, Goliath of Otaria.
//
// Type: Legendary Creature — Frog Spirit
// Cost: {4}{U}
//
// Oracle text:
var YargleGoliathOfOtaria = newYargleGoliathOfOtaria

func newYargleGoliathOfOtaria() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Yargle, Goliath of Otaria",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Frog, types.Spirit},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 9}),
		},
	}
}
