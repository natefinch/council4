package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ShivanFire is the card definition for Shivan Fire.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Kicker {4} (You may pay an additional {4} as you cast this spell.)
//	Shivan Fire deals 2 damage to target creature. If this spell was kicked, it deals 4 damage instead.
var ShivanFire = newShivanFire

func newShivanFire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Shivan Fire",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(4)}},
					},
				},
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
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:         true,
								SpellWasKicked: true,
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
								SpellWasKicked: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {4} (You may pay an additional {4} as you cast this spell.)
			Shivan Fire deals 2 damage to target creature. If this spell was kicked, it deals 4 damage instead.
		`,
		},
	}
}
