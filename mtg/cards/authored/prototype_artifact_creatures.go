package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var GoringWarplow = cardset.Entry{
	OracleID:     "02457c7f-eced-4bf3-9b61-309b4ceab2d5",
	Name:         "Goring Warplow",
	New:          newGoringWarplow,
	Verification: "mtg/rules/prototype_artifact_creatures_test.go: TestGoringWarplowScenario",
}

var CombatThresher = cardset.Entry{
	OracleID:     "0c460bdd-460b-4312-9ed2-32f3809a7d36",
	Name:         "Combat Thresher",
	New:          newCombatThresher,
	Verification: "mtg/rules/prototype_artifact_creatures_test.go: TestCombatThresherScenario",
}

var BlitzAutomaton = cardset.Entry{
	OracleID:     "19f344ee-031d-4346-9553-c5d32b02df43",
	Name:         "Blitz Automaton",
	New:          newBlitzAutomaton,
	Verification: "mtg/rules/prototype_artifact_creatures_test.go: TestBlitzAutomatonScenario",
}

var CradleClearcutter = cardset.Entry{
	OracleID:     "225f53ef-da42-47a0-b773-d2aad3338c72",
	Name:         "Cradle Clearcutter",
	New:          newCradleClearcutter,
	Verification: "mtg/rules/prototype_artifact_creatures_test.go: TestCradleClearcutterScenario",
}

var RustGoliath = cardset.Entry{
	OracleID:     "61df991d-32e7-4b6b-b8e6-700ba915c3f2",
	Name:         "Rust Goliath",
	New:          newRustGoliath,
	Verification: "mtg/rules/prototype_artifact_creatures_test.go: TestRustGoliathScenario",
}

var FrogmyrEnforcer = cardset.Entry{
	OracleID:     "ccc7de36-aafc-4d87-80f7-fc2effc976eb",
	Name:         "Frogmyr Enforcer",
	New:          newFrogmyrEnforcer,
	Verification: "mtg/rules/prototype_artifact_creatures_test.go: TestFrogmyrEnforcerScenario",
}

var FallajiDragonEngine = cardset.Entry{
	OracleID:     "dc8cca38-df2a-4cef-973c-a482505be66f",
	Name:         "Fallaji Dragon Engine",
	New:          newFallajiDragonEngine,
	Verification: "mtg/rules/prototype_artifact_creatures_test.go: TestFallajiDragonEngineScenario",
}

var DepthChargeColossus = cardset.Entry{
	OracleID:     "f376f4c6-4526-4ae3-a84a-e8ea50fe78b6",
	Name:         "Depth Charge Colossus",
	New:          newDepthChargeColossus,
	Verification: "mtg/rules/prototype_artifact_creatures_test.go: TestDepthChargeColossusScenario",
}

func prototypeArtifactCreature(name string, identity color.Color, manaCost cost.Mana, subtypes []types.Sub, power, toughness int, prototypeCost cost.Mana, prototypePower, prototypeToughness int, oracleText string, staticAbilities []game.StaticAbility) *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(identity),
		CardFace: game.CardFace{
			Name:            name,
			ManaCost:        opt.Val(manaCost),
			Types:           []types.Card{types.Artifact, types.Creature},
			Subtypes:        subtypes,
			Power:           opt.Val(game.PT{Value: power}),
			Toughness:       opt.Val(game.PT{Value: toughness}),
			Prototype:       opt.Val(game.PrototypeProfile{ManaCost: prototypeCost, Power: game.PT{Value: prototypePower}, Toughness: game.PT{Value: prototypeToughness}}),
			StaticAbilities: staticAbilities,
			OracleText:      oracleText,
		},
	}
}

func newGoringWarplow() *game.CardDef {
	return prototypeArtifactCreature("Goring Warplow", color.Black, cost.Mana{cost.O(6)}, []types.Sub{types.Construct}, 5, 4,
		cost.Mana{cost.O(1), cost.B}, 1, 1,
		"Prototype {1}{B} — 1/1 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nDeathtouch",
		[]game.StaticAbility{game.DeathtouchStaticBody})
}

func newCombatThresher() *game.CardDef {
	def := prototypeArtifactCreature("Combat Thresher", color.White, cost.Mana{cost.O(7)}, []types.Sub{types.Construct}, 3, 3,
		cost.Mana{cost.O(2), cost.W}, 1, 1,
		"Prototype {2}{W} — 1/1 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nDouble strike\nWhen this creature enters, draw a card.",
		[]game.StaticAbility{game.DoubleStrikeStaticBody})
	def.TriggeredAbilities = []game.TriggeredAbility{{
		Text: "When this creature enters, draw a card.",
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event:  game.EventPermanentEnteredBattlefield,
			Source: game.TriggerSourceSelf,
		}},
		Content: game.AbilityContent{
			Modes: []game.Mode{{Sequence: []game.Instruction{{Primitive: game.Draw{
				Amount: game.Fixed(1),
				Player: game.ControllerReference(),
			}}}}},
			MinModes: 1,
			MaxModes: 1,
		},
	}}
	return def
}

