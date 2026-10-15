package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ScrollOfGriselbrand is the card definition for Scroll of Griselbrand.
//
// Type: Artifact
// Cost: {1}
//
// Oracle text:
//
//	{1}, Sacrifice this artifact: Target opponent discards a card. If you control a Demon, that player loses 3 life.
var ScrollOfGriselbrand = newScrollOfGriselbrand

func newScrollOfGriselbrand() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Scroll of Griselbrand",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}, Sacrifice this artifact: Target opponent discards a card. If you control a Demon, that player loses 3 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this artifact",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "Target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
											Player: game.TargetPlayerReference(0),
										},
									},
									game.Instruction{
										Primitive: game.LoseLife{
											Amount: game.Fixed(3),
											Player: game.TargetPlayerReference(0),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text: "If you control a Demon",
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{SubtypesAny: []types.Sub{types.Demon}},
												}),
												SourceCounterKind: counter.PlusOnePlusOne,
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
			{1}, Sacrifice this artifact: Target opponent discards a card. If you control a Demon, that player loses 3 life.
		`,
		},
	}
}
