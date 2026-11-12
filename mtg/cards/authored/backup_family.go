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
	saibaCryptomancerOracleID     = "15f89b06-5025-4c9f-bf51-e25bc5c2dd6b"
	chompingKavuOracleID          = "20a362c3-d278-437f-881c-e7d9fe78fc2d"
	consumingAetherbornOracleID   = "4d472308-5bc7-4a3c-81b5-0a757f00d9e2"
	gloomfangMaulerOracleID       = "5437f560-7f61-4326-be8e-15d74bf4e6ec"
	cragsmasherYetiOracleID       = "57120486-0316-4591-abca-438ab8283d3f"
	serpentBladeAssailantOracleID = "66a131ae-dbb2-4e7b-b6ce-7695f47c8f87"
	sigiledSentinelOracleID       = "79c90a68-1c31-4e16-be53-1304132da80f"
	deathGreetersChampionOracleID = "81abb451-5ee0-4f90-bacc-a8ff69726a27"
	scornBladeBerserkerOracleID   = "93107139-b743-4d62-b0ac-9997202e0562"
	voldarenThrillseekerOracleID  = "b3cc2ed9-1da3-4ad2-a9a7-de7665d800ff"
	enduringBondwardenOracleID    = "c17d4818-5825-4bff-8fc5-5fd740cc0c70"
	hangarScroungerOracleID       = "d1b74fb9-8ef3-40be-bc79-45ffb4c597eb"
	redcapHeelslasherOracleID     = "d2134b68-d73f-4090-aaec-b2ee5c9f9f04"
	goldenScaleAeronautOracleID   = "d388080d-5ad4-4e27-b361-340f410423a7"
	fearlessSkaldOracleID         = "f055a941-8b83-464e-a57f-8f188dc69aaa"
)

const (
	saibaBackupText        = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	kavuBackupText         = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	aetherbornBackupText   = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	gloomfangBackupText    = "Backup 2 (When this creature enters, put two +1/+1 counters on target creature. If that's another creature, it gains the following ability until end of turn.)"
	cragsmasherBackupText  = "Backup 2 (When this creature enters, put two +1/+1 counters on target creature. If that's another creature, it gains the following ability until end of turn.)"
	serpentBladeBackupText = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	sigiledBackupText      = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	deathGreeterBackupText = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	scornBladeBackupText   = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	voldarenBackupText     = "Backup 2 (When this creature enters, put two +1/+1 counters on target creature. If that's another creature, it gains the following ability until end of turn.)"
	enduringBackupText     = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	hangarBackupText       = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	redcapBackupText       = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	goldenBackupText       = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"
	fearlessBackupText     = "Backup 1 (When this creature enters, put a +1/+1 counter on target creature. If that's another creature, it gains the following ability until end of turn.)"

	kavuBlockText         = "This creature can't be blocked by creatures with power 2 or less."
	scornBladeAbilityText = "{1}, Sacrifice this creature: Draw a card."
	voldarenAbilityText   = "{1}, Sacrifice this creature: It deals damage equal to its power to any target."
	enduringDeathText     = "When this creature dies, put its counters on target creature you control."
	hangarTapText         = "Whenever this creature becomes tapped, you may discard a card. If you do, draw a card."
)

