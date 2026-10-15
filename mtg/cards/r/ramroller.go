package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ramroller is the card definition for Ramroller.
//
// Type: Artifact Creature — Juggernaut
// Cost: {3}
//
// Oracle text:
//
//	This creature attacks each combat if able.
//	This creature gets +2/+0 as long as you control another artifact.
var Ramroller = newRamroller

func newRamroller() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ramroller",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Juggernaut},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
				game.StaticAbility{
					Text: "This creature gets +2/+0 as long as you control another artifact.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control another artifact",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
						},
					},
				},
			},
			OracleText: `
			This creature attacks each combat if able.
			This creature gets +2/+0 as long as you control another artifact.
		`,
		},
	}
}
