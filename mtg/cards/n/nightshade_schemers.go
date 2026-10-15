package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NightshadeSchemers is the card definition for Nightshade Schemers.
//
// Type: Creature — Faerie Wizard
// Cost: {4}{B}
//
// Oracle text:
//
//	Flying
//	Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, each opponent loses 2 life.
var NightshadeSchemers = newNightshadeSchemers

func newNightshadeSchemers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nightshade Schemers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, each opponent loses 2 life.",
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
								Text: "You may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, each opponent loses 2 life.",
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
										Primitive: game.LoseLife{
											Amount:      game.Fixed(2),
											PlayerGroup: game.OpponentsReference(),
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
			Flying
			Kinship — At the beginning of your upkeep, you may look at the top card of your library. If it shares a creature type with this creature, you may reveal it. If you do, each opponent loses 2 life.
		`,
		},
	}
}
