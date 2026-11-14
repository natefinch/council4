package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	rosethornAcolyteOracleID = "639a8804-8513-4ba4-8683-ea1e2e5a9530"
	crystalDragonOracleID    = "56b0cb64-b886-413e-bb20-392bf12d3926"
	pollenShieldHareOracleID = "3ef02f2b-b611-40c3-a1d5-8b34989643cf"
	murderousRiderOracleID   = "1080c5b5-6651-4c6a-93e6-099fbe389e26"
	gumdropPoisonerOracleID  = "4f5ebbb8-f49b-488f-bb85-888aaa918bdd"
	questingDruidOracleID    = "555df32f-f8e9-49af-bc4a-1194ea325a97"

	seasonalRitualText  = "Add one mana of any color. (Then exile this card. You may cast the creature later from exile.)"
	robTheHoardText     = "Return target artifact, enchantment, or legendary card from your graveyard to your hand. (Then exile this card. You may cast the creature later from exile.)"
	hareRaisingText     = "Target creature you control gains vigilance and gets +X/+X until end of turn, where X is the number of creatures you control."
	swiftEndText        = "Destroy target creature or planeswalker. You lose 2 life. (Then exile this card. You may cast the creature later from exile.)"
	temptWithTreatsText = "Create a Food token. (Then exile this card. You may cast the creature later from exile.)"
	seekTheBeastText    = "Exile the top two cards of your library. Until your next end step, you may play those cards."
)

var (
	RosethornAcolyte = cardset.Entry{
		OracleID: rosethornAcolyteOracleID,
		Name:     "Rosethorn Acolyte",
		New:      newRosethornAcolyte,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestRosethornAcolyteSeasonalRitualScenario",
	}
	CrystalDragon = cardset.Entry{
		OracleID: crystalDragonOracleID,
		Name:     "Crystal Dragon",
		New:      newCrystalDragon,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestCrystalDragonRobTheHoardScenario",
	}
	PollenShieldHare = cardset.Entry{
		OracleID: pollenShieldHareOracleID,
		Name:     "Pollen-Shield Hare",
		New:      newPollenShieldHare,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestPollenShieldHareRaisingScenario",
	}
	MurderousRider = cardset.Entry{
		OracleID: murderousRiderOracleID,
		Name:     "Murderous Rider",
		New:      newMurderousRider,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestMurderousRiderSwiftEndScenario",
	}
	GumdropPoisoner = cardset.Entry{
		OracleID: gumdropPoisonerOracleID,
		Name:     "Gumdrop Poisoner",
		New:      newGumdropPoisoner,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestGumdropPoisonerTemptWithTreatsScenario",
	}
	QuestingDruid = cardset.Entry{
		OracleID: questingDruidOracleID,
		Name:     "Questing Druid",
		New:      newQuestingDruid,
		Verification: "mtg/rules/adventure_cards_test.go: " +
			"TestQuestingDruidSeekTheBeastScenario",
	}
)

func newRosethornAcolyte() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:          "Rosethorn Acolyte",
			ManaCost:      opt.Val(cost.Mana{cost.O(2), cost.G}),
			Colors:        []color.Color{color.Green},
			Types:         []types.Card{types.Creature},
			Subtypes:      []types.Sub{types.Elf, types.Druid},
			Power:         opt.Val(game.PT{Value: 2}),
			Toughness:     opt.Val(game.PT{Value: 3}),
			OracleText:    "{T}: Add one mana of any color.",
			ManaAbilities: []game.ManaAbility{game.TapAnyColorManaAbility()},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Seasonal Ritual",
			ManaCost:   opt.Val(cost.Mana{cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Sorcery},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: seasonalRitualText,
			SpellAbility: opt.Val(game.Mode{
				Text: seasonalRitualText,
				Sequence: []game.Instruction{
					{Primitive: game.Choose{
						Choice: game.ResolutionChoice{
							Kind:   game.ResolutionChoiceMana,
							Prompt: "Choose a color",
							Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
						},
						PublishChoice: game.ChoiceKey("seasonal-ritual-color"),
					}},
					{Primitive: game.AddMana{
						Amount:     game.Fixed(1),
						ChoiceFrom: game.ChoiceKey("seasonal-ritual-color"),
					}},
				},
			}.Ability()),
		}),
	}
}

func newCrystalDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:            "Crystal Dragon",
			ManaCost:        opt.Val(cost.Mana{cost.O(4), cost.W, cost.W}),
			Colors:          []color.Color{color.White},
			Types:           []types.Card{types.Creature},
			Subtypes:        []types.Sub{types.Dragon},
			Power:           opt.Val(game.PT{Value: 4}),
			Toughness:       opt.Val(game.PT{Value: 4}),
			OracleText:      "Flying, vigilance",
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody, game.VigilanceStaticBody},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Rob the Hoard",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:     []color.Color{color.White},
			Types:      []types.Card{types.Sorcery},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: robTheHoardText,
			SpellAbility: opt.Val(game.Mode{
				Text: robTheHoardText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target artifact, enchantment, or legendary card from your graveyard",
					Allow:      game.TargetAllowCard,
					TargetZone: zone.Graveyard,
					Selection: opt.Val(game.Selection{
						Controller: game.ControllerYou,
						AnyOf: []game.Selection{
							{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}},
							{Supertypes: []types.Super{types.Legendary}},
						},
					}),
				}},
				Sequence: []game.Instruction{{Primitive: game.MoveCard{
					Card:        game.CardReference{Kind: game.CardReferenceTarget},
					FromZone:    zone.Graveyard,
					Destination: zone.Hand,
				}}},
			}.Ability()),
		}),
	}
}

func newPollenShieldHare() *game.CardDef {
	creaturesYouControl := game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		Controller:    game.ControllerYou,
	}
	creaturesYouControlGroup := game.BattlefieldGroup(creaturesYouControl)
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:     "Pollen-Shield Hare",
			ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:   []color.Color{color.White},
			Types:    []types.Card{types.Creature},
			Subtypes: []types.Sub{types.Rabbit},
			Power:    opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{
				Value: 2,
			}),
			OracleText: "Creature tokens you control get +1/+1.",
			StaticAbilities: []game.StaticAbility{{
				Text: "Creature tokens you control get +1/+1.",
				ContinuousEffects: []game.ContinuousEffect{{
					Duration: game.DurationPermanent,
					Layer:    game.LayerPowerToughnessModify,
					Group: game.BattlefieldGroup(game.Selection{
						RequiredTypes: []types.Card{types.Creature},
						Controller:    game.ControllerYou,
						TokenOnly:     true,
					}),
					PowerDelta:     1,
					ToughnessDelta: 1,
				}},
			}},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Hare Raising",
			ManaCost:   opt.Val(cost.Mana{cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Sorcery},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: hareRaisingText,
			SpellAbility: opt.Val(game.Mode{
				Text: hareRaisingText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature you control",
					Allow:      game.TargetAllowPermanent,
					Selection:  opt.Val(creaturesYouControl),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance},
						}},
						Duration: game.DurationUntilEndOfTurn,
					}},
					{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0),
						PowerDelta: game.Dynamic(game.DynamicAmount{
							Kind:  game.DynamicAmountCountSelector,
							Group: creaturesYouControlGroup,
						}),
						ToughnessDelta: game.Dynamic(game.DynamicAmount{
							Kind:  game.DynamicAmountCountSelector,
							Group: creaturesYouControlGroup,
						}),
						Duration: game.DurationUntilEndOfTurn,
					}},
				},
			}.Ability()),
		}),
	}
}

func newMurderousRider() *game.CardDef {
	const riderText = "Lifelink\nWhen this creature dies, put it on the bottom of its owner's library."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:            "Murderous Rider",
			ManaCost:        opt.Val(cost.Mana{cost.O(1), cost.B, cost.B}),
			Colors:          []color.Color{color.Black},
			Types:           []types.Card{types.Creature},
			Subtypes:        []types.Sub{types.Zombie, types.Knight},
			Power:           opt.Val(game.PT{Value: 2}),
			Toughness:       opt.Val(game.PT{Value: 3}),
			OracleText:      riderText,
			StaticAbilities: []game.StaticAbility{game.LifelinkStaticBody},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: "When this creature dies, put it on the bottom of its owner's library.",
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentDied,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{
					Text: "Put Murderous Rider on the bottom of its owner's library",
					Sequence: []game.Instruction{{Primitive: game.MoveCard{
						Card:              game.CardReference{Kind: game.CardReferenceSource},
						FromZone:          zone.Graveyard,
						Destination:       zone.Library,
						DestinationBottom: true,
					}}},
				}.Ability(),
			}},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Swift End",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.B, cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Instant},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: swiftEndText,
			SpellAbility: opt.Val(game.Mode{
				Text: swiftEndText,
				Targets: []game.TargetSpec{{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "target creature or planeswalker",
					Allow:      game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker},
					}),
				}},
				Sequence: []game.Instruction{
					{Primitive: game.Destroy{Object: game.TargetPermanentReference(0)}},
					{Primitive: game.LoseLife{Amount: game.Fixed(2), Player: game.ControllerReference()}},
				},
			}.Ability()),
		}),
	}
}

