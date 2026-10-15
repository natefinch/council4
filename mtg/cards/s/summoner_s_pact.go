package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SummonerSPact is the card definition for Summoner's Pact.
//
// Type: Instant
// Cost: {0}
//
// Oracle text:
//
//	Search your library for a green creature card, reveal it, put it into your hand, then shuffle.
//	At the beginning of your next upkeep, pay {2}{G}{G}. If you don't, you lose the game.
var SummonerSPact = newSummonerSPact

func newSummonerSPact() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Summoner's Pact",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Search{
							Player: game.ControllerReference(),
							Spec: game.SearchSpec{
								SourceZone:  zone.Library,
								Destination: zone.Hand,
								Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, ColorsAny: []color.Color{color.Green}},
								Reveal:      true,
							},
							Amount: game.Fixed(1),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextUpkeep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Pay{
														Payment: game.ResolutionPayment{
															Prompt: "Pay {2}{G}{G}?",
															ManaCost: opt.Val(cost.Mana{
																cost.O(2),
																cost.G,
																cost.G,
															}),
														},
													},
													PublishResult: game.ResultKey("pact-paid"),
												},
												game.Instruction{
													Primitive: game.PlayerLosesGame{
														Player: game.ControllerReference(),
													},
													ResultGate: opt.Val(game.InstructionResultGate{
														Key:       game.ResultKey("pact-paid"),
														Succeeded: game.TriFalse,
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
					},
				},
			}.Ability()),
			OracleText: `
			Search your library for a green creature card, reveal it, put it into your hand, then shuffle.
			At the beginning of your next upkeep, pay {2}{G}{G}. If you don't, you lose the game.
		`,
		},
	}
}
