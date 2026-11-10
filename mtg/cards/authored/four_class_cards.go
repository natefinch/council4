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
	fortuneTellersTalentOracleID = "1b430f67-3686-4452-8594-b060f1a5a04e"
	rangerClassOracleID          = "d1586808-efc2-469a-8384-ba4acf231e69"
	warlockClassOracleID         = "766c2c67-d1ee-44ff-b9bb-355361646ef1"
	artistsTalentOracleID        = "ea734863-5793-49d6-b10e-109432aec0a3"
)

var (
	FortuneTellersTalent = cardset.Entry{
		OracleID: fortuneTellersTalentOracleID,
		Name:     "Fortune Teller's Talent",
		Verification: "mtg/rules/four_class_cards_test.go: " +
			"TestFortuneTellersTalentOwnText",
		New: newFortuneTellersTalent,
	}
	RangerClass = cardset.Entry{
		OracleID: rangerClassOracleID,
		Name:     "Ranger Class",
		Verification: "mtg/rules/four_class_cards_test.go: " +
			"TestRangerClassOwnText",
		New: newRangerClass,
	}
	WarlockClass = cardset.Entry{
		OracleID: warlockClassOracleID,
		Name:     "Warlock Class",
		Verification: "mtg/rules/four_class_cards_test.go: " +
			"TestWarlockClassOwnText",
		New: newWarlockClass,
	}
	ArtistsTalent = cardset.Entry{
		OracleID: artistsTalentOracleID,
		Name:     "Artist's Talent",
		Verification: "mtg/rules/four_class_cards_test.go: " +
			"TestArtistsTalentOwnText",
		New: newArtistsTalent,
	}
)

func newFortuneTellersTalent() *game.CardDef {
	const text = "(Gain the next level as a sorcery to add its ability.)\n" +
		"You may look at the top card of your library any time.\n" +
		"{3}{U}: Level 2\n" +
		"As long as you've cast a spell this turn, you may play cards from the top of your library.\n" +
		"{2}{U}: Level 3\n" +
		"Spells you cast from anywhere other than your hand cost {2} less to cast."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Fortune Teller's Talent",
			ManaCost:   opt.Val(cost.Mana{cost.U}),
			Colors:     []color.Color{color.Blue},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Class},
			OracleText: text,
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{3}{U}: Level 2", cost.Mana{cost.O(3), cost.U}, 2),
				classLevelAbility("{2}{U}: Level 3", cost.Mana{cost.O(2), cost.U}, 3),
			},
			StaticAbilities: []game.StaticAbility{
				game.LookAtTopCardAnyTimeStaticBody,
				{
					Text: "As long as you've cast a spell this turn, you may play cards from the top of your library.",
					Condition: opt.Val(game.Condition{
						EventHistory: opt.Val(game.EventHistoryCondition{
							Pattern: game.TriggerPattern{
								Event:  game.EventSpellCast,
								Player: game.TriggerPlayerYou,
							},
							Window: game.EventHistoryCurrentTurn,
						}),
					}),
					RuleEffects: []game.RuleEffect{
						{
							Kind:           game.RuleEffectCastSpellsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
						},
						{
							Kind:           game.RuleEffectPlayLandsFromZone,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							PermanentTypes: []types.Card{types.Land},
							CastFromZone:   zone.Library,
							TopCardOnly:    true,
						},
					},
				},
				{
					Text: "Spells you cast from anywhere other than your hand cost {2} less to cast.",
					Condition: opt.Val(game.Condition{
						SourceClassLevelAtLeast: 3,
					}),
					RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectCostModifier,
						Duration:       game.DurationPermanent,
						AffectedPlayer: game.PlayerYou,
						CostModifier: game.CostModifier{
							Kind:             game.CostModifierSpell,
							GenericReduction: 2,
							SourceZones: []zone.Type{
								zone.Graveyard, zone.Exile, zone.Library, zone.Command,
							},
						},
					}},
				},
			},
		},
	}
}

