package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BlueSunSTwilight is the card definition for Blue Sun's Twilight.
//
// Type: Sorcery
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	Gain control of target creature with mana value X or less. If X is 5 or more, create a token that's a copy of that creature.
var BlueSunSTwilight = newBlueSunSTwilight

func newBlueSunSTwilight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Blue Sun's Twilight",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:       1,
						MaxTargets:       1,
						Constraint:       "target creature with mana value X or less",
						Allow:            game.TargetAllowPermanent,
						Selection:        opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						ManaValueAtMostX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:         game.LayerControl,
									NewController: opt.Val(game.Player1),
								},
							},
							Duration: game.DurationPermanent,
						},
					},
					{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenCopyOf(game.TokenCopySpec{
								Source: game.TokenCopySourceObject,
								Object: game.TargetPermanentReference(0),
							}),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateSpellX, Op: compare.GreaterOrEqual, Value: 5}},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Gain control of target creature with mana value X or less. If X is 5 or more, create a token that's a copy of that creature.
		`,
		},
	}
}
