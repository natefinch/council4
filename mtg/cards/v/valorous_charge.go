package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ValorousCharge is the card definition for Valorous Charge.
//
// Type: Sorcery
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	White creatures get +2/+0 until end of turn.
var ValorousCharge = newValorousCharge

func newValorousCharge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Valorous Charge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:      game.LayerPowerToughnessModify,
									Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.White}}),
									PowerDelta: 2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			White creatures get +2/+0 until end of turn.
		`,
		},
	}
}