var rangerClassWolfToken = &game.CardDef{CardFace: game.CardFace{
	Name:      "Wolf",
	Colors:    []color.Color{color.Green},
	Types:     []types.Card{types.Creature},
	Subtypes:  []types.Sub{types.Wolf},
	Power:     opt.Val(game.PT{Value: 2}),
	Toughness: opt.Val(game.PT{Value: 2}),
}}

func newRangerClass() *game.CardDef {
	const text = "(Gain the next level as a sorcery to add its ability.)\n" +
		"When this Class enters, create a 2/2 green Wolf creature token.\n" +
		"{1}{G}: Level 2\n" +
		"Whenever you attack, put a +1/+1 counter on target attacking creature.\n" +
		"{3}{G}: Level 3\n" +
		"You may look at the top card of your library any time.\n" +
		"You may cast creature spells from the top of your library."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Ranger Class",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Colors:     []color.Color{color.Green},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Class},
			OracleText: text,
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{1}{G}: Level 2", cost.Mana{cost.O(1), cost.G}, 2),
				classLevelAbility("{3}{G}: Level 3", cost.Mana{cost.O(3), cost.G}, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "When this Class enters, create a 2/2 green Wolf creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.CreateToken{
							Amount: game.Fixed(1),
							Source: game.TokenDef(rangerClassWolfToken),
						},
					}}}.Ability(),
				},
				{
					Text: "Whenever you attack, put a +1/+1 counter on target attacking creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Controller: game.TriggerControllerYou,
							OneOrMore:  true,
						},
						InterveningCondition: opt.Val(game.Condition{
							SourceClassLevelAtLeast: 2,
						}),
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target attacking creature",
							Allow:      game.TargetAllowPermanent,
							Selection: opt.Val(game.Selection{
								RequiredTypes: []types.Card{types.Creature},
								CombatState:   game.CombatStateAttacking,
							}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.AddCounter{
								Amount:      game.Fixed(1),
								Object:      game.TargetPermanentReference(0),
								CounterKind: counter.PlusOnePlusOne,
							},
						}},
					}.Ability(),
				},
			},
			StaticAbilities: []game.StaticAbility{
				{
					Text: "You may look at the top card of your library any time.",
					Condition: opt.Val(game.Condition{
						SourceClassLevelAtLeast: 3,
					}),
					RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectLookAtTopCardAnyTime,
						AffectedPlayer: game.PlayerYou,
					}},
				},
				{
					Text: "You may cast creature spells from the top of your library.",
					Condition: opt.Val(game.Condition{
						SourceClassLevelAtLeast: 3,
					}),
					RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectCastSpellsFromZone,
						Duration:       game.DurationPermanent,
						AffectedPlayer: game.PlayerYou,
						SpellTypes:     []types.Card{types.Creature},
						CastFromZone:   zone.Library,
						TopCardOnly:    true,
					}},
				},
			},
		},
	}
}

