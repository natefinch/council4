package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HoldTheLine is the card definition for Hold the Line.
//
// Type: Instant
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	Blocking creatures get +7/+7 until end of turn.
var HoldTheLine = newHoldTheLine

func newHoldTheLine() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Hold the Line",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateBlocking}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     7,
									ToughnessDelta: 7,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Blocking creatures get +7/+7 until end of turn.
		`,
		},
	}
}
