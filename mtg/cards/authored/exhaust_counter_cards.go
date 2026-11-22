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

var GreenbeltGuardian = cardset.Entry{
	OracleID:     "2d8aa053-289d-40d9-baa7-9bd1c5b8e957",
	Name:         "Greenbelt Guardian",
	New:          newGreenbeltGuardian,
	Verification: "mtg/rules/exhaust_counter_cards_test.go: TestGreenbeltGuardianScenario",
}

func newGreenbeltGuardian() *game.CardDef {
	const pumpText = "{G}: Target creature gains trample until end of turn."
	const exhaustText = "Exhaust — {3}{G}: Put three +1/+1 counters on this creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Greenbelt Guardian",
			OracleText: pumpText + "\n" + exhaustText + " (Activate each exhaust ability only once.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.G}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Elf, types.Ranger},
			Colors:     []color.Color{color.Green},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				{
					Text:           pumpText,
					ManaCost:       opt.Val(cost.Mana{cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{exhaustCreatureTargetSpec()},
						Sequence: []game.Instruction{{
							Primitive: game.ApplyContinuous{
								Object: opt.Val(game.TargetPermanentReference(0)),
								ContinuousEffects: []game.ContinuousEffect{{
									Layer:       game.LayerAbility,
									AddKeywords: []game.Keyword{game.Trample},
								}},
								Duration: game.DurationUntilEndOfTurn,
							},
						}},
					}.Ability(),
				},
				{
					Text:             exhaustText,
					ManaCost:         opt.Val(cost.Mana{cost.O(3), cost.G}),
					ZoneOfFunction:   zone.Battlefield,
					ActivateOnlyOnce: true,
					Content: game.Mode{Sequence: []game.Instruction{{
						Primitive: game.AddCounter{
							Object:      game.SourcePermanentReference(),
							Amount:      game.Fixed(3),
							CounterKind: counter.PlusOnePlusOne,
						},
					}}}.Ability(),
				},
			},
		},
	}
}

var SkystreakEngineer = cardset.Entry{
	OracleID:     "2f920bf2-6a49-47c2-be3c-9787c377673b",
	Name:         "Skystreak Engineer",
	New:          newSkystreakEngineer,
	Verification: "mtg/rules/exhaust_counter_cards_test.go: TestSkystreakEngineerScenario",
}

func newSkystreakEngineer() *game.CardDef {
	const exhaustText = "Exhaust — {4}{U}: Put two +1/+1 counters on this creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Skystreak Engineer",
			OracleText: "Flying\n" + exhaustText + " (Activate each exhaust ability only once.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Pilot},
			Colors:     []color.Color{color.Blue},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:             exhaustText,
				ManaCost:         opt.Val(cost.Mana{cost.O(4), cost.U}),
				ZoneOfFunction:   zone.Battlefield,
				ActivateOnlyOnce: true,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddCounter{
						Object:      game.SourcePermanentReference(),
						Amount:      game.Fixed(2),
						CounterKind: counter.PlusOnePlusOne,
					},
				}}}.Ability(),
			}},
		},
	}
}

var ProwcatcherSpecialist = cardset.Entry{
	OracleID:     "594132ac-32a7-41d5-b7f0-3692bbed7f6f",
	Name:         "Prowcatcher Specialist",
	New:          newProwcatcherSpecialist,
	Verification: "mtg/rules/exhaust_counter_cards_test.go: TestProwcatcherSpecialistScenario",
}

func newProwcatcherSpecialist() *game.CardDef {
	const exhaustText = "Exhaust — {3}{R}: Put two +1/+1 counters on this creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Prowcatcher Specialist",
			OracleText: "Haste\n" + exhaustText + " (Activate each exhaust ability only once.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Goblin, types.Warrior},
			Colors:     []color.Color{color.Red},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:             exhaustText,
				ManaCost:         opt.Val(cost.Mana{cost.O(3), cost.R}),
				ZoneOfFunction:   zone.Battlefield,
				ActivateOnlyOnce: true,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddCounter{
						Object:      game.SourcePermanentReference(),
						Amount:      game.Fixed(2),
						CounterKind: counter.PlusOnePlusOne,
					},
				}}}.Ability(),
			}},
		},
	}
}

var HazardOfTheDunes = cardset.Entry{
	OracleID:     "799664fc-279c-40c1-8e40-44a68c38801b",
	Name:         "Hazard of the Dunes",
	New:          newHazardOfTheDunes,
	Verification: "mtg/rules/exhaust_counter_cards_test.go: TestHazardOfTheDunesScenario",
}

func newHazardOfTheDunes() *game.CardDef {
	const exhaustText = "Exhaust — {6}{G}: Put three +1/+1 counters on this creature."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Hazard of the Dunes",
			OracleText: "Reach, trample\n" + exhaustText + " (Activate each exhaust ability only once.)",
			ManaCost:   opt.Val(cost.Mana{cost.O(3), cost.G}),
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Wurm},
			Colors:     []color.Color{color.Green},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.TrampleStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{{
				Text:             exhaustText,
				ManaCost:         opt.Val(cost.Mana{cost.O(6), cost.G}),
				ZoneOfFunction:   zone.Battlefield,
				ActivateOnlyOnce: true,
				Content: game.Mode{Sequence: []game.Instruction{{
					Primitive: game.AddCounter{
						Object:      game.SourcePermanentReference(),
						Amount:      game.Fixed(3),
						CounterKind: counter.PlusOnePlusOne,
					},
				}}}.Ability(),
			}},
		},
	}
}

func exhaustCreatureTargetSpec() game.TargetSpec {
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: "target creature",
		Allow:      game.TargetAllowPermanent,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
	}
}
