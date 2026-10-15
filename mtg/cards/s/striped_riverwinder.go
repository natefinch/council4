package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StripedRiverwinder is the card definition for Striped Riverwinder.
//
// Type: Creature — Serpent
// Cost: {6}{U}
//
// Oracle text:
//
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
//	Cycling {U} ({U}, Discard this card: Draw a card.)
var StripedRiverwinder = newStripedRiverwinder

func newStripedRiverwinder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Striped Riverwinder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Serpent},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.HexproofStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.U}),
			},
			OracleText: `
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
			Cycling {U} ({U}, Discard this card: Draw a card.)
		`,
		},
	}
}
