package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MacetailHystrodon is the card definition for Macetail Hystrodon.
//
// Type: Creature — Beast
// Cost: {6}{R}
//
// Oracle text:
//
//	First strike, haste
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var MacetailHystrodon = newMacetailHystrodon

func newMacetailHystrodon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Macetail Hystrodon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			OracleText: `
			First strike, haste
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
