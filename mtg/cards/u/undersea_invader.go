package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnderseaInvader is the card definition for Undersea Invader.
//
// Type: Creature — Giant Rogue
// Cost: {4}{U}{U}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	This creature enters tapped.
var UnderseaInvader = newUnderseaInvader

func newUnderseaInvader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Undersea Invader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Rogue},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This creature enters tapped."),
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			This creature enters tapped.
		`,
		},
	}
}
