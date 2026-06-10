package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Karox Bladewing
//
// Type: Token Legendary Creature — Dragon
//
// Oracle text:
//   Flying

// KaroxBladewingToken2aa23bbfecd84ae8b4a3fd1e82e5e796 is the card definition for Karox Bladewing.
var KaroxBladewingToken2aa23bbfecd84ae8b4a3fd1e82e5e796 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:       "Karox Bladewing",
		Colors:     []color.Color{color.Red},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Dragon},
		Power:      opt.Val(game.PT{Value: 4}),
		Toughness:  opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
