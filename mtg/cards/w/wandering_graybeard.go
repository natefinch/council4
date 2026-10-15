package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WanderingGraybeard is the card definition for Wandering Graybeard.
//
// Type: Creature — Giant Wizard
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, you gain 4 life.
var WanderingGraybeard = newWanderingGraybeard

func newWanderingGraybeard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Wandering Graybeard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Wizard},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, you gain 4 life.",
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
								Text: "You may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, you gain 4 life.",
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LookAtLibraryTop{
											Player:        game.ControllerReference(),
											PublishLinked: game.LinkedKey("kinship-top"),
										},
										Optional: true,
									},
									game.Instruction{
										Primitive: game.Reveal{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "kinship-top",
											},
										},
										CardCondition: opt.Val(game.CardSelection{
											Card: game.CardReference{
												Kind:   game.CardReferenceLinked,
												LinkID: "kinship-top",
											},
											Selection: game.Selection{SharesCreatureTypeWithSource: true},
										}),
										Optional:      true,
										PublishResult: game.ResultKey("kinship-revealed"),
									},
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(4),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("kinship-revealed"),
											Accepted:  game.TriTrue,
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
			Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, you gain 4 life.
		`,
		},
	}
}