func newWarlockClass() *game.CardDef {
	const text = "(Gain the next level as a sorcery to add its ability.)\n" +
		"At the beginning of your end step, if a creature died this turn, each opponent loses 1 life.\n" +
		"{1}{B}: Level 2\n" +
		"When this Class becomes level 2, look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.\n" +
		"{6}{B}: Level 3\n" +
		"At the beginning of your end step, each opponent loses life equal to the life they lost this turn. (Damage causes loss of life.)"
	creatureDiedThisTurn := game.Condition{
		EventHistory: opt.Val(game.EventHistoryCondition{
			Pattern: game.TriggerPattern{
				Event:            game.EventPermanentDied,
				SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
			},
			Window: game.EventHistoryCurrentTurn,
		}),
	}
	groupMember := game.GroupOfferMemberReference()
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Warlock Class",
			ManaCost:   opt.Val(cost.Mana{cost.B}),
			Colors:     []color.Color{color.Black},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Class},
			OracleText: text,
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{1}{B}: Level 2", cost.Mana{cost.O(1), cost.B}, 2),
				classLevelAbility("{6}{B}: Level 3", cost.Mana{cost.O(6), cost.B}, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				{
					Text: "At the beginning of your end step, if a creature died this turn, each opponent loses 1 life.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningCondition: opt.Val(creatureDiedThisTurn),
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.LoseLife{
							Amount:      game.Fixed(1),
							PlayerGroup: game.OpponentsReference(),
						},
					}}}.Ability(),
				},
				{
					Text: "When this Class becomes level 2, look at the top three cards of your library. Put one of them into your hand and the rest into your graveyard.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventClassLevelGained,
							Source:           game.TriggerSourceSelf,
							ClassBecameLevel: 2,
						},
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.Dig{
							Player:    game.ControllerReference(),
							Look:      game.Fixed(3),
							Take:      game.Fixed(1),
							Remainder: game.DigRemainderGraveyard,
						},
					}}}.Ability(),
				},
				{
					Text: "At the beginning of your end step, each opponent loses life equal to the life they lost this turn.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
						InterveningCondition: opt.Val(game.Condition{
							SourceClassLevelAtLeast: 3,
						}),
					},
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.LoseLife{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:   game.DynamicAmountLifeLostThisTurn,
								Player: &groupMember,
							}),
							Player: game.GroupOfferMemberReference(),
						},
						ForEachPlayerGroup: opt.Val(game.OpponentsReference()),
					}}}.Ability(),
				},
			},
		},
	}
}

const artistsTalentDiscardResult game.ResultKey = "artists-talent-discard"

func newArtistsTalent() *game.CardDef {
	const text = "(Gain the next level as a sorcery to add its ability.)\n" +
		"Whenever you cast a noncreature spell, you may discard a card. If you do, draw a card.\n" +
		"{2}{R}: Level 2\n" +
		"Noncreature spells you cast cost {1} less to cast.\n" +
		"{2}{R}: Level 3\n" +
		"If a source you control would deal noncombat damage to an opponent or a permanent an opponent controls, it deals that much damage plus 2 instead."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Artist's Talent",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Enchantment},
			Subtypes:   []types.Sub{types.Class},
			OracleText: text,
			ActivatedAbilities: []game.ActivatedAbility{
				classLevelAbility("{2}{R}: Level 2", cost.Mana{cost.O(2), cost.R}, 2),
				classLevelAbility("{2}{R}: Level 3", cost.Mana{cost.O(2), cost.R}, 3),
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: "Whenever you cast a noncreature spell, you may discard a card. If you do, draw a card.",
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:            game.EventSpellCast,
						Player:           game.TriggerPlayerYou,
						ExcludeCardTypes: []types.Card{types.Creature},
					},
				},
				Content: game.Mode{Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Optional:      true,
						PublishResult: artistsTalentDiscardResult,
					},
					{
						Primitive: game.Draw{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:         artistsTalentDiscardResult,
							AmountRange: opt.Val(game.IntRange{Min: 1, Max: 1}),
						}),
					},
				}}.Ability(),
			}},
			StaticAbilities: []game.StaticAbility{
				{
					Text: "Noncreature spells you cast cost {1} less to cast.",
					Condition: opt.Val(game.Condition{
						SourceClassLevelAtLeast: 2,
					}),
					RuleEffects: []game.RuleEffect{{
						Kind:           game.RuleEffectCostModifier,
						Duration:       game.DurationPermanent,
						AffectedPlayer: game.PlayerYou,
						CostModifier: game.CostModifier{
							Kind:             game.CostModifierSpell,
							GenericReduction: 1,
							CardSelection:    game.Selection{ExcludedTypes: []types.Card{types.Creature}},
						},
					}},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.ClassLevelGatedReplacement(
					game.DamageReplacementFiltered(
						"If a source you control would deal noncombat damage to an opponent or a permanent an opponent controls, it deals that much damage plus 2 instead.",
						&game.DamageReplacementSpec{
							Addend:            2,
							RecipientOpponent: true,
							NoncombatOnly:     true,
							Controller:        game.TriggerControllerYou,
						},
					),
					3,
				),
			},
		},
	}
}
