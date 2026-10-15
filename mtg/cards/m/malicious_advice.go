package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MaliciousAdvice is the card definition for Malicious Advice.
//
// Type: Instant
// Cost: {X}{U}{B}
//
// Oracle text:
//
//	Tap X target artifacts, creatures, and/or lands. You lose X life.
var MaliciousAdvice = newMaliciousAdvice

func newMaliciousAdvice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Malicious Advice",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets:   0,
						MaxTargets:   20,
						Constraint:   "x target artifacts, creatures, and/or lands",
						Allow:        game.TargetAllowPermanent,
						Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Land}}),
						CountEqualsX: true,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Tap{
							Object: game.AllTargetPermanentsReference(0),
						},
					},
					{
						Primitive: game.LoseLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Tap X target artifacts, creatures, and/or lands. You lose X life.
		`,
		},
	}
}
