package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KaitoSPursuit is the card definition for Kaito's Pursuit.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Target player discards two cards. Ninjas and Rogues you control gain menace until end of turn. (They can't be blocked except by two or more creatures.)
var KaitoSPursuit = newKaitoSPursuit

func newKaitoSPursuit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Kaito's Pursuit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
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
					},
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									Group: game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Ninja"), types.Sub("Rogue")}, Controller: game.ControllerYou}),
									AddKeywords: []game.Keyword{
										game.Menace,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target player discards two cards. Ninjas and Rogues you control gain menace until end of turn. (They can't be blocked except by two or more creatures.)
		`,
		},
	}
}
