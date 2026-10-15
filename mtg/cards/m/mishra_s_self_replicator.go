package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MishraSSelfReplicator is the card definition for Mishra's Self-Replicator.
//
// Type: Artifact Creature — Assembly-Worker
// Cost: {5}
//
// Oracle text:
//
//	Whenever you cast a historic spell, you may pay {1}. If you do, create a token that's a copy of this creature. (Artifacts, legendaries, and Sagas are historic.)
var MishraSSelfReplicator = newMishraSSelfReplicator

func newMishraSSelfReplicator() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Mishra's Self-Replicator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.AssemblyWorker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a historic spell, you may pay {1}. If you do, create a token that's a copy of this creature. (Artifacts, legendaries, and Sagas are historic.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							RequireHistoric: true,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenCopyOf(game.TokenCopySpec{
												Source: game.TokenCopySourceObject,
												Object: game.SourcePermanentReference(),
											}),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("controller-paid"),
											Succeeded: game.TriTrue,
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Whenever you cast a historic spell, you may pay {1}. If you do, create a token that's a copy of this creature. (Artifacts, legendaries, and Sagas are historic.)
		`,
		},
	}
}
