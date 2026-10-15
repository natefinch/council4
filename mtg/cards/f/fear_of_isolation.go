package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FearOfIsolation is the card definition for Fear of Isolation.
//
// Type: Enchantment Creature — Nightmare
// Cost: {1}{U}
//
// Oracle text:
//
//	As an additional cost to cast this spell, return a permanent you control to its owner's hand.
//	Flying
var FearOfIsolation = newFearOfIsolation

func newFearOfIsolation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Fear of Isolation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Enchantment, types.Creature},
			Subtypes:  []types.Sub{types.Nightmare},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			AdditionalCosts: []cost.Additional{
				{
					Kind:   cost.AdditionalReturnToHand,
					Text:   "return a permanent you control to its owner's hand",
					Amount: 1,
				},
			},
			OracleText: `
			As an additional cost to cast this spell, return a permanent you control to its owner's hand.
			Flying
		`,
		},
	}
}
