package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlayingFire is the card definition for Slaying Fire.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	Slaying Fire deals 3 damage to any target.
//	Adamant — If at least three red mana was spent to cast this spell, it deals 4 damage instead.
var SlayingFire = newSlayingFire

func newSlayingFire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Slaying Fire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:              true,
								SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Red, Count: 3},
							}),
						}),
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								SpellColorManaSpent: game.ColorManaSpendThreshold{Color: color.Red, Count: 3},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Slaying Fire deals 3 damage to any target.
			Adamant — If at least three red mana was spent to cast this spell, it deals 4 damage instead.
		`,
		},
	}
}
