package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChainLightning is the card definition for Chain Lightning.
//
// Type: Sorcery
// Cost: {R}
//
// Oracle text:
//
//	Chain Lightning deals 3 damage to any target. Then that player or that permanent's controller may pay {R}{R}. If the player does, they may copy this spell and may choose a new target for that copy.
var ChainLightning = newChainLightning

func newChainLightning() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Chain Lightning",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(3),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.Pay{
							Payment: game.ResolutionPayment{
								Prompt: "Pay {R}{R}?",
								Payer:  opt.Val(game.AffectedTargetControllerReference(0)),
								ManaCost: opt.Val(cost.Mana{
									cost.R,
									cost.R,
								}),
							},
						},
						PublishResult: game.ResultKey("copy-chain-paid"),
					},
					{
						Primitive: game.CopyStackObject{
							Object:              game.ResolvingStackObjectReference(),
							MayChooseNewTargets: true,
							Chooser:             opt.Val(game.AffectedTargetControllerReference(0)),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "copy-chain-paid",
							Succeeded: game.TriTrue,
						}),
						Optional:      true,
						OptionalActor: opt.Val(game.AffectedTargetControllerReference(0)),
					},
				},
			}.Ability()),
			OracleText: `
			Chain Lightning deals 3 damage to any target. Then that player or that permanent's controller may pay {R}{R}. If the player does, they may copy this spell and may choose a new target for that copy.
		`,
		},
	}
}
