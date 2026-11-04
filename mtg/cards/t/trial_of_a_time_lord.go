package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TrialOfATimeLord is the card definition for Trial of a Time Lord.
//
// Type: Enchantment — Saga
// Cost: {1}{W}{W}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)
//	I, II, III — Exile target nontoken creature an opponent controls until this Saga leaves the battlefield.
//	IV — Starting with you, each player votes for innocent or guilty. If guilty gets more votes, the owner of each card exiled with this Saga puts that card on the bottom of their library.
var TrialOfATimeLord = newTrialOfATimeLord

func newTrialOfATimeLord() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Trial of a Time Lord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.W,
			}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I, II, III — Exile target nontoken creature an opponent controls until this Saga leaves the battlefield.",
					Chapters: []int{1, 2, 3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target nontoken creature an opponent controls",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, NonToken: true, Controller: game.ControllerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:            game.TargetPermanentReference(0),
											Destination:       zone.Exile,
											PublishLinked:     game.LinkedKey("exile-until-leaves"),
											UntilSourceLeaves: true,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "IV — Starting with you, each player votes for innocent or guilty. If guilty gets more votes, the owner of each card exiled with this Saga puts that card on the bottom of their library.",
					Chapters: []int{4},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Vote{
											Options: []string{"innocent", "guilty"},
										},
										PublishResult: game.ResultKey("vote-result"),
									},
									game.Instruction{
										Primitive: game.PutLinkedExiledCardsInLibrary{
											LinkedKey: game.LinkedKey("exile-until-leaves"),
											Bottom:    true,
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key: game.ResultKey("vote-result"),
											AmountRange: opt.Val(game.IntRange{
												Min: -4,
												Max: -1,
											}),
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
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after IV.)
			I, II, III — Exile target nontoken creature an opponent controls until this Saga leaves the battlefield.
			IV — Starting with you, each player votes for innocent or guilty. If guilty gets more votes, the owner of each card exiled with this Saga puts that card on the bottom of their library.
		`,
		},
	}
}
