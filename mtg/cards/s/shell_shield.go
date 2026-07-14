package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShellShield is the card definition for Shell Shield.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Kicker {1} (You may pay an additional {1} as you cast this spell.)
//	Target creature you control gets +0/+3 until end of turn. If this spell was kicked, that creature also gains hexproof until end of turn. (It can't be the target of spells or abilities your opponents control.)
var ShellShield = newShellShield

func newShellShield() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Shell Shield",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(1)}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature you control",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:         game.TargetPermanentReference(0),
							PowerDelta:     game.Fixed(0),
							ToughnessDelta: game.Fixed(3),
							Duration:       game.DurationUntilEndOfTurn,
							PublishLinked:  game.LinkedKey("gain-keyword-1"),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.LinkedObjectReference("gain-keyword-1")),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Hexproof,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellWasKicked: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {1} (You may pay an additional {1} as you cast this spell.)
			Target creature you control gets +0/+3 until end of turn. If this spell was kicked, that creature also gains hexproof until end of turn. (It can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
