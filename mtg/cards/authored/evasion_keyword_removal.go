package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	hammerheimOracleID         = "c7476beb-7923-4994-8476-bc69187ecb72"
	crashLandingOracleID       = "5d36e5ee-1e23-48fe-ad13-081cc153342b"
	kinscaerHarpoonistOracleID = "8ee98d6a-6a7e-411d-818c-816d84cfb7cd"
	burningPalmEfreetOracleID  = "55cbc0e8-eeec-433a-b70c-3501238b658b"
	windShearOracleID          = "fa683059-b9de-4063-ae71-597ed4e2df30"
	leeringGargoyleOracleID    = "fc0907eb-a3fc-40ef-950d-b05aee9d5da7"

	hammerheimLandwalkText = "{T}: Target creature loses all landwalk abilities until end of turn."
	hammerheimText         = "{T}: Add {R}.\n" + hammerheimLandwalkText
	crashLandingText       = "Target creature with flying loses flying until end of turn. " +
		"Crash Landing deals damage to that creature equal to the number of Forests you control."
	kinscaerHarpoonistText     = "Flying\nWhenever this creature attacks, you may have target creature lose flying until end of turn."
	burningPalmEfreetText      = "{1}{R}{R}: This creature deals 2 damage to target creature with flying and that creature loses flying until end of turn."
	windShearText              = "Attacking creatures with flying get -2/-2 and lose flying until end of turn."
	leeringGargoyleAbilityText = "{T}: This creature gets -2/+2 and loses flying until end of turn."
	leeringGargoyleText        = "Flying\n" + leeringGargoyleAbilityText
)

var (
	Hammerheim = cardset.Entry{
		OracleID: hammerheimOracleID,
		Name:     "Hammerheim",
		New:      newHammerheim,
		Verification: "mtg/rules/evasion_keyword_removal_test.go: " +
			"TestHammerheimRemovesAllLandwalkAbilities, TestHammerheimCanTargetCreatureWithoutLandwalk, " +
			"TestHammerheimAddsRedMana",
	}

	CrashLanding = cardset.Entry{
		OracleID: crashLandingOracleID,
		Name:     "Crash Landing",
		New:      newCrashLanding,
		Verification: "mtg/rules/evasion_keyword_removal_test.go: " +
			"TestCrashLandingRemovesFlyingThenDealsDamageFromForests",
	}

	KinscaerHarpoonist = cardset.Entry{
		OracleID: kinscaerHarpoonistOracleID,
		Name:     "Kinscaer Harpoonist",
		New:      newKinscaerHarpoonist,
		Verification: "mtg/rules/evasion_keyword_removal_test.go: " +
			"TestKinscaerHarpoonistAttackMayRemoveFlying, TestKinscaerHarpoonistMayTargetCreatureWithoutFlying, " +
			"TestKinscaerHarpoonistMayDeclineFlyingRemoval",
	}

	BurningPalmEfreet = cardset.Entry{
		OracleID: burningPalmEfreetOracleID,
		Name:     "Burning Palm Efreet",
		New:      newBurningPalmEfreet,
		Verification: "mtg/rules/evasion_keyword_removal_test.go: " +
			"TestBurningPalmEfreetDealsDamageAndRemovesFlying",
	}

	WindShear = cardset.Entry{
		OracleID: windShearOracleID,
		Name:     "Wind Shear",
		New:      newWindShear,
		Verification: "mtg/rules/evasion_keyword_removal_test.go: " +
			"TestWindShearAffectsOnlyAttackingCreaturesWithFlying",
	}

	LeeringGargoyle = cardset.Entry{
		OracleID: leeringGargoyleOracleID,
		Name:     "Leering Gargoyle",
		New:      newLeeringGargoyle,
		Verification: "mtg/rules/evasion_keyword_removal_test.go: " +
			"TestLeeringGargoyleTapGetsMinusTwoPlusTwoAndLosesFlying",
	}
)

func newHammerheim() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Hammerheim",
			OracleText: hammerheimText,
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text: hammerheimLandwalkText,
				AdditionalCosts: []cost.Additional{{
					Kind: cost.AdditionalTap,
				}},
				ZoneOfFunction: zone.Battlefield,
				Content: game.AbilityContent{
					Modes: []game.Mode{{
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						}},
						Sequence: []game.Instruction{{
							Primitive: game.ApplyContinuous{
								Object: opt.Val(game.TargetPermanentReference(0)),
								ContinuousEffects: []game.ContinuousEffect{{
									Duration:       game.DurationPermanent,
									Layer:          game.LayerAbility,
									RemoveKeywords: []game.Keyword{game.Landwalk},
								}},
								Duration: game.DurationUntilEndOfTurn,
							},
						}},
					}},
					MinModes: 1,
					MaxModes: 1,
				},
			}},
		},
	}
}

