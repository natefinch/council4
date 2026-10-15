package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PlatedCrusher is the card definition for Plated Crusher.
//
// Type: Creature — Beast
// Cost: {4}{G}{G}{G}
//
// Oracle text:
//
//	Trample
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
var PlatedCrusher = newPlatedCrusher

func newPlatedCrusher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Plated Crusher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HexproofStaticBody,
			},
			OracleText: `
			Trample
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
