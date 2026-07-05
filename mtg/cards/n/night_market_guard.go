package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NightMarketGuard is the card definition for Night Market Guard.
//
// Type: Artifact Creature — Construct
// Cost: {3}
//
// Oracle text:
//
//	This creature can block an additional creature each combat.
var NightMarketGuard = newNightMarketGuard()

func newNightMarketGuard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Night Market Guard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                 game.RuleEffectCanBlockAdditional,
							AffectedSource:       true,
							AdditionalBlockCount: 1,
						},
					},
				},
			},
			OracleText: `
			This creature can block an additional creature each combat.
		`,
		},
	}
}
