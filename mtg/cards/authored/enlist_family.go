package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const enlistReminderText = "Enlist (As this creature attacks, you may tap a nonattacking creature you control without summoning sickness. When you do, add its power to this creature's until end of turn.)"

var (
	BenalishFaithbonder = cardset.Entry{
		OracleID: "278db664-4235-40f0-a926-3074a0e038c1",
		Name:     "Benalish Faithbonder",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyBenalishFaithbonder; printing 084d77df-a899-406f-9d79-e3fa3abeaebc",
		New: newBenalishFaithbonder,
	}
	BalduvianBerserker = cardset.Entry{
		OracleID: "3a7b1d1c-1724-4208-943b-650a5d773653",
		Name:     "Balduvian Berserker",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyBalduvianBerserker; printing c35bab2f-8464-4968-8c50-8807638d687b",
		New: newBalduvianBerserker,
	}
	CoalitionWarbrute = cardset.Entry{
		OracleID: "84d67d87-2696-4c2f-9f10-29d5832bfbb6",
		Name:     "Coalition Warbrute",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyCoalitionWarbrute; printing 554c422f-2a1b-47c0-9358-42b85cdbb84e",
		New: newCoalitionWarbrute,
	}
	YavimayaSteelcrusher = cardset.Entry{
		OracleID: "ad181ee7-756f-4002-9009-023ea6172991",
		Name:     "Yavimaya Steelcrusher",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyYavimayaSteelcrusher; printing 25175622-7b68-4916-adf2-3bf2d82e8356",
		New: newYavimayaSteelcrusher,
	}
	LinebreakerBaloth = cardset.Entry{
		OracleID: "c3b395b3-fea0-41d9-a3f2-6c90f89d8bdb",
		Name:     "Linebreaker Baloth",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyLinebreakerBaloth; printing 15cc59c7-6478-4f84-898a-7e0fed10f004",
		New: newLinebreakerBaloth,
	}
	ArgivianCavalier = cardset.Entry{
		OracleID: "c45090bd-a589-45a0-8388-025b8e0b3487",
		Name:     "Argivian Cavalier",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyArgivianCavalier; printing 57746e72-ba9f-4717-a4a2-6c6a039080c3",
		New: newArgivianCavalier,
	}
	BarkweaveCrusher = cardset.Entry{
		OracleID: "cbac75c6-6006-4025-b893-8828965b5f97",
		Name:     "Barkweave Crusher",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyBarkweaveCrusher; printing 41c62c55-7b04-4987-9047-1322c842bb59",
		New: newBarkweaveCrusher,
	}
	CoalitionSkyknight = cardset.Entry{
		OracleID: "dc763f20-73e7-4db7-9e98-6f419b8fba84",
		Name:     "Coalition Skyknight",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyCoalitionSkyknight; printing 8b12da45-aff4-406d-83a3-39ecaff6ff56",
		New: newCoalitionSkyknight,
	}
	HexbaneTortoise = cardset.Entry{
		OracleID: "f3b4097f-b84f-4645-81f1-018366e40eeb",
		Name:     "Hexbane Tortoise",
		Verification: "mtg/rules/enlist_family_test.go: " +
			"TestEnlistFamilyHexbaneTortoise; printing acd5a635-bcf5-4a05-b00b-bf40322823fe",
		New: newHexbaneTortoise,
	}
)

func newBenalishFaithbonder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Benalish Faithbonder",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W}),
			Colors:     []color.Color{color.White},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Human"), types.Sub("Cleric")},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: "Vigilance\n" + enlistReminderText,
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
				game.EnlistStaticBody,
			},
		},
	}
}

func newBalduvianBerserker() *game.CardDef {
	const diesText = "When this creature dies, it deals damage equal to its power to any target."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Balduvian Berserker",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Kor"), types.Sub("Berserker")},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			OracleText: enlistReminderText + "\n" + diesText,
			StaticAbilities: []game.StaticAbility{
				game.EnlistStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: diesText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:            game.EventPermanentDied,
						Source:           game.TriggerSourceSelf,
						SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
					},
				},
				Content: game.Mode{
					Text: diesText,
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					}},
					Sequence: []game.Instruction{{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountObjectPower,
								Multiplier: 1,
								Object:     game.SourcePermanentReference(),
							}),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					}},
				}.Ability(),
			}},
		},
	}
}

func newCoalitionWarbrute() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Coalition Warbrute",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Minotaur"), types.Sub("Berserker")},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			OracleText: enlistReminderText + "\nTrample",
			StaticAbilities: []game.StaticAbility{
				game.EnlistStaticBody,
				game.TrampleStaticBody,
			},
		},
	}
}

func newYavimayaSteelcrusher() *game.CardDef {
	const abilityText = "{1}, Sacrifice this creature: Destroy target artifact."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Yavimaya Steelcrusher",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Ape"), types.Sub("Warrior")},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: enlistReminderText + "\n" + abilityText,
			StaticAbilities: []game.StaticAbility{
				game.EnlistStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:            abilityText,
				ZoneOfFunction:  zone.Battlefield,
				ManaCost:        opt.Val(cost.Mana{cost.O(1)}),
				AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this creature"}},
				Content: game.Mode{
					Text: abilityText,
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target artifact",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
					}},
					Sequence: []game.Instruction{{
						Primitive: game.Destroy{Object: game.TargetPermanentReference(0)},
					}},
				}.Ability(),
			}},
		},
	}
}

func newLinebreakerBaloth() *game.CardDef {
	const evasionText = "This creature can't be blocked by creatures with power 2 or less."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Linebreaker Baloth",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.G, cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Beast")},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			OracleText: enlistReminderText + "\n" + evasionText,
			StaticAbilities: []game.StaticAbility{
				game.EnlistStaticBody,
				{
					Text: evasionText,
					RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
						AffectedSource: true,
						BlockerRestriction: game.BlockerRestriction{
							Kind:  game.BlockerRestrictionPowerLessOrEqual,
							Power: 2,
						},
					}},
				},
			},
		},
	}
}

func newArgivianCavalier() *game.CardDef {
	const enterText = "When this creature enters, create a 1/1 white Soldier creature token."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Argivian Cavalier",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.W}),
			Colors:     []color.Color{color.White},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Orc"), types.Sub("Knight")},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: enlistReminderText + "\n" + enterText,
			StaticAbilities: []game.StaticAbility{
				game.EnlistStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: enterText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhen,
					Pattern: game.TriggerPattern{
						Event:  game.EventPermanentEnteredBattlefield,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{
					Text: enterText,
					Sequence: []game.Instruction{{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(argivianCavalierSoldierToken),
						},
					}},
				}.Ability(),
			}},
		},
	}
}

var argivianCavalierSoldierToken = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Soldier",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Sub("Soldier")},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}

func newBarkweaveCrusher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Barkweave Crusher",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Elemental"), types.Sub("Warrior")},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			OracleText: enlistReminderText,
			StaticAbilities: []game.StaticAbility{
				game.EnlistStaticBody,
			},
		},
	}
}

func newCoalitionSkyknight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Coalition Skyknight",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.W}),
			Colors:     []color.Color{color.White},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Human"), types.Sub("Knight")},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: "Flying\n" + enlistReminderText,
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.EnlistStaticBody,
			},
		},
	}
}

func newHexbaneTortoise() *game.CardDef {
	const wardText = "Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Hexbane Tortoise",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Sub("Turtle")},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: wardText + "\n" + enlistReminderText,
			StaticAbilities: []game.StaticAbility{
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
				game.EnlistStaticBody,
			},
		},
	}
}
