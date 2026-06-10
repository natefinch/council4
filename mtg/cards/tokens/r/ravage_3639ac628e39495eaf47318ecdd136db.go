package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ravage
//
// Type: Token Legendary Artifact Creature — Robot
//
// Oracle text:
//   Menace, deathtouch

// RavageToken3639ac628e39495eaf47318ecdd136db is the card definition for Ravage.
var RavageToken3639ac628e39495eaf47318ecdd136db = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:       "Ravage",
		Colors:     []color.Color{color.Black},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Artifact, types.Creature},
		Subtypes:   []types.Sub{types.Robot},
		Power:      opt.Val(game.PT{Value: 3}),
		Toughness:  opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
			game.DeathtouchStaticBody,
		},
		OracleText: `
			Menace, deathtouch
		`,
	},
}
