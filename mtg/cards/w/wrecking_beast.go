package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WreckingBeast is the card definition for Wrecking Beast.
//
// Type: Creature — Beast
// Cost: {5}{G}{G}
//
// Oracle text:
//
//	Riot (This creature enters with your choice of a +1/+1 counter or haste.)
//	Trample
var WreckingBeast = newWreckingBeast

func newWreckingBeast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Wrecking Beast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.RiotStaticBody,
				game.TrampleStaticBody,
			},
			OracleText: `
			Riot (This creature enters with your choice of a +1/+1 counter or haste.)
			Trample
		`,
		},
	}
}
