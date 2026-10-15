package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EnterTheEnigma is the card definition for Enter the Enigma.
//
// Type: Sorcery
// Cost: {U}
//
// Oracle text:
//
//	Target creature can't be blocked this turn.
//	Draw a card.
var EnterTheEnigma = newEnterTheEnigma

func newEnterTheEnigma() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Enter the Enigma",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
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
						Primitive: game.ApplyRule{
							Object: opt.Val(game.TargetPermanentReference(0)),
							RuleEffects: []game.RuleEffect{
								game.RuleEffect{
									Kind:     game.RuleEffectCantBeBlocked,
									Duration: game.DurationPermanent,
								},
							},
							Duration: game.DurationThisTurn,
						},
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Target creature can't be blocked this turn.
			Draw a card.
		`,
		},
	}
}
