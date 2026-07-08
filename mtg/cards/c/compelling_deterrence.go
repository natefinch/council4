package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CompellingDeterrence is the card definition for Compelling Deterrence.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Return target nonland permanent to its owner's hand. Then that player discards a card if you control a Zombie.
var CompellingDeterrence = newCompellingDeterrence

func newCompellingDeterrence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Compelling Deterrence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target nonland permanent",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Bounce{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Zombie")}},
								}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Return target nonland permanent to its owner's hand. Then that player discards a card if you control a Zombie.
		`,
		},
	}
}