var (
	SaibaCryptomancer = cardset.Entry{
		OracleID: saibaCryptomancerOracleID, Name: "Saiba Cryptomancer", New: newSaibaCryptomancer,
		Verification: "mtg/rules/backup_family_test.go: TestSaibaCryptomancerFlashesInAndGrantsHexproof",
	}
	ChompingKavu = cardset.Entry{
		OracleID: chompingKavuOracleID, Name: "Chomping Kavu", New: newChompingKavu,
		Verification: "mtg/rules/backup_family_test.go: TestChompingKavuBackupRestrictsBlockersByPower",
	}
	ConsumingAetherborn = cardset.Entry{
		OracleID: consumingAetherbornOracleID, Name: "Consuming Aetherborn", New: newConsumingAetherborn,
		Verification: "mtg/rules/backup_family_test.go: TestConsumingAetherbornBackupLifelinkGainsLifeFromCombat",
	}
	GloomfangMauler = cardset.Entry{
		OracleID: gloomfangMaulerOracleID, Name: "Gloomfang Mauler", New: newGloomfangMauler,
		Verification: "mtg/rules/backup_family_test.go: TestGloomfangMaulerMenaceAndSwampcycling",
	}
	CragsmasherYeti = cardset.Entry{
		OracleID: cragsmasherYetiOracleID, Name: "Cragsmasher Yeti", New: newCragsmasherYeti,
		Verification: "mtg/rules/backup_family_test.go: TestCragsmasherYetiTramplesAndMountaincycles",
	}
	SerpentBladeAssailant = cardset.Entry{
		OracleID: serpentBladeAssailantOracleID, Name: "Serpent-Blade Assailant", New: newSerpentBladeAssailant,
		Verification: "mtg/rules/backup_family_test.go: TestSerpentBladeAssailantBackupDeathtouchMarksLethalDamage",
	}
	SigiledSentinel = cardset.Entry{
		OracleID: sigiledSentinelOracleID, Name: "Sigiled Sentinel", New: newSigiledSentinel,
		Verification: "mtg/rules/backup_family_test.go: TestSigiledSentinelBackupVigilanceLeavesAttackerUntapped",
	}
	DeathGreetersChampion = cardset.Entry{
		OracleID: deathGreetersChampionOracleID, Name: "Death-Greeter's Champion", New: newDeathGreetersChampion,
		Verification: "mtg/rules/backup_family_test.go: TestDeathGreetersChampionDashesAndGrantsDoubleStrike",
	}
	ScornBladeBerserker = cardset.Entry{
		OracleID: scornBladeBerserkerOracleID, Name: "Scorn-Blade Berserker", New: newScornBladeBerserker,
		Verification: "mtg/rules/backup_family_test.go: TestScornBladeBerserkerBackupGrantsSacrificeDraw",
	}
	VoldarenThrillseeker = cardset.Entry{
		OracleID: voldarenThrillseekerOracleID, Name: "Voldaren Thrillseeker", New: newVoldarenThrillseeker,
		Verification: "mtg/rules/backup_family_test.go: TestVoldarenThrillseekerBackupAbilityUsesLastKnownPower",
	}
	EnduringBondwarden = cardset.Entry{
		OracleID: enduringBondwardenOracleID, Name: "Enduring Bondwarden", New: newEnduringBondwarden,
		Verification: "mtg/rules/backup_family_test.go: TestEnduringBondwardenBackupDeathTriggerCopiesEveryCounterKind",
	}
	HangarScrounger = cardset.Entry{
		OracleID: hangarScroungerOracleID, Name: "Hangar Scrounger", New: newHangarScrounger,
		Verification: "mtg/rules/backup_family_test.go: TestHangarScroungerBackupTapTriggerDiscardsAndDraws",
	}
	RedcapHeelslasher = cardset.Entry{
		OracleID: redcapHeelslasherOracleID, Name: "Redcap Heelslasher", New: newRedcapHeelslasher,
		Verification: "mtg/rules/backup_family_test.go: TestRedcapHeelslasherFirstStrikeKillsBlockerBeforeDamage",
	}
	GoldenScaleAeronaut = cardset.Entry{
		OracleID: goldenScaleAeronautOracleID, Name: "Golden-Scale Aeronaut", New: newGoldenScaleAeronaut,
		Verification: "mtg/rules/backup_family_test.go: TestGoldenScaleAeronautFlyingAllowsOnlyAerialBlocks",
	}
	FearlessSkald = cardset.Entry{
		OracleID: fearlessSkaldOracleID, Name: "Fearless Skald", New: newFearlessSkald,
		Verification: "mtg/rules/backup_family_test.go: TestFearlessSkaldBackupDoubleStrikeDealsDamageTwice",
	}
)

func backupCreature(name string, manaCost cost.Mana, colorValue color.Color, subtypes []types.Sub, power, toughness int, oracleText string, static []game.StaticAbility, activated []game.ActivatedAbility, triggered []game.TriggeredAbility, alternatives []cost.Alternative) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(colorValue),
		CardFace: game.CardFace{
			Name:               name,
			OracleText:         oracleText,
			ManaCost:           opt.Val(manaCost),
			Types:              []types.Card{types.Creature},
			Subtypes:           subtypes,
			Colors:             []color.Color{colorValue},
			Power:              opt.Val(game.PT{Value: power}),
			Toughness:          opt.Val(game.PT{Value: toughness}),
			StaticAbilities:    static,
			ActivatedAbilities: activated,
			TriggeredAbilities: triggered,
			AlternativeCosts:   alternatives,
		},
	}
}

