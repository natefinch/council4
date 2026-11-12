package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	tokenT "github.com/natefinch/council4/mtg/cards/tokens/t"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	exquisiteHuntmasterOracleID = "393ef069-ea6b-4969-b878-d942281cc918"
	soulOfEternityOracleID      = "4d211447-9f18-4d23-8f19-cdf6ad3c68ba"
	angelOfIndemnityOracleID    = "4f57a320-b6b5-49de-af0a-f2e2649758a7"
	briarbladeAdeptOracleID     = "6680a5f7-3ac1-4550-a29a-9553ad7ae465"
	impulsivePilfererOracleID   = "7d9fc9e7-d80b-49c3-871c-ed25b3059ae8"
	kinsbaileCourierOracleID    = "92c0346d-c375-4851-9806-5d3b9b83425f"
	broodmateTyrantOracleID     = "9b984236-39cf-4552-827b-e81c26cfb388"
	finCladeFugitivesOracleID   = "dad2965a-ceb4-42ac-bb62-de7eea500011"
	troveTrackerOracleID        = "f27972da-ad2f-4df5-82d3-c909cce8cc91"
)

var (
	ExquisiteHuntmaster = cardset.Entry{
		OracleID: exquisiteHuntmasterOracleID, Name: "Exquisite Huntmaster", New: newExquisiteHuntmaster,
		Verification: "mtg/rules/encore_creatures_test.go: TestExquisiteHuntmasterEncoreCopyCreatesElfAfterDying",
	}
	SoulOfEternity = cardset.Entry{
		OracleID: soulOfEternityOracleID, Name: "Soul of Eternity", New: newSoulOfEternity,
		Verification: "mtg/rules/encore_creatures_test.go: TestSoulOfEternityEncoreCopyUsesControllerLifeTotal",
	}
	AngelOfIndemnity = cardset.Entry{
		OracleID: angelOfIndemnityOracleID, Name: "Angel of Indemnity", New: newAngelOfIndemnity,
		Verification: "mtg/rules/encore_creatures_test.go: TestAngelOfIndemnityEncoreCopyReturnsTargetPermanentCard",
	}
	BriarbladeAdept = cardset.Entry{
		OracleID: briarbladeAdeptOracleID, Name: "Briarblade Adept", New: newBriarbladeAdept,
		Verification: "mtg/rules/encore_creatures_test.go: TestBriarbladeAdeptEncoreCopyWeakensCreatureWhenAttacking",
	}
	ImpulsivePilferer = cardset.Entry{
		OracleID: impulsivePilfererOracleID, Name: "Impulsive Pilferer", New: newImpulsivePilferer,
		Verification: "mtg/rules/encore_creatures_test.go: TestImpulsivePilfererEncoreCopyCreatesTreasureAfterDying",
	}
	KinsbaileCourier = cardset.Entry{
		OracleID: kinsbaileCourierOracleID, Name: "Kinsbaile Courier", New: newKinsbaileCourier,
		Verification: "mtg/rules/encore_creatures_test.go: TestKinsbaileCourierEncoreCopyAddsCounterOnEntry",
	}
	BroodmateTyrant = cardset.Entry{
		OracleID: broodmateTyrantOracleID, Name: "Broodmate Tyrant", New: newBroodmateTyrant,
		Verification: "mtg/rules/encore_creatures_test.go: TestBroodmateTyrantEncoreCopyCreatesFlyingDragon",
	}
	FinCladeFugitives = cardset.Entry{
		OracleID: finCladeFugitivesOracleID, Name: "Fin-Clade Fugitives", New: newFinCladeFugitives,
		Verification: "mtg/rules/encore_creatures_test.go: TestFinCladeFugitivesEncoreCopyRejectsSmallBlocker",
	}
	TroveTracker = cardset.Entry{
		OracleID: troveTrackerOracleID, Name: "Trove Tracker", New: newTroveTracker,
		Verification: "mtg/rules/encore_creatures_test.go: TestTroveTrackerEncoreCopyDrawsWhenItDies",
	}
)

