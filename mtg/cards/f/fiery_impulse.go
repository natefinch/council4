package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FieryImpulse is the card definition for Fiery Impulse.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Fiery Impulse deals 2 damage to target creature.
//	Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Fiery Impulse deals 3 damage instead.
var FieryImpulse = newFieryImpulse

func newFieryImpulse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Fiery Impulse",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
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
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								ControllerGraveyardInstantOrSorceryCountAtLeast: 2,
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
								ControllerGraveyardInstantOrSorceryCountAtLeast: 2,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Fiery Impulse deals 2 damage to target creature.
			Spell mastery — If there are two or more instant and/or sorcery cards in your graveyard, Fiery Impulse deals 3 damage instead.
		`,
		},
	}
}
