package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// StormchaserSTalent is the card definition for Stormchaser's Talent.
//
// Type: Enchantment — Class
// Cost: {U}
//
// Oracle text:
//
//	(Gain the next level as a sorcery to add its ability.)
//	When this Class enters, create a 1/1 blue and red Otter creature token with prowess.
//	{3}{U}: Level 2
//	When this Class becomes level 2, return target instant or sorcery card from your graveyard to your hand.
//	{5}{U}: Level 3
//	Whenever you cast an instant or sorcery spell, create a 1/1 blue and red Otter creature token with prowess.
var StormchaserSTalent = newStormchaserSTalent

func newStormchaserSTalent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Stormchaser's Talent",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Class},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{U}: Level 2",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.U,
					}),
					Timing: game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						SourceClassLevelLessThan: 2,
						SourceCounterKind:        counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SetClassLevel{
											Object: game.SourcePermanentReference(),
											Amount: game.Fixed(2),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ActivatedAbility{
					Text: "{5}{U}: Level 3",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.U,
					}),
					Timing: game.SorceryOnly,
					ActivationCondition: opt.Val(game.Condition{
						SourceClassLevelAtLeast:  2,
						SourceClassLevelLessThan: 3,
						SourceCounterKind:        counter.PlusOnePlusOne,
					}),
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.SetClassLevel{
											Object: game.SourcePermanentReference(),
											Amount: game.Fixed(3),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this Class enters, create a 1/1 blue and red Otter creature token with prowess.",
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
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(stormchaserSTalentToken),
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
					Text: "When this Class becomes level 2, return target instant or sorcery card from your graveyard to your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventClassLevelGained,
							Source:           game.TriggerSourceSelf,
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
							ClassBecameLevel: 2,
						},
						InterveningCondition: opt.Val(game.Condition{
							SourceClassLevelAtLeast: 2,
							SourceCounterKind:       counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target instant or sorcery card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}, Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:    zone.Graveyard,
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
					Text: "Whenever you cast an instant or sorcery spell, create a 1/1 blue and red Otter creature token with prowess.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{RequiredTypesAny: []types.Card{types.Instant, types.Sorcery}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningCondition: opt.Val(game.Condition{
							SourceClassLevelAtLeast: 3,
							SourceCounterKind:       counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(stormchaserSTalentToken),
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
			(Gain the next level as a sorcery to add its ability.)
			When this Class enters, create a 1/1 blue and red Otter creature token with prowess.
			{3}{U}: Level 2
			When this Class becomes level 2, return target instant or sorcery card from your graveyard to your hand.
			{5}{U}: Level 3
			Whenever you cast an instant or sorcery spell, create a 1/1 blue and red Otter creature token with prowess.
		`,
		},
	}
}

var stormchaserSTalentToken = newStormchaserSTalentToken()

func newStormchaserSTalentToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Otter",
			Colors:    []color.Color{color.Blue, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Otter},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
			},
		},
	}
}
