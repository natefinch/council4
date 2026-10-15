package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PrimocEscapee is the card definition for Primoc Escapee.
//
// Type: Creature — Bird Beast
// Cost: {6}{U}
//
// Oracle text:
//
//	Flying
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var PrimocEscapee = newPrimocEscapee

func newPrimocEscapee() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Primoc Escapee",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird, types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			OracleText: `
			Flying
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
