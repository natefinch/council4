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

// ChainOfVapor is the card definition for Chain of Vapor.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Return target nonland permanent to its owner's hand. Then that permanent's controller may sacrifice a land of their choice. If the player does, they may copy this spell and may choose a new target for that copy.
var ChainOfVapor = newChainOfVapor

func newChainOfVapor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Chain of Vapor",
			ManaCost: opt.Val(cost.Mana{
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
						Primitive: game.MovePermanent{
							Object:      game.TargetPermanentReference(0),
							Destination: zone.Hand,
						},
					},
					{
						Primitive: game.Pay{
							Payment: game.ResolutionPayment{
								Prompt: "Sacrifice a land?",
								Payer:  opt.Val(game.AffectedTargetControllerReference(0)),
								AdditionalCosts: []cost.Additional{
									cost.Additional{
										Kind:               cost.AdditionalSacrifice,
										Text:               "sacrifice a land",
										Amount:             1,
										MatchPermanentType: true,
										PermanentType:      types.Land,
										CounterKind:        counter.PlusOnePlusOne,
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
			Return target nonland permanent to its owner's hand. Then that permanent's controller may sacrifice a land of their choice. If the player does, they may copy this spell and may choose a new target for that copy.
		`,
		},
	}
}
