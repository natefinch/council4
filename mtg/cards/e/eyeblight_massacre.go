package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EyeblightMassacre is the card definition for Eyeblight Massacre.
//
// Type: Sorcery
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Non-Elf creatures get -2/-2 until end of turn.
var EyeblightMassacre = newEyeblightMassacre

func newEyeblightMassacre() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Eyeblight Massacre",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:          game.LayerPowerToughnessModify,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedSubtype: types.Sub("Elf")}),
									PowerDelta:     -2,
									ToughnessDelta: -2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Non-Elf creatures get -2/-2 until end of turn.
		`,
		},
	}
}
