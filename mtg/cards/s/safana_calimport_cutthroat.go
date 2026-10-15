package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SafanaCalimportCutthroat is the card definition for Safana, Calimport Cutthroat.
//
// Type: Legendary Creature — Human Rogue
// Cost: {2}{B}
//
// Oracle text:
//
//	Menace
//	At the beginning of your end step, if you have the initiative, create a Treasure token. Create three of those tokens instead if you've completed a dungeon.
//	Choose a Background (You can have a Background as a second commander.)
var SafanaCalimportCutthroat = newSafanaCalimportCutthroat

func newSafanaCalimportCutthroat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Safana, Calimport Cutthroat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Rogue},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.ChooseABackgroundStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, if you have the initiative, create a Treasure token. Create three of those tokens instead if you've completed a dungeon.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
						InterveningIf: "if you have the initiative",
						InterveningCondition: opt.Val(game.Condition{
							Text:                    "if you have the initiative",
							SourceCounterKind:       counter.PlusOnePlusOne,
							ControllerHasInitiative: true,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(safanaCalimportCutthroatToken),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "if you've completed a dungeon",
												Negate:                      true,
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(3),
											Source: game.TokenDef(safanaCalimportCutthroatToken),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "if you've completed a dungeon",
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
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
			Menace
			At the beginning of your end step, if you have the initiative, create a Treasure token. Create three of those tokens instead if you've completed a dungeon.
			Choose a Background (You can have a Background as a second commander.)
		`,
		},
	}
}

var safanaCalimportCutthroatToken = newSafanaCalimportCutthroatToken()

func newSafanaCalimportCutthroatToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Treasure",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}, Sacrifice this artifact: Add one mana of any color.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Choose{
											Choice: game.ResolutionChoice{
												Kind:   game.ResolutionChoiceMana,
												Prompt: "Choose a color",
												Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
											},
											PublishChoice: game.ChoiceKey("oracle-mana-color"),
										},
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:     game.Fixed(1),
											ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
										},
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
	}
}
