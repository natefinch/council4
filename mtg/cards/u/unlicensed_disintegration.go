package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnlicensedDisintegration is the card definition for Unlicensed Disintegration.
//
// Type: Instant
// Cost: {1}{B}{R}
//
// Oracle text:
//
//	Destroy target creature. If you control an artifact, Unlicensed Disintegration deals 3 damage to that creature's controller.
var UnlicensedDisintegration = newUnlicensedDisintegration

func newUnlicensedDisintegration() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Unlicensed Disintegration",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
				cost.R,
			}),
			Colors: []color.Color{color.Black, color.Red},
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
						Primitive: game.Destroy{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetPermanentReference(0))),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Destroy target creature. If you control an artifact, Unlicensed Disintegration deals 3 damage to that creature's controller.
		`,
		},
	}
}
