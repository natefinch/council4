package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RoilEruption is the card definition for Roil Eruption.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Kicker {5} (You may pay an additional {5} as you cast this spell.)
//	Roil Eruption deals 3 damage to any target. If this spell was kicked, it deals 5 damage instead.
var RoilEruption = newRoilEruption

func newRoilEruption() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Roil Eruption",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(5)}},
					},
				},
			},
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
								Negate:         true,
								SpellWasKicked: true,
							}),
						}),
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(5),
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
			Kicker {5} (You may pay an additional {5} as you cast this spell.)
			Roil Eruption deals 3 damage to any target. If this spell was kicked, it deals 5 damage instead.
		`,
		},
	}
}