func newBlitzAutomaton() *game.CardDef {
	return prototypeArtifactCreature("Blitz Automaton", color.Red, cost.Mana{cost.O(7)}, []types.Sub{types.Construct}, 6, 4,
		cost.Mana{cost.O(2), cost.R}, 3, 2,
		"Prototype {2}{R} — 3/2 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nHaste",
		[]game.StaticAbility{game.HasteStaticBody})
}

func newCradleClearcutter() *game.CardDef {
	def := prototypeArtifactCreature("Cradle Clearcutter", color.Green, cost.Mana{cost.O(6)}, []types.Sub{types.Golem}, 3, 6,
		cost.Mana{cost.O(2), cost.G}, 1, 3,
		"Prototype {2}{G} — 1/3 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\n{T}: Add an amount of {G} equal to this creature's power.", nil)
	def.ManaAbilities = []game.ManaAbility{{
		Text:            "{T}: Add an amount of {G} equal to this creature's power.",
		AdditionalCosts: []cost.Additional{{Kind: cost.AdditionalTap}},
		Content: game.AbilityContent{
			Modes: []game.Mode{{Sequence: []game.Instruction{{Primitive: game.AddMana{
				Amount: game.Dynamic(game.DynamicAmount{
					Kind:       game.DynamicAmountObjectPower,
					Multiplier: 1,
					Object:     game.SourcePermanentReference(),
				}),
				ManaColor: mana.G,
			}}}}},
			MinModes: 1,
			MaxModes: 1,
		},
	}}
	return def
}

func newRustGoliath() *game.CardDef {
	return prototypeArtifactCreature("Rust Goliath", color.Green, cost.Mana{cost.O(10)}, []types.Sub{types.Construct}, 10, 10,
		cost.Mana{cost.O(3), cost.G, cost.G}, 3, 5,
		"Prototype {3}{G}{G} — 3/5 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nReach, trample",
		[]game.StaticAbility{game.ReachStaticBody, game.TrampleStaticBody})
}

func newFrogmyrEnforcer() *game.CardDef {
	def := prototypeArtifactCreature("Frogmyr Enforcer", color.Red, cost.Mana{cost.O(7)}, []types.Sub{types.Frog, types.Myr}, 4, 4,
		cost.Mana{cost.O(3), cost.R}, 2, 2,
		"Prototype {3}{R} — 2/2 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nAffinity for artifacts (This spell costs {1} less to cast for each artifact you control.)",
		[]game.StaticAbility{{
			Text: "This spell costs {1} less to cast for each artifact you control.",
			RuleEffects: []game.RuleEffect{{
				Kind:           game.RuleEffectCostModifier,
				AffectedSource: true,
				Duration:       game.DurationPermanent,
				CostModifier: game.CostModifier{
					Kind:               game.CostModifierSpell,
					PerObjectReduction: 1,
					CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou},
				},
			}},
		}})
	return def
}

func newFallajiDragonEngine() *game.CardDef {
	def := prototypeArtifactCreature("Fallaji Dragon Engine", color.Red, cost.Mana{cost.O(8)}, []types.Sub{types.Dragon}, 5, 5,
		cost.Mana{cost.O(2), cost.R}, 1, 3,
		"Prototype {2}{R} — 1/3 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nFlying\n{2}: This creature gets +1/+0 until end of turn.",
		[]game.StaticAbility{game.FlyingStaticBody})
	def.ActivatedAbilities = []game.ActivatedAbility{{
		Text:     "{2}: This creature gets +1/+0 until end of turn.",
		ManaCost: opt.Val(cost.Mana{cost.O(2)}),
		Content: game.AbilityContent{
			Modes: []game.Mode{{Sequence: []game.Instruction{{Primitive: game.ModifyPT{
				Object:     game.SourcePermanentReference(),
				PowerDelta: game.Fixed(1),
				Duration:   game.DurationUntilEndOfTurn,
			}}}}},
			MinModes: 1,
			MaxModes: 1,
		},
	}}
	return def
}

func newDepthChargeColossus() *game.CardDef {
	def := prototypeArtifactCreature("Depth Charge Colossus", color.Blue, cost.Mana{cost.O(9)}, []types.Sub{types.Dreadnought}, 9, 9,
		cost.Mana{cost.O(4), cost.U, cost.U}, 6, 6,
		"Prototype {4}{U}{U} — 6/6 (You may cast this spell with different mana cost, color, and size. It keeps its abilities and types.)\nThis creature doesn't untap during your untap step.\n{3}: Untap this creature.",
		[]game.StaticAbility{func() game.StaticAbility {
			ability := game.DoesntUntapStaticBody
			ability.Text = "This creature doesn't untap during your untap step."
			return ability
		}()})
	def.ActivatedAbilities = []game.ActivatedAbility{{
		Text:     "{3}: Untap this creature.",
		ManaCost: opt.Val(cost.Mana{cost.O(3)}),
		Content: game.AbilityContent{
			Modes: []game.Mode{{Sequence: []game.Instruction{{Primitive: game.Untap{
				Object: game.SourcePermanentReference(),
			}}}}},
			MinModes: 1,
			MaxModes: 1,
		},
	}}
	return def
}