func newCrashLanding() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Crash Landing",
			OracleText: crashLandingText,
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.G}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Green},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{{
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature with flying",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypesAny: []types.Card{types.Creature},
							Keyword:          game.Flying,
						}),
					}},
					Sequence: []game.Instruction{
						{
							Primitive: game.ApplyContinuous{
								Object: opt.Val(game.TargetPermanentReference(0)),
								ContinuousEffects: []game.ContinuousEffect{{
									Duration:       game.DurationPermanent,
									Layer:          game.LayerAbility,
									RemoveKeywords: []game.Keyword{game.Flying},
								}},
								Duration: game.DurationUntilEndOfTurn,
							},
						},
						{
							Primitive: game.Damage{
								Amount: game.Dynamic(game.DynamicAmount{
									Kind: game.DynamicAmountCountSelector,
									Group: game.BattlefieldGroup(game.Selection{
										RequiredTypes: []types.Card{types.Land},
										SubtypesAny:   []types.Sub{types.Forest},
										Controller:    game.ControllerYou,
									}),
								}),
								Recipient: game.AnyTargetDamageRecipient(0),
							},
						},
					},
				}},
				MinModes: 1,
				MaxModes: 1,
			}),
		},
	}
}

func newKinscaerHarpoonist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Kinscaer Harpoonist",
			OracleText: kinscaerHarpoonistText,
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.U}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Kithkin, types.Soldier},
			Colors:     []color.Color{color.Blue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: "Whenever this creature attacks, you may have target creature lose flying until end of turn.",
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventAttackerDeclared,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.AbilityContent{
					Modes: []game.Mode{{
						Targets: []game.TargetSpec{{
							MinTargets: 1,
							MaxTargets: 1,
							Constraint: "target creature",
							Allow:      game.TargetAllowPermanent,
							Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
						}},
						Sequence: []game.Instruction{{
							Optional: true,
							Primitive: game.ApplyContinuous{
								Object: opt.Val(game.TargetPermanentReference(0)),
								ContinuousEffects: []game.ContinuousEffect{{
									Duration:       game.DurationPermanent,
									Layer:          game.LayerAbility,
									RemoveKeywords: []game.Keyword{game.Flying},
								}},
								Duration: game.DurationUntilEndOfTurn,
							},
						}},
					}},
					MinModes: 1,
					MaxModes: 1,
				},
			}},
		},
	}
}

func newBurningPalmEfreet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Burning Palm Efreet",
			OracleText: burningPalmEfreetText,
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.R, cost.R}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Efreet},
			Colors:     []color.Color{color.Red},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:           burningPalmEfreetText,
				ManaCost:       opt.Val(cost.Mana{cost.O(1), cost.R, cost.R}),
				ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature with flying",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypesAny: []types.Card{types.Creature},
							Keyword:          game.Flying,
						}),
					}},
					Sequence: []game.Instruction{
						{Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.ObjectDamageRecipient(game.TargetPermanentReference(0)),
						}},
						{Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{{
								Layer:          game.LayerAbility,
								RemoveKeywords: []game.Keyword{game.Flying},
							}},
							Duration: game.DurationUntilEndOfTurn,
						}},
					},
				}.Ability(),
			}},
		},
	}
}

func newWindShear() *game.CardDef {
	attackingFlyers := game.BattlefieldGroup(game.Selection{
		RequiredTypes: []types.Card{types.Creature},
		CombatState:   game.CombatStateAttacking,
		Keyword:       game.Flying,
	})
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Wind Shear",
			OracleText: windShearText,
			ManaCost:   opt.Val(cost.Mana{cost.O(2), cost.G}),
			Types:      []types.Card{types.Instant},
			Colors:     []color.Color{color.Green},
			SpellAbility: opt.Val(game.Mode{Sequence: []game.Instruction{{
				Primitive: game.ApplyContinuous{
					ContinuousEffects: []game.ContinuousEffect{
						{
							Layer:          game.LayerPowerToughnessModify,
							Group:          attackingFlyers,
							PowerDelta:     -2,
							ToughnessDelta: -2,
						},
						{
							Layer:          game.LayerAbility,
							Group:          attackingFlyers,
							RemoveKeywords: []game.Keyword{game.Flying},
						},
					},
					Duration: game.DurationUntilEndOfTurn,
				},
			}}}.Ability()),
		},
	}
}

func newLeeringGargoyle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:       "Leering Gargoyle",
			OracleText: leeringGargoyleText,
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.W, cost.U}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Gargoyle},
			Colors:     []color.Color{color.White, color.Blue},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text: leeringGargoyleAbilityText,
				AdditionalCosts: []cost.Additional{{
					Kind: cost.AdditionalTap,
				}},
				ZoneOfFunction: zone.Battlefield,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.ApplyContinuous{
						Object: opt.Val(game.SourcePermanentReference()),
						ContinuousEffects: []game.ContinuousEffect{
							{
								Layer:          game.LayerPowerToughnessModify,
								PowerDelta:     -2,
								ToughnessDelta: 2,
							},
							{
								Layer:          game.LayerAbility,
								RemoveKeywords: []game.Keyword{game.Flying},
							},
						},
						Duration: game.DurationUntilEndOfTurn,
					},
				}}}.Ability(),
			}},
		},
	}
}
