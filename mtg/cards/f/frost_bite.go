package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FrostBite is the card definition for Frost Bite.
//
// Type: Snow Instant
// Cost: {R}
//
// Oracle text:
//
//	Frost Bite deals 2 damage to target creature or planeswalker. If you control three or more snow permanents, it deals 3 damage instead.
var FrostBite = newFrostBite

func newFrostBite() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Frost Bite",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors:     []color.Color{color.Red},
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{Supertypes: []types.Super{types.Snow}},
									MinCount:  3,
								}),
							}),
						}),
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{Supertypes: []types.Super{types.Snow}},
									MinCount:  3,
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Frost Bite deals 2 damage to target creature or planeswalker. If you control three or more snow permanents, it deals 3 damage instead.
		`,
		},
	}
}
