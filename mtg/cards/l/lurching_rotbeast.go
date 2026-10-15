package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LurchingRotbeast is the card definition for Lurching Rotbeast.
//
// Type: Creature — Zombie Beast
// Cost: {3}{B}
//
// Oracle text:
//
//	Cycling {B} ({B}, Discard this card: Draw a card.)
var LurchingRotbeast = newLurchingRotbeast

func newLurchingRotbeast() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Lurching Rotbeast",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.B}),
			},
			OracleText: `
			Cycling {B} ({B}, Discard this card: Draw a card.)
		`,
		},
	}
}
