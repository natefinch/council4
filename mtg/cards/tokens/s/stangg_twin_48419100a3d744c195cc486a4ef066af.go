package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Stangg Twin
//
// Type: Token Legendary Creature — Human Warrior
//
// Oracle text:

// StanggTwinToken48419100a3d744c195cc486a4ef066af is the card definition for Stangg Twin.
var StanggTwinToken48419100a3d744c195cc486a4ef066af = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red, color.Green),
	CardFace: game.CardFace{
		Name:       "Stangg Twin",
		Colors:     []color.Color{color.Green, color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Human, types.Warrior},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 4}),
	},
}
