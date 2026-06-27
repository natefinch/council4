package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VortexRunner is the card definition for Vortex Runner.
//
// Type: Creature — Human Wizard
// Cost: {2}{U}
//
// Oracle text:
//
//	As long as you control eight or more lands, this creature gets +1/+0 and can't be blocked.
var VortexRunner = newVortexRunner()

func newVortexRunner() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vortex Runner",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
							MinCount:  8,
						}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     1,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlocked,
							AffectedSource: true,
						},
					},
				},
			},
			OracleText: `
			As long as you control eight or more lands, this creature gets +1/+0 and can't be blocked.
		`,
		},
	}
}
