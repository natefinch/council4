package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RampagingRendhorn is the card definition for Rampaging Rendhorn.
//
// Type: Creature — Beast
// Cost: {4}{G}
//
// Oracle text:
//
//	Riot (This creature enters with your choice of a +1/+1 counter or haste.)
var RampagingRendhorn = newRampagingRendhorn

func newRampagingRendhorn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Rampaging Rendhorn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.RiotStaticBody,
			},
			OracleText: `
			Riot (This creature enters with your choice of a +1/+1 counter or haste.)
		`,
		},
	}
}