func backupGrantedAbility(body game.StaticAbility) game.Ability {
	copy := body
	return &copy
}

func backupTriggeredAbility(amount int, text string, granted game.Ability) game.TriggeredAbility {
	anotherCreature := opt.Val(game.EffectCondition{
		Text: "the target is another creature",
		Condition: opt.Val(game.Condition{
			Object:        opt.Val(game.TargetPermanentReference(0)),
			ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true}),
		}),
	})
	return game.TriggeredAbility{
		Text: text,
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  game.EventPermanentEnteredBattlefield,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1,
				MaxTargets: 1,
				Constraint: "target creature",
				Allow:      game.TargetAllowPermanent,
				Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}},
			Sequence: []game.Instruction{
				{Primitive: game.AddCounter{
					Object:      game.TargetPermanentReference(0),
					Amount:      game.Fixed(amount),
					CounterKind: counter.PlusOnePlusOne,
				}},
				{
					Condition: anotherCreature,
					Primitive: game.ApplyContinuous{
						Object: opt.Val(game.TargetPermanentReference(0)),
						ContinuousEffects: []game.ContinuousEffect{{
							Layer:        game.LayerAbility,
							AddAbilities: []game.Ability{granted},
						}},
						Duration: game.DurationUntilEndOfTurn,
					},
				},
			},
		}.Ability(),
	}
}

func newSaibaCryptomancer() *game.CardDef {
	return backupCreature("Saiba Cryptomancer", cost.Mana{cost.O(1), cost.U}, color.Blue,
		[]types.Sub{types.Moonfolk, types.Ninja}, 0, 1, "Flash\n"+saibaBackupText+"\nHexproof",
		[]game.StaticAbility{game.FlashStaticBody, game.HexproofStaticBody}, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, saibaBackupText, backupGrantedAbility(game.HexproofStaticBody))}, nil)
}

func chompingKavuAbility() game.StaticAbility {
	return game.StaticAbility{
		Text: kavuBlockText,
		RuleEffects: []game.RuleEffect{{
			Kind:           game.RuleEffectCantBeBlockedByCreaturesWith,
			AffectedSource: true,
			BlockerRestriction: game.BlockerRestriction{
				Kind:  game.BlockerRestrictionPowerLessOrEqual,
				Power: 2,
			},
		}},
	}
}

func newChompingKavu() *game.CardDef {
	ability := chompingKavuAbility()
	return backupCreature("Chomping Kavu", cost.Mana{cost.O(3), cost.G}, color.Green,
		[]types.Sub{types.Kavu}, 3, 3, kavuBackupText+"\n"+kavuBlockText,
		[]game.StaticAbility{ability}, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, kavuBackupText, backupGrantedAbility(ability))}, nil)
}

func newConsumingAetherborn() *game.CardDef {
	return backupCreature("Consuming Aetherborn", cost.Mana{cost.O(3), cost.B}, color.Black,
		[]types.Sub{types.Aetherborn, types.Vampire}, 2, 2, aetherbornBackupText+"\nLifelink",
		[]game.StaticAbility{game.LifelinkStaticBody}, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, aetherbornBackupText, backupGrantedAbility(game.LifelinkStaticBody))}, nil)
}

func landcyclingAbility(subtype types.Sub) game.ActivatedAbility {
	return game.LandcyclingActivatedAbility(cost.Mana{cost.O(2)}, game.SearchSpec{
		Filter: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{subtype}},
	})
}

func newGloomfangMauler() *game.CardDef {
	return backupCreature("Gloomfang Mauler", cost.Mana{cost.O(5), cost.B, cost.B}, color.Black,
		[]types.Sub{types.Nightmare}, 5, 5,
		"Swampcycling {2} ({2}, Discard this card: Search your library for a Swamp card, reveal it, put it into your hand, then shuffle.)\n"+gloomfangBackupText+"\nMenace",
		[]game.StaticAbility{game.MenaceStaticBody}, []game.ActivatedAbility{landcyclingAbility(types.Swamp)},
		[]game.TriggeredAbility{backupTriggeredAbility(2, gloomfangBackupText, backupGrantedAbility(game.MenaceStaticBody))}, nil)
}

