package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Goblin // Soldier
//
// Type: Token Creature — Goblin // Token Creature — Soldier
// Face: Soldier — Token Creature — Soldier
//
// Oracle text:
//   Soldier
//   Lifelink

// GoblinToken0823fbf27b394869b3f0091b7d7d0a5d is the card definition for Goblin.
var GoblinToken0823fbf27b394869b3f0091b7d7d0a5d = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Goblin",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
	Layout: game.LayoutDoubleFacedToken,
	Back: opt.Val(game.CardFace{
		Name:      "Soldier",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
		},
		OracleText: `
			Lifelink
		`,
	}),
}