var encoreElfWarriorToken = &game.CardDef{CardFace: game.CardFace{
	Name:      "Elf Warrior",
	Colors:    []color.Color{color.Green},
	Types:     []types.Card{types.Creature},
	Subtypes:  []types.Sub{types.Sub("Elf"), types.Sub("Warrior")},
	Power:     opt.Val(game.PT{Value: 1}),
	Toughness: opt.Val(game.PT{Value: 1}),
}}

var encoreBroodmateDragonToken = &game.CardDef{CardFace: game.CardFace{
	Name:            "Dragon",
	Colors:          []color.Color{color.Red},
	Types:           []types.Card{types.Creature},
	Subtypes:        []types.Sub{types.Sub("Dragon")},
	Power:           opt.Val(game.PT{Value: 5}),
	Toughness:       opt.Val(game.PT{Value: 5}),
	StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
}}

const (
	exquisiteHuntmasterDiesText = "When this creature dies, create a 1/1 green Elf Warrior creature token."
	soulOfEternityPowerText     = "Soul of Eternity's power and toughness are each equal to your life total."
	angelOfIndemnityEnterText   = "When this creature enters, return target permanent card with mana value 4 or less from your graveyard to the battlefield."
	briarbladeAdeptAttackText   = "Whenever this creature attacks, target creature an opponent controls gets -1/-1 until end of turn."
	impulsivePilfererDiesText   = "When this creature dies, create a Treasure token. (It's an artifact with \"{T}, Sacrifice this token: Add one mana of any color.\")"
	kinsbaileCourierEnterText   = "When this creature enters, put a +1/+1 counter on target creature."
	broodmateTyrantEnterText    = "When this creature enters, create a 5/5 red Dragon creature token with flying."
	finCladeFugitivesStaticText = "This creature can't be blocked by creatures with power 2 or less."
	troveTrackerDiesText        = "When this creature dies, draw a card."
)

func encoreCreature(face game.CardFace, encoreCost cost.Mana) game.CardFace {
	face.ActivatedAbilities = append(face.ActivatedAbilities, game.EncoreActivatedBody(encoreCost))
	return face
}

func encoreSelfETB(text string, targets []game.TargetSpec, instructions ...game.Instruction) game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: text,
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event: game.EventPermanentEnteredBattlefield, Source: game.TriggerSourceSelf,
		}},
		Content: game.Mode{Text: text, Targets: targets, Sequence: instructions}.Ability(),
	}
}

func encoreSelfDies(text string, instructions ...game.Instruction) game.TriggeredAbility {
	return game.TriggeredAbility{
		Text: text,
		Trigger: game.TriggerCondition{Type: game.TriggerWhen, Pattern: game.TriggerPattern{
			Event: game.EventPermanentDied, Source: game.TriggerSourceSelf,
			SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
		}},
		Content: game.Mode{Text: text, Sequence: instructions}.Ability(),
	}
}

func newExquisiteHuntmaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: encoreCreature(game.CardFace{
			Name: "Exquisite Huntmaster", ManaCost: opt.Val(cost.Mana{cost.O(3), cost.B}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Elf"), types.Sub("Warrior")},
			Colors: []color.Color{color.Black}, Power: opt.Val(game.PT{Value: 4}), Toughness: opt.Val(game.PT{Value: 2}),
			OracleText: exquisiteHuntmasterDiesText + "\nEncore {4}{B} ({4}{B}, Exile this card from your graveyard: For each opponent, create a token copy that attacks that opponent this turn if able. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
			TriggeredAbilities: []game.TriggeredAbility{encoreSelfDies(exquisiteHuntmasterDiesText, game.Instruction{Primitive: game.CreateToken{
				Amount: game.Fixed(1), Source: game.TokenDef(encoreElfWarriorToken),
			}})},
		}, cost.Mana{cost.O(4), cost.B}),
	}
}

func newSoulOfEternity() *game.CardDef {
	life := game.DynamicValue{Kind: game.DynamicValueControllerLifeTotal}
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: encoreCreature(game.CardFace{
			Name: "Soul of Eternity", ManaCost: opt.Val(cost.Mana{cost.O(5), cost.W, cost.W}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Avatar")}, Colors: []color.Color{color.White},
			Power: opt.Val(game.PT{IsStar: true}), Toughness: opt.Val(game.PT{IsStar: true}),
			DynamicPower: opt.Val(life), DynamicToughness: opt.Val(life),
			OracleText: soulOfEternityPowerText + "\nEncore {7}{W}{W} ({7}{W}{W}, Exile this card from your graveyard: For each opponent, create a token copy that attacks that opponent this turn if able. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
		}, cost.Mana{cost.O(7), cost.W, cost.W}),
	}
}

