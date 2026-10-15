package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SwiftSpinner is the card definition for Swift Spinner.
//
// Type: Creature — Spider
// Cost: {3}{G}
//
// Oracle text:
//
//	Flash (You may cast this spell any time you could cast an instant.)
//	Reach (This creature can block creatures with flying.)
var SwiftSpinner = newSwiftSpinner

func newSwiftSpinner() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Swift Spinner",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.ReachStaticBody,
			},
			OracleText: `
			Flash (You may cast this spell any time you could cast an instant.)
			Reach (This creature can block creatures with flying.)
		`,
		},
	}
}
