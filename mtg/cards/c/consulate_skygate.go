package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ConsulateSkygate is the card definition for Consulate Skygate.
//
// Type: Artifact Creature — Wall
// Cost: {2}
//
// Oracle text:
//
//	Defender
//	Reach (This creature can block creatures with flying.)
var ConsulateSkygate = newConsulateSkygate

func newConsulateSkygate() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Consulate Skygate",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Wall},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
				game.ReachStaticBody,
			},
			OracleText: `
			Defender
			Reach (This creature can block creatures with flying.)
		`,
		},
	}
}
