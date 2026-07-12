package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Probe is the card definition for Probe.
//
// Type: Sorcery
// Cost: {2}{U}
//
// Oracle text:
//
//	Kicker {1}{B} (You may pay an additional {1}{B} as you cast this spell.)
//	Draw three cards, then discard two cards. If this spell was kicked, target player discards two cards.
var Probe = newProbe

func newProbe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Probe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{Cost: cost.Mana{cost.O(1), cost.B}},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target player",
						Allow:      game.TargetAllowPlayer,
						Gate:       game.TargetGateSpellKicked,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(3),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(2),
							Player: game.TargetPlayerReference(0),
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
			Kicker {1}{B} (You may pay an additional {1}{B} as you cast this spell.)
			Draw three cards, then discard two cards. If this spell was kicked, target player discards two cards.
		`,
		},
	}
}
