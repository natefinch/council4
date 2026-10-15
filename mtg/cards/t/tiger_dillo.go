package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TigerDillo is the card definition for Tiger-Dillo.
//
// Type: Creature — Cat Armadillo
// Cost: {1}{R}
//
// Oracle text:
//
//	This creature can't attack or block unless you control another creature with power 4 or greater.
var TigerDillo = newTigerDillo

func newTigerDillo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tiger-Dillo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Armadillo},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature can't attack or block unless you control another creature with power 4 or greater.",
					Condition: opt.Val(game.Condition{
						Text:   "unless you control another creature with power 4 or greater",
						Negate: true,
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true, Power: opt.Val(compare.Int{
								Op:    compare.GreaterOrEqual,
								Value: 4,
							})},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantAttack,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
						},
					},
				},
			},
			OracleText: `
			This creature can't attack or block unless you control another creature with power 4 or greater.
		`,
		},
	}
}
