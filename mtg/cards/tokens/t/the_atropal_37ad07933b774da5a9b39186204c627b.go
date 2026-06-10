package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// The Atropal
//
// Type: Token Legendary Creature — God Horror
//
// Oracle text:
//   Deathtouch

// TheAtropalToken37ad07933b774da5a9b39186204c627b is the card definition for The Atropal.
var TheAtropalToken37ad07933b774da5a9b39186204c627b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:       "The Atropal",
		Colors:     []color.Color{color.Black},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.God, types.Horror},
		Power:      opt.Val(game.PT{Value: 4}),
		Toughness:  opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.DeathtouchStaticBody,
		},
		OracleText: `
			Deathtouch
		`,
	},
}
