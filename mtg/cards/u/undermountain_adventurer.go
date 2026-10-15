package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UndermountainAdventurer is the card definition for Undermountain Adventurer.
//
// Type: Creature — Giant Warrior
// Cost: {3}{G}
//
// Oracle text:
//
//	Vigilance
//	When this creature enters, you take the initiative.
//	{T}: Add {G}{G}. If you've completed a dungeon, add six {G} instead.
var UndermountainAdventurer = newUndermountainAdventurer

func newUndermountainAdventurer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Undermountain Adventurer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {G}{G}. If you've completed a dungeon, add six {G} instead.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
												Negate:                      true,
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
												Negate:                      true,
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
												SourceCounterKind:           counter.PlusOnePlusOne,
												ControllerCompletedADungeon: true,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.G,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                        "If you've completed a dungeon",
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, you take the initiative.",
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
										Primitive: game.TakeInitiative{
											Player: game.ControllerReference(),
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
			OracleText: `
			Vigilance
			When this creature enters, you take the initiative.
			{T}: Add {G}{G}. If you've completed a dungeon, add six {G} instead.
		`,
		},
	}
}
