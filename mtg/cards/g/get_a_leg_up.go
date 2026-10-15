package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GetALegUp is the card definition for Get a Leg Up.
//
// Type: Instant
// Cost: {G}
//
// Oracle text:
//
//	Until end of turn, target creature gets +1/+1 for each creature you control and gains reach.
var GetALegUp = newGetALegUp

func newGetALegUp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Get a Leg Up",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors: []color.Color{color.Green},
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
									Duration: game.DurationPermanent,
									Layer:    game.LayerPowerToughnessModify,
									PowerDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountCountSelector,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
										Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
									ToughnessDeltaDynamic: opt.Val(game.DynamicAmount{
										Kind:        game.DynamicAmountCountSelector,
										Multiplier:  1,
										CounterKind: counter.PlusOnePlusOne,
										Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									}),
								},
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Reach},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, target creature gets +1/+1 for each creature you control and gains reach.
		`,
		},
	}
}
