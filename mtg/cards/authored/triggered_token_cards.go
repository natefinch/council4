package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	partyDudeOracleID          = "8df7a58c-053f-4ead-a778-2747718e5f10"
	sengirAutocratOracleID     = "0561043a-3896-4bb1-8ff5-ee674ddfea2a"
	whiteGloveGourmandOracleID = "3771c0d7-cea3-4c03-933d-f236f32ef2db"
)

var (
	PartyDude = cardset.Entry{
		OracleID: partyDudeOracleID, Name: "Party Dude", New: newPartyDude,
		Verification: "mtg/rules/triggered_token_cards_test.go: TestPartyDudeOwnText",
	}
	SengirAutocrat = cardset.Entry{
		OracleID: sengirAutocratOracleID, Name: "Sengir Autocrat", New: newSengirAutocrat,
		Verification: "mtg/rules/triggered_token_cards_test.go: TestSengirAutocratOwnText",
	}
	WhiteGloveGourmand = cardset.Entry{
		OracleID: whiteGloveGourmandOracleID, Name: "White Glove Gourmand", New: newWhiteGloveGourmand,
		Verification: "mtg/rules/triggered_token_cards_test.go: TestWhiteGloveGourmandOwnText",
	}
)

func newPartyDude() *game.CardDef {
	const text = "(Gain the next level as a sorcery to add its ability.)\n" +
		"When this Class enters, each player creates a Food token.\n" +
		"{1}{G}: Level 2\n" +
		"Whenever an artifact an opponent controls is put into a graveyard from the battlefield, draw a card.\n" +
		"{4}{G}: Level 3\n" +
		"Whenever one or more of your opponents are attacked, up to one target attacking creature gets +X/+X until end of turn, where X is the number of cards in your hand."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Party Dude",
			ManaCost:   opt.Val(cost.Mana{cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Class},
			OracleText: text,
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{1}{G}: Level 2", cost.Mana{cost.O(1), cost.G}, 2),
				classLevelAbility("{4}{G}: Level 3", cost.Mana{cost.O(4), cost.G}, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "When this Class enters, each player creates a Food token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.CreateToken{
							Amount:         game.Fixed(1),
							Source:         game.TokenDef(triggeredTokenFood),
							RecipientGroup: game.AllPlayersReference(),
						},
					}}}.Ability(),
				},
				{
					Text: "Whenever an artifact an opponent controls is put into a graveyard from the battlefield, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventZoneChanged,
							Controller:       game.TriggerControllerOpponent,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							MatchFromZone:    true,
							FromZone:         zone.Battlefield,
							MatchToZone:      true,
							ToZone:           zone.Graveyard,
						},
						InterveningCondition: opt.Val(game.Condition{SourceClassLevelAtLeast: 2}),
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
					}}}.Ability(),
				},
				{
					Text: "Whenever one or more of your opponents are attacked, up to one target attacking creature gets +X/+X until end of turn, where X is the number of cards in your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                    game.EventAttackerDeclared,
							Player:                   game.TriggerPlayerOpponent,
							AttackRecipient:          game.AttackRecipientPlayer,
							OneOrMore:                true,
							OneOrMorePerAttackTarget: true,
						},
						InterveningCondition: opt.Val(game.Condition{SourceClassLevelAtLeast: 3}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{{
							MinTargets: 0,
							MaxTargets: 1,
							Constraint: "up to one target attacking creature",
							Allow:      game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								CombatState:   game.CombatStateAttacking,
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.ModifyPT{
								Object:         game.TargetPermanentReference(0),
								PowerDelta:     game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountControllerHandSize}),
								ToughnessDelta: game.Dynamic(game.DynamicAmount{Kind: game.DynamicAmountControllerHandSize}),
								Duration:       game.DurationUntilEndOfTurn,
							},
						}},
					}.Ability(),
				},
			},
		},
	}
}

var sengirAutocratSerfToken = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Serf",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Serf},
		Power:     opt.Val(game.PT{Value: 0}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}

func newSengirAutocrat() *game.CardDef {
	const text = "When this creature enters, create three 0/1 black Serf creature tokens.\n" +
		"When this creature leaves the battlefield, exile all Serf tokens."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Sengir Autocrat",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: text,
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "When this creature enters, create three 0/1 black Serf creature tokens.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.CreateToken{
							Amount: game.Fixed(3),
							Source: game.TokenDef(sengirAutocratSerfToken),
						},
					}}}.Ability(),
				},
				{
					Text: "When this creature leaves the battlefield, exile all Serf tokens.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:         game.EventZoneChanged,
							Source:        game.TriggerSourceSelf,
							MatchFromZone: true,
							FromZone:      zone.Battlefield,
						},
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.MovePermanent{
							Group: game.BattlefieldGroup(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								SubtypesAny:   []types.Sub{types.Serf},
								TokenOnly:     true,
							}),
							Destination: zone.Exile,
						},
					}}}.Ability(),
				},
			},
		},
	}
}

var whiteGloveHumanSoldierToken = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Human Soldier",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Soldier},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
}

func newWhiteGloveGourmand() *game.CardDef {
	const text = "When this creature enters, create two 1/1 white Human Soldier creature tokens.\n" +
		"At the beginning of your end step, if another Human died under your control this turn, create a Food token."
	humanDiedThisTurn := game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentDied,
				Controller:       game.TriggerControllerYou,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Human}},
				ExcludeSelf:      true,
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:       "White Glove Gourmand",
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.W, cost.B}),
			Colors:     []color.Color{color.White, color.Black},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Noble},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: text,
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "When this creature enters, create two 1/1 white Human Soldier creature tokens.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.CreateToken{
							Amount: game.Fixed(2),
							Source: game.TokenDef(whiteGloveHumanSoldierToken),
						},
					}}}.Ability(),
				},
				{
					Text: "At the beginning of your end step, if another Human died under your control this turn, create a Food token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningCondition: opt.Val(humanDiedThisTurn),
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(triggeredTokenFood),
						},
					}}}.Ability(),
				},
			},
		},
	}
}

var triggeredTokenFood = &game.CardDef{
	CardFace: game.CardFace{
		Name:       "Food",
		Types:      []types.Card{types.Artifact},
		Subtypes:   []types.Sub{types.Food},
		OracleText: "{2}, {T}, Sacrifice this token: You gain 3 life.",
		ActivatedAbilities: []game.ActivatedAbility{{
			Text:     "{2}, {T}, Sacrifice this token: You gain 3 life.",
			ManaCost: opt.Val(cost.Mana{cost.O(2)}),
			AdditionalCosts: []cost.Additional{
				{Kind: cost.AdditionalTap, CounterKind: counter.PlusOnePlusOne},
				{Kind: cost.AdditionalSacrificeSource, Text: "Sacrifice this token", Amount: 1, CounterKind: counter.PlusOnePlusOne},
			},
			ZoneOfFunction: zone.Battlefield,
			Content: game.Mode{Sequence: []game.Instruction{{
				Primitive: game.GainLife{Amount: game.Fixed(3), Player: game.ControllerReference()},
			}}}.Ability(),
		}},
	},
}