func newCragsmasherYeti() *game.CardDef {
	return backupCreature("Cragsmasher Yeti", cost.Mana{cost.O(4), cost.R, cost.R}, color.Red,
		[]types.Sub{types.Yeti}, 4, 2,
		"Mountaincycling {2} ({2}, Discard this card: Search your library for a Mountain card, reveal it, put it into your hand, then shuffle.)\n"+cragsmasherBackupText+"\nTrample",
		[]game.StaticAbility{game.TrampleStaticBody}, []game.ActivatedAbility{landcyclingAbility(types.Mountain)},
		[]game.TriggeredAbility{backupTriggeredAbility(2, cragsmasherBackupText, backupGrantedAbility(game.TrampleStaticBody))}, nil)
}

func newSerpentBladeAssailant() *game.CardDef {
	return backupCreature("Serpent-Blade Assailant", cost.Mana{cost.O(2), cost.G}, color.Green,
		[]types.Sub{types.Elf, types.Warrior}, 2, 1, serpentBladeBackupText+"\nDeathtouch",
		[]game.StaticAbility{game.DeathtouchStaticBody}, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, serpentBladeBackupText, backupGrantedAbility(game.DeathtouchStaticBody))}, nil)
}

func newSigiledSentinel() *game.CardDef {
	return backupCreature("Sigiled Sentinel", cost.Mana{cost.O(2), cost.W}, color.White,
		[]types.Sub{types.Human, types.Knight}, 2, 2, sigiledBackupText+"\nVigilance",
		[]game.StaticAbility{game.VigilanceStaticBody}, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, sigiledBackupText, backupGrantedAbility(game.VigilanceStaticBody))}, nil)
}

func newDeathGreetersChampion() *game.CardDef {
	const dashText = "Dash {3}{R} (You may cast this spell for its dash cost. If you do, it gains haste, and it's returned from the battlefield to its owner's hand at the beginning of the next end step.)"
	return backupCreature("Death-Greeter's Champion", cost.Mana{cost.O(2), cost.R}, color.Red,
		[]types.Sub{types.Human, types.Warrior}, 2, 1, dashText+"\n"+deathGreeterBackupText+"\nDouble strike",
		[]game.StaticAbility{game.DoubleStrikeStaticBody}, nil,
		[]game.TriggeredAbility{game.DashTriggeredAbility(), backupTriggeredAbility(1, deathGreeterBackupText, backupGrantedAbility(game.DoubleStrikeStaticBody))},
		[]cost.Alternative{{
			Label:    "Dash",
			Mechanic: cost.AlternativeMechanicDash,
			ManaCost: opt.Val(cost.Mana{cost.O(3), cost.R}),
		}})
}

func scornBladeActivatedAbility() game.ActivatedAbility {
	return game.ActivatedAbility{
		Text:           scornBladeAbilityText,
		ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
		ZoneOfFunction: zone.Battlefield,
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalSacrificeSource,
			Text: "Sacrifice this creature",
		}},
		Content: game.Mode{Sequence: []game.Instruction{{
			Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
		}}}.Ability(),
	}
}

func newScornBladeBerserker() *game.CardDef {
	ability := scornBladeActivatedAbility()
	return backupCreature("Scorn-Blade Berserker", cost.Mana{cost.B}, color.Black,
		[]types.Sub{types.Human, types.Berserker}, 0, 1, scornBladeBackupText+"\n"+scornBladeAbilityText,
		nil, []game.ActivatedAbility{ability},
		[]game.TriggeredAbility{backupTriggeredAbility(1, scornBladeBackupText, &ability)}, nil)
}

func voldarenActivatedAbility() game.ActivatedAbility {
	return game.ActivatedAbility{
		Text:           voldarenAbilityText,
		ManaCost:       opt.Val(cost.Mana{cost.O(1)}),
		ZoneOfFunction: zone.Battlefield,
		AdditionalCosts: []cost.Additional{{
			Kind: cost.AdditionalSacrificeSource,
			Text: "Sacrifice this creature",
		}},
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1,
				Constraint: "any target",
				Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
			}},
			Sequence: []game.Instruction{{
				Primitive: game.Damage{
					Amount: game.Dynamic(game.DynamicAmount{
						Kind:   game.DynamicAmountObjectPower,
						Object: game.SourcePermanentReference(),
					}),
					Recipient: game.AnyTargetDamageRecipient(0),
				},
			}},
		}.Ability(),
	}
}

