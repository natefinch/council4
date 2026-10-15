package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WallOfShields is the card definition for Wall of Shields.
//
// Type: Artifact Creature — Wall
// Cost: {3}
//
// Oracle text:
//
//	Defender (This creature can't attack.)
//	Banding (If any creatures with banding you control are blocking a creature, you divide that creature's combat damage, not its controller, among any of the creatures it's being blocked by.)
var WallOfShields = newWallOfShields

func newWallOfShields() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Wall of Shields",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.BandingStaticBody,
			},
			OracleText: `
			Defender (This creature can't attack.)
			Banding (If any creatures with banding you control are blocking a creature, you divide that creature's combat damage, not its controller, among any of the creatures it's being blocked by.)
		`,
		},
	}
}