func newAngelOfIndemnity() *game.CardDef {
	const manaValueFour = 4
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: encoreCreature(game.CardFace{
			Name: "Angel of Indemnity", ManaCost: opt.Val(cost.Mana{cost.O(5), cost.W}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Angel"), types.Sub("Warrior")},
			Colors: []color.Color{color.White}, Power: opt.Val(game.PT{Value: 5}), Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody, game.LifelinkStaticBody},
			OracleText:      "Flying, lifelink\n" + angelOfIndemnityEnterText + "\nEncore {6}{W}{W} ({6}{W}{W}, Exile this card from your graveyard: For each opponent, create a token copy that attacks that opponent this turn if able. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
			TriggeredAbilities: []game.TriggeredAbility{encoreSelfETB(angelOfIndemnityEnterText,
				[]game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Constraint: "target permanent card with mana value 4 or less from your graveyard",
					Allow: game.TargetAllowCard, TargetZone: zone.Graveyard,
					Selection: opt.Val(game.Selection{
						RequiredTypesAny: []types.Card{types.Artifact, types.Battle, types.Creature, types.Enchantment, types.Land, types.Planeswalker},
						Controller:       game.ControllerYou,
						ManaValue:        opt.Val(compare.Int{Op: compare.LessOrEqual, Value: manaValueFour}),
					}),
				}}, game.Instruction{Primitive: game.PutOnBattlefield{Source: game.CardBattlefieldSource(game.CardReference{
					Kind: game.CardReferenceTarget, TargetIndex: 0,
				}), Recipient: opt.Val(game.ControllerReference())}})},
		}, cost.Mana{cost.O(6), cost.W, cost.W}),
	}
}

func newBriarbladeAdept() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: encoreCreature(game.CardFace{
			Name: "Briarblade Adept", ManaCost: opt.Val(cost.Mana{cost.O(4), cost.B}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Elf"), types.Sub("Assassin")},
			Colors: []color.Color{color.Black}, Power: opt.Val(game.PT{Value: 3}), Toughness: opt.Val(game.PT{Value: 4}),
			OracleText: briarbladeAdeptAttackText + "\nEncore {3}{B} ({3}{B}, Exile this card from your graveyard: For each opponent, create a token copy that attacks that opponent this turn if able. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: briarbladeAdeptAttackText,
				Trigger: game.TriggerCondition{Type: game.TriggerWhenever, Pattern: game.TriggerPattern{
					Event: game.EventAttackerDeclared, Source: game.TriggerSourceSelf,
				}},
				Content: game.Mode{
					Text: briarbladeAdeptAttackText,
					Targets: []game.TargetSpec{{
						MinTargets: 1, MaxTargets: 1, Constraint: "target creature an opponent controls",
						Allow:     game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
					}},
					Sequence: []game.Instruction{{Primitive: game.ModifyPT{
						Object: game.TargetPermanentReference(0), PowerDelta: game.Fixed(-1), ToughnessDelta: game.Fixed(-1),
						Duration: game.DurationUntilEndOfTurn,
					}}},
				}.Ability(),
			}},
		}, cost.Mana{cost.O(3), cost.B}),
	}
}

func newImpulsivePilferer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: encoreCreature(game.CardFace{
			Name: "Impulsive Pilferer", ManaCost: opt.Val(cost.Mana{cost.R}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Goblin"), types.Sub("Pirate")},
			Colors: []color.Color{color.Red}, Power: opt.Val(game.PT{Value: 1}), Toughness: opt.Val(game.PT{Value: 1}),
			OracleText: impulsivePilfererDiesText + "\nEncore {3}{R} ({3}{R}, Exile this card from your graveyard: For each opponent, create a token copy that attacks that opponent this turn if able. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
			TriggeredAbilities: []game.TriggeredAbility{encoreSelfDies(impulsivePilfererDiesText, game.Instruction{Primitive: game.CreateToken{
				Amount: game.Fixed(1), Source: game.TokenDef(tokenT.TreasureToken3c5493746c3742e08d88a8555d46732d),
			}})},
		}, cost.Mana{cost.O(3), cost.R}),
	}
}