func newGumdropPoisoner() *game.CardDef {
	const poisonerText = "Lifelink\nWhen this creature enters, up to one target creature gets -X/-X until end of turn, where X is the amount of life you gained this turn."
	lifeGained := game.DynamicAmount{Kind: game.DynamicAmountLifeGainedThisTurn, Multiplier: -1}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:            "Gumdrop Poisoner",
			ManaCost:        opt.Val(cost.Mana{cost.O(2), cost.B}),
			Colors:          []color.Color{color.Black},
			Types:           []types.Card{types.Creature},
			Subtypes:        []types.Sub{types.Human, types.Warlock},
			Power:           opt.Val(game.PT{Value: 3}),
			Toughness:       opt.Val(game.PT{Value: 2}),
			OracleText:      poisonerText,
			StaticAbilities: []game.StaticAbility{game.LifelinkStaticBody},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: "When this creature enters, up to one target creature gets -X/-X until end of turn, where X is the amount of life you gained this turn.",
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{
					Text: "Target creature gets -X/-X until end of turn",
					Targets: []game.TargetSpec{{
						MinTargets: 0,
						MaxTargets: 1,
						Constraint: "up to one target creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
					}},
					Sequence: []game.Instruction{{Primitive: game.ModifyPT{
						Object:         game.TargetPermanentReference(0),
						PowerDelta:     game.Dynamic(lifeGained),
						ToughnessDelta: game.Dynamic(lifeGained),
						Duration:       game.DurationUntilEndOfTurn,
					}}},
				}.Ability(),
			}},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Tempt with Treats",
			ManaCost:   opt.Val(cost.Mana{cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Instant},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: temptWithTreatsText,
			SpellAbility: opt.Val(game.Mode{
				Text: temptWithTreatsText,
				Sequence: []game.Instruction{{Primitive: game.CreateToken{
					Amount: game.Fixed(1),
					Source: game.TokenDef(triggeredTokenFood),
				}}},
			}.Ability()),
		}),
	}
}

func newQuestingDruid() *game.CardDef {
	const druidText = "Whenever you cast a spell that's white, blue, black, or red, put a +1/+1 counter on this creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green, color.Red),
		Layout:        game.LayoutAdventure,
		CardFace: game.CardFace{
			Name:       "Questing Druid",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Druid},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			OracleText: druidText,
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: druidText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventSpellCast,
						Player: game.TriggerPlayerYou,
						CardSelection: game.Selection{
							ColorsAny: []color.Color{color.White, color.Blue, color.Black, color.Red},
						},
					},
				},
				Content: game.Mode{
					Text: druidText,
					Sequence: []game.Instruction{{Primitive: game.AddCounter{
						Object:      game.SourcePermanentReference(),
						Amount:      game.Fixed(1),
						CounterKind: counter.PlusOnePlusOne,
					}}},
				}.Ability(),
			}},
		},
		Alternate: opt.Val(game.CardFace{
			Name:       "Seek the Beast",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Instant},
			Subtypes:   []types.Sub{types.Adventure},
			OracleText: seekTheBeastText,
			SpellAbility: opt.Val(game.Mode{
				Text: seekTheBeastText,
				Sequence: []game.Instruction{{Primitive: game.ImpulseExile{
					Player:   game.ControllerReference(),
					Amount:   game.Fixed(2),
					Duration: game.DurationUntilYourNextEndStep,
				}}},
			}.Ability()),
		}),
	}
}
