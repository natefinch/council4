package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Unwavering Initiate
//
// Type: Token Creature — Zombie Human Warrior
//
// Oracle text:
//   Vigilance

// UnwaveringInitiateToken92837e49ebc24a0586fdcf014314e1e0 is the card definition for Unwavering Initiate.
var UnwaveringInitiateToken92837e49ebc24a0586fdcf014314e1e0 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Unwavering Initiate",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie, types.Human, types.Warrior},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
}
