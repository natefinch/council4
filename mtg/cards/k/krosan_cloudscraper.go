package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KrosanCloudscraper is the card definition for Krosan Cloudscraper.
//
// Type: Creature — Beast Mutant
// Cost: {7}{G}{G}{G}
//
// Oracle text:
//
//	At the beginning of your upkeep, sacrifice this creature unless you pay {G}{G}.
//	Morph {7}{G}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
var KrosanCloudscraper = newKrosanCloudscraper

func newKrosanCloudscraper() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Krosan Cloudscraper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast, types.Mutant},
			Power:     opt.Val(game.PT{Value: 13}),
			Toughness: opt.Val(game.PT{Value: 13}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Morph {7}{G}{G}",
					KeywordAbilities: []game.KeywordAbility{
						game.MorphKeyword{
							Cost: cost.Mana{
								cost.O(7),
								cost.G,
								cost.G,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, sacrifice this creature unless you pay {G}{G}.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {G}{G}?",
												ManaCost: opt.Val(cost.Mana{
													cost.G,
													cost.G,
												}),
											},
										},
										PublishResult: game.ResultKey("sacrifice-unless-paid"),
									},
									game.Instruction{
										Primitive: game.Sacrifice{
											Object: game.SourcePermanentReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("sacrifice-unless-paid"),
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
			OracleText: `
			At the beginning of your upkeep, sacrifice this creature unless you pay {G}{G}.
			Morph {7}{G}{G} (You may cast this card face down as a 2/2 creature for {3}. Turn it face up any time for its morph cost.)
		`,
		},
	}
}
