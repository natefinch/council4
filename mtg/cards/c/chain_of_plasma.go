package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ChainOfPlasma is the card definition for Chain of Plasma.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Chain of Plasma deals 3 damage to any target. Then that player or that permanent's controller may discard a card. If the player does, they may copy this spell and may choose a new target for that copy.
var ChainOfPlasma = newChainOfPlasma

func newChainOfPlasma() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Chain of Plasma",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
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
								Prompt: "Discard a card?",
								Payer:  opt.Val(game.AffectedTargetControllerReference(0)),
								AdditionalCosts: []cost.Additional{
									cost.Additional{
										Kind:        cost.AdditionalDiscard,
										Text:        "discard a card",
										Amount:      1,
										Source:      zone.Hand,
										CounterKind: counter.PlusOnePlusOne,
									},
								},
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
			Chain of Plasma deals 3 damage to any target. Then that player or that permanent's controller may discard a card. If the player does, they may copy this spell and may choose a new target for that copy.
		`,
		},
	}
}
