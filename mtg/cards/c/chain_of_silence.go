package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChainOfSilence is the card definition for Chain of Silence.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Prevent all damage target creature would deal this turn. That creature's controller may sacrifice a land of their choice. If the player does, they may copy this spell and may choose a new target for that copy.
var ChainOfSilence = newChainOfSilence

func newChainOfSilence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Chain of Silence",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
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
						Primitive: game.PreventDamage{
							Object:   game.TargetPermanentReference(0),
							All:      true,
							BySource: true,
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
			Prevent all damage target creature would deal this turn. That creature's controller may sacrifice a land of their choice. If the player does, they may copy this spell and may choose a new target for that copy.
		`,
		},
	}
}
