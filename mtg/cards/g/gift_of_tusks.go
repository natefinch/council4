package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GiftOfTusks is the card definition for Gift of Tusks.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Until end of turn, target creature loses all abilities and becomes a green Elephant with base power and toughness 3/3.
var GiftOfTusks = newGiftOfTusks

func newGiftOfTusks() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Gift of Tusks",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
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
									SetSubtypes: []types.Sub{types.Elephant},
								},
								game.ContinuousEffect{
									Duration: game.DurationPermanent,
									Layer:    game.LayerPowerToughnessSet,
									SetPower: opt.Val(game.PT{
										Value: 3,
									}),
									SetToughness: opt.Val(game.PT{
										Value: 3,
									}),
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Until end of turn, target creature loses all abilities and becomes a green Elephant with base power and toughness 3/3.
		`,
		},
	}
}
