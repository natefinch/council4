package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SeraphOfTheSuns is the card definition for Seraph of the Suns.
//
// Type: Creature — Angel
// Cost: {5}{W}{W}
//
// Oracle text:
//
//	Flying
//	Indestructible (Damage and effects that say "destroy" don't destroy this creature. If its toughness is 0 or less, it still dies.)
var SeraphOfTheSuns = newSeraphOfTheSuns

func newSeraphOfTheSuns() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Seraph of the Suns",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Angel},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.IndestructibleStaticBody,
			},
			OracleText: `
			Flying
			Indestructible (Damage and effects that say "destroy" don't destroy this creature. If its toughness is 0 or less, it still dies.)
		`,
		},
	}
}
