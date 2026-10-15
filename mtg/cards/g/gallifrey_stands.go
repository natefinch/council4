package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GallifreyStands is the card definition for Gallifrey Stands.
//
// Type: Legendary Enchantment
// Cost: {4}{W}{U}
//
// Oracle text:
//
//	When Gallifrey Stands enters, return all Doctor cards from your graveyard to your hand.
//	At the beginning of your upkeep, you may put a Doctor creature card from your hand onto the battlefield. Then if you control thirteen or more Doctors, you win the game.
var GallifreyStands = newGallifreyStands

func newGallifreyStands() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Gallifrey Stands",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Gallifrey Stands enters, return all Doctor cards from your graveyard to your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MassReturnFromGraveyard{
											Player:      game.ControllerReference(),
											Selection:   game.Selection{SubtypesAny: []types.Sub{types.Doctor}, Controller: game.ControllerYou},
											Destination: zone.Hand,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "At the beginning of your upkeep, you may put a Doctor creature card from your hand onto the battlefield. Then if you control thirteen or more Doctors, you win the game.",
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
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Hand,
											Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Doctor}},
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Battlefield,
											},
											Prompt: "Choose a card to put onto the battlefield",
										},
										Optional: true,
									},
									game.Instruction{
										Primitive: game.PlayerWinsGame{
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text: "if you control thirteen or more Doctors",
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{SubtypesAny: []types.Sub{types.Doctor}},
													MinCount:  13,
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
			When Gallifrey Stands enters, return all Doctor cards from your graveyard to your hand.
			At the beginning of your upkeep, you may put a Doctor creature card from your hand onto the battlefield. Then if you control thirteen or more Doctors, you win the game.
		`,
		},
	}
}
