package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SlipThroughSpace is the card definition for Slip Through Space.
//
// Type: Sorcery
// Cost: {U}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	Target creature can't be blocked this turn.
//	Draw a card.
var SlipThroughSpace = newSlipThroughSpace

func newSlipThroughSpace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Slip Through Space",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Types: []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
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
			Devoid (This card has no color.)
			Target creature can't be blocked this turn.
			Draw a card.
		`,
		},
	}
}
