package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AdornedPouncer is the card definition for Adorned Pouncer.
//
// Type: Creature — Cat
// Cost: {1}{W}
//
// Oracle text:
//
//	Double strike
//	Eternalize {3}{W}{W} ({3}{W}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a 4/4 black Zombie Cat with no mana cost. Eternalize only as a sorcery.)
var AdornedPouncer = newAdornedPouncer

func newAdornedPouncer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Adorned Pouncer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.EternalizeActivatedBody(cost.Mana{cost.O(3), cost.W, cost.W}, types.Sub("Cat")),
			},
			OracleText: `
			Double strike
			Eternalize {3}{W}{W} ({3}{W}{W}, Exile this card from your graveyard: Create a token that's a copy of it, except it's a 4/4 black Zombie Cat with no mana cost. Eternalize only as a sorcery.)
		`,
		},
	}
}
