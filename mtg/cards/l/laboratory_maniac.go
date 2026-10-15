package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LaboratoryManiac is the card definition for Laboratory Maniac.
//
// Type: Creature — Human Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	If you would draw a card while your library has no cards in it, you win the game instead.
var LaboratoryManiac = newLaboratoryManiac

func newLaboratoryManiac() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Laboratory Maniac",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.DrawFromEmptyLibraryWinReplacement("If you would draw a card while your library has no cards in it, you win the game instead."),
			},
			OracleText: `
			If you would draw a card while your library has no cards in it, you win the game instead.
		`,
		},
	}
}
