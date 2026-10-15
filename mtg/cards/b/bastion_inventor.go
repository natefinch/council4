package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BastionInventor is the card definition for Bastion Inventor.
//
// Type: Creature — Vedalken Artificer
// Cost: {5}{U}
//
// Oracle text:
//
//	Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
//	Hexproof (This creature can't be the target of spells or abilities your opponents control.)
var BastionInventor = newBastionInventor

func newBastionInventor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Bastion Inventor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vedalken, types.Artificer},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ImproviseStaticBody,
				game.HexproofStaticBody,
			},
			OracleText: `
			Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
			Hexproof (This creature can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