func newKinsbaileCourier() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: encoreCreature(game.CardFace{
			Name: "Kinsbaile Courier", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.W}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Kithkin"), types.Sub("Soldier")},
			Colors: []color.Color{color.White}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 1}),
			OracleText: kinsbaileCourierEnterText + "\nEncore {2}{W} ({2}{W}, Exile this card from your graveyard: For each opponent, create a token copy that attacks that opponent this turn if able. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
			TriggeredAbilities: []game.TriggeredAbility{encoreSelfETB(kinsbaileCourierEnterText,
				[]game.TargetSpec{{
					MinTargets: 1, MaxTargets: 1, Constraint: "target creature", Allow: game.TargetAllowPermanent,
					Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
				}}, game.Instruction{Primitive: game.AddCounter{
					Object: game.TargetPermanentReference(0), Amount: game.Fixed(1), CounterKind: counter.PlusOnePlusOne,
				}})},
		}, cost.Mana{cost.O(2), cost.W}),
	}
}

func newBroodmateTyrant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: encoreCreature(game.CardFace{
			Name: "Broodmate Tyrant", ManaCost: opt.Val(cost.Mana{cost.O(4), cost.B, cost.R, cost.G}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Dragon")},
			Colors: []color.Color{color.Black, color.Green, color.Red},
			Power:  opt.Val(game.PT{Value: 5}), Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{game.FlyingStaticBody},
			OracleText:      "Flying\n" + broodmateTyrantEnterText + "\nEncore {5}{B}{R}{G}",
			TriggeredAbilities: []game.TriggeredAbility{encoreSelfETB(broodmateTyrantEnterText, nil, game.Instruction{Primitive: game.CreateToken{
				Amount: game.Fixed(1), Source: game.TokenDef(encoreBroodmateDragonToken),
			}})},
		}, cost.Mana{cost.O(5), cost.B, cost.R, cost.G}),
	}
}

func newFinCladeFugitives() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: encoreCreature(game.CardFace{
			Name: "Fin-Clade Fugitives", ManaCost: opt.Val(cost.Mana{cost.O(5), cost.G}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Elf"), types.Sub("Salamander"), types.Sub("Rogue")},
			Colors: []color.Color{color.Green}, Power: opt.Val(game.PT{Value: 7}), Toughness: opt.Val(game.PT{Value: 4}),
			OracleText: finCladeFugitivesStaticText + "\nEncore {4}{G} ({4}{G}, Exile this card from your graveyard: For each opponent, create a token copy that attacks that opponent this turn if able. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
			StaticAbilities: []game.StaticAbility{{
				Text: finCladeFugitivesStaticText,
				RuleEffects: []game.RuleEffect{{
					Kind: game.RuleEffectCantBeBlockedByCreaturesWith, AffectedSource: true,
					BlockerRestriction: game.BlockerRestriction{Kind: game.BlockerRestrictionPowerLessOrEqual, Power: 2},
				}},
			}},
		}, cost.Mana{cost.O(4), cost.G}),
	}
}

func newTroveTracker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: encoreCreature(game.CardFace{
			Name: "Trove Tracker", ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
			Types: []types.Card{types.Creature}, Subtypes: []types.Sub{types.Sub("Human"), types.Sub("Pirate")},
			Colors: []color.Color{color.Blue}, Power: opt.Val(game.PT{Value: 2}), Toughness: opt.Val(game.PT{Value: 2}),
			OracleText: troveTrackerDiesText + "\nEncore {5}{U}{U} ({5}{U}{U}, Exile this card from your graveyard: For each opponent, create a token copy that attacks that opponent this turn if able. They gain haste. Sacrifice them at the beginning of the next end step. Activate only as a sorcery.)",
			TriggeredAbilities: []game.TriggeredAbility{encoreSelfDies(troveTrackerDiesText, game.Instruction{Primitive: game.Draw{
				Amount: game.Fixed(1), Player: game.ControllerReference(),
			}})},
		}, cost.Mana{cost.O(5), cost.U, cost.U}),
	}
}