func newVoldarenThrillseeker() *game.CardDef {
	ability := voldarenActivatedAbility()
	return backupCreature("Voldaren Thrillseeker", cost.Mana{cost.O(2), cost.R}, color.Red,
		[]types.Sub{types.Vampire, types.Warrior}, 1, 1, voldarenBackupText+"\n"+voldarenAbilityText,
		nil, []game.ActivatedAbility{ability},
		[]game.TriggeredAbility{backupTriggeredAbility(2, voldarenBackupText, &ability)}, nil)
}

func enduringDeathAbility() game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: enduringDeathText,
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhen,
			Pattern: game.TriggerPattern{
				Event:  game.EventPermanentDied,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{
			Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1,
				Constraint: "target creature you control",
				Allow:      game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{
					RequiredTypes: []types.Card{types.Creature},
					Controller:    game.ControllerYou,
				}),
			}},
			Sequence: []game.Instruction{{
				Primitive: game.MoveCounters{
					Object:   game.TargetPermanentReference(0),
					Source:   game.CounterSourceSpec{Kind: game.CounterSourceEventPermanent},
					AllKinds: true,
				},
			}},
		}.Ability(),
	}
}

func newEnduringBondwarden() *game.CardDef {
	death := enduringDeathAbility()
	return backupCreature("Enduring Bondwarden", cost.Mana{cost.W}, color.White,
		[]types.Sub{types.Human, types.Scout}, 0, 1, enduringBackupText+"\n"+enduringDeathText,
		nil, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, enduringBackupText, &death), death}, nil)
}

func hangarTapAbility() game.TriggeredAbility {
	const discardResult game.ResultKey = "hangar-scrounger-discard"
	return game.TriggeredAbility{
		Text: hangarTapText,
		Trigger: game.TriggerCondition{
			Type: game.TriggerWhenever,
			Pattern: game.TriggerPattern{
				Event:  game.EventPermanentTapped,
				Source: game.TriggerSourceSelf,
			},
		},
		Content: game.Mode{Sequence: []game.Instruction{
			{
				Primitive:     game.Discard{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				Optional:      true,
				PublishResult: discardResult,
			},
			{
				Primitive: game.Draw{Player: game.ControllerReference(), Amount: game.Fixed(1)},
				ResultGate: opt.Val(game.InstructionResultGate{
					Key: discardResult, Accepted: game.TriTrue, Succeeded: game.TriTrue,
				}),
			},
		}}.Ability(),
	}
}

func newHangarScrounger() *game.CardDef {
	tapAbility := hangarTapAbility()
	return backupCreature("Hangar Scrounger", cost.Mana{cost.O(2), cost.R}, color.Red,
		[]types.Sub{types.Dwarf, types.Pilot}, 2, 1, hangarBackupText+"\n"+hangarTapText,
		nil, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, hangarBackupText, &tapAbility), tapAbility}, nil)
}

func newRedcapHeelslasher() *game.CardDef {
	return backupCreature("Redcap Heelslasher", cost.Mana{cost.O(3), cost.R}, color.Red,
		[]types.Sub{types.Goblin, types.Rogue}, 2, 3, redcapBackupText+"\nFirst strike",
		[]game.StaticAbility{game.FirstStrikeStaticBody}, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, redcapBackupText, backupGrantedAbility(game.FirstStrikeStaticBody))}, nil)
}

func newGoldenScaleAeronaut() *game.CardDef {
	return backupCreature("Golden-Scale Aeronaut", cost.Mana{cost.O(4), cost.W}, color.White,
		[]types.Sub{types.Dwarf, types.Pilot}, 2, 3, goldenBackupText+"\nFlying",
		[]game.StaticAbility{game.FlyingStaticBody}, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, goldenBackupText, backupGrantedAbility(game.FlyingStaticBody))}, nil)
}

func newFearlessSkald() *game.CardDef {
	return backupCreature("Fearless Skald", cost.Mana{cost.O(4), cost.R}, color.Red,
		[]types.Sub{types.Dwarf, types.Berserker}, 3, 2, fearlessBackupText+"\nDouble strike",
		[]game.StaticAbility{game.DoubleStrikeStaticBody}, nil,
		[]game.TriggeredAbility{backupTriggeredAbility(1, fearlessBackupText, backupGrantedAbility(game.DoubleStrikeStaticBody))}, nil)
}
