package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Drone
//
// Type: Token Artifact Creature — Drone
//
// Oracle text:
//   Flying
//   This token can block only creatures with flying.

// DroneToken1c68ec5e2ffa48b7b3694b6148d9dce4 is the card definition for Drone.
var DroneToken1c68ec5e2ffa48b7b3694b6148d9dce4 = newDroneToken1c68ec5e2ffa48b7b3694b6148d9dce4()

func newDroneToken1c68ec5e2ffa48b7b3694b6148d9dce4() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Drone",
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Drone},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCanBlockOnlyCreaturesWith,
							AffectedSource: true,
							BlockerRestriction: game.BlockerRestriction{
								Kind: game.BlockerRestrictionFlying,
							},
						},
					},
				},
			},
			OracleText: `
			Flying
			This token can block only creatures with flying.
		`,
		},
	}
}
