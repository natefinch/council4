package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DinosaurStampede is the card definition for Dinosaur Stampede.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	Attacking creatures get +2/+0 until end of turn. Dinosaurs you control gain trample until end of turn.
var DinosaurStampede = newDinosaurStampede()

func newDinosaurStampede() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Dinosaur Stampede",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:      game.LayerPowerToughnessModify,
									Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, CombatState: game.CombatStateAttacking}),
									PowerDelta: 2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Dinosaur")}, Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Trample,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Attacking creatures get +2/+0 until end of turn. Dinosaurs you control gain trample until end of turn.
		`,
		},
	}
}
