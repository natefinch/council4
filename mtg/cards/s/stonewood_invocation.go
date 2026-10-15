package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StonewoodInvocation is the card definition for Stonewood Invocation.
//
// Type: Instant
// Cost: {3}{G}
//
// Oracle text:
//
//	Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
//	Target creature gets +5/+5 and gains shroud until end of turn. (It can't be the target of spells or abilities.)
var StonewoodInvocation = newStonewoodInvocation

func newStonewoodInvocation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Stonewood Invocation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.SplitSecondStaticBody,
			},
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
									Duration:       game.DurationPermanent,
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     5,
									ToughnessDelta: 5,
								},
								game.ContinuousEffect{
									Duration:    game.DurationPermanent,
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Shroud},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Split second (As long as this spell is on the stack, players can't cast spells or activate abilities that aren't mana abilities.)
			Target creature gets +5/+5 and gains shroud until end of turn. (It can't be the target of spells or abilities.)
		`,
		},
	}
}
