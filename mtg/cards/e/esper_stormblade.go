package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EsperStormblade is the card definition for Esper Stormblade.
//
// Type: Artifact Creature — Vedalken Wizard
// Cost: {W/B}{U}
//
// Oracle text:
//
//	As long as you control another multicolored permanent, this creature gets +1/+1 and has flying.
var EsperStormblade = newEsperStormblade

func newEsperStormblade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Esper Stormblade",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.W, mana.B),
				cost.U,
			}),
			Colors:    []color.Color{color.Black, color.Blue, color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Vedalken, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control another multicolored permanent, this creature gets +1/+1 and has flying.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control another multicolored permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{Multicolored: true, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Flying},
						},
					},
				},
			},
			OracleText: `
			As long as you control another multicolored permanent, this creature gets +1/+1 and has flying.
		`,
		},
	}
}
