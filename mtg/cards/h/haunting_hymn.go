package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HauntingHymn is the card definition for Haunting Hymn.
//
// Type: Instant
// Cost: {4}{B}{B}
//
// Oracle text:
//
//	Target player discards two cards. If you cast this spell during your main phase, that player discards four cards instead.
var HauntingHymn = newHauntingHymn

func newHauntingHymn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Haunting Hymn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target player",
						Allow:      game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:                        true,
								CastDuringControllerMainPhase: true,
							}),
						}),
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(4),
							Player: game.TargetPlayerReference(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								CastDuringControllerMainPhase: true,
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Target player discards two cards. If you cast this spell during your main phase, that player discards four cards instead.
		`,
		},
	}
}
