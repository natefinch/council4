package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Snakeform is the card definition for Snakeform.
//
// Type: Instant
// Cost: {2}{G/U}
//
// Oracle text:
//
//	Until end of turn, target creature loses all abilities and becomes a green Snake with base power and toughness 1/1.
//	Draw a card.
var Snakeform = newSnakeform

func newSnakeform() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Snakeform",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.G, mana.U),
			}),
			Colors: []color.Color{color.Green, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:           game.DurationPermanent,
									Layer:              game.LayerAbility,
									RemoveAllAbilities: true,
								},
								game.ContinuousEffect{
									Duration:  game.DurationPermanent,
									Layer:     game.LayerColor,
									SetColors: []color.Color{color.Green},
								},
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Layer:       game.LayerType,
									SetTypes:    []types.Card{types.Creature},
									SetSubtypes: []types.Sub{types.Snake},
								},
								game.ContinuousEffect{
									Duration: game.DurationPermanent,
									Layer:    game.LayerPowerToughnessSet,
									SetPower: opt.Val(game.PT{
										Value: 1,
									}),
									SetToughness: opt.Val(game.PT{
										Value: 1,
									}),
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, target creature loses all abilities and becomes a green Snake with base power and toughness 1/1.
			Draw a card.
		`,
		},
	}
}
