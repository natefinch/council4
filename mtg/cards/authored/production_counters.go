package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

var (
	anOfferProgram        = game.ProgramReference{ID: "234a734b-ba28-4f1b-9d01-3c3e7d516590/spell/0", Version: "1"}
	counterspellProgram   = game.ProgramReference{ID: "cc187110-1148-4090-bbb8-e205694a39f5/spell/0", Version: "1"}
	negateProgram         = game.ProgramReference{ID: "3407fe41-fdd3-4119-8f70-4bc4590a379f/spell/0", Version: "1"}
	essenceScatterProgram = game.ProgramReference{ID: "46665089-aa3d-44c3-964d-6638dfbb5782/spell/0", Version: "1"}
	disallowProgram       = game.ProgramReference{ID: "88b51e15-6630-4e14-a6b8-db0aa12e34ef/spell/0", Version: "1"}
)

// AnOfferYouCantRefuse creates Treasures even for a legal uncounterable target.
var AnOfferYouCantRefuse = cardset.Entry{
	OracleID: "234a734b-ba28-4f1b-9d01-3c3e7d516590", Name: "An Offer You Can't Refuse", New: newAnOfferYouCantRefuse,
	Programs:     []cardprogram.Program{{Reference: anOfferProgram, Resolve: resolveAnOfferYouCantRefuse}},
	Verification: "mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/An_Offer_You_Can't_Refuse, TestProductionCountersDepartedTargetAndClone, TestProductionOfferTokensAndCounterReplacements, TestProductionOfferTreasuresPayAndReplace; TestProductionBatchLiveAndSimulator; WotC 2022-04-29; CR 2026-09-25 111.10a, 608.2b/h, 700.7, 701.6",
}

// Counterspell targets spells, not activated or triggered abilities.
var Counterspell = cardset.Entry{
	OracleID: "cc187110-1148-4090-bbb8-e205694a39f5", Name: "Counterspell", New: newCounterspell,
	Programs:     []cardprogram.Program{{Reference: counterspellProgram, Resolve: resolveCounterTarget}},
	Verification: "mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/Counterspell, TestProductionCountersDepartedTargetAndClone, TestProductionCountersRetargetedCopy; TestProductionBatchLiveAndSimulator; CR 2026-09-25 608.2b, 701.6",
}

// Negate excludes any spell carrying the creature type, including artifact creatures.
var Negate = cardset.Entry{
	OracleID: "3407fe41-fdd3-4119-8f70-4bc4590a379f", Name: "Negate", New: newNegate,
	Programs:     []cardprogram.Program{{Reference: negateProgram, Resolve: resolveCounterTarget}},
	Verification: "mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/Negate, TestProductionCountersDepartedTargetAndClone, TestProductionCountersRetargetedCopy; TestProductionBatchLiveAndSimulator; WotC 2026-01-27; CR 2026-09-25 608.2b, 701.6",
}

// EssenceScatter accepts creature spells even when they have additional types.
var EssenceScatter = cardset.Entry{
	OracleID: "46665089-aa3d-44c3-964d-6638dfbb5782", Name: "Essence Scatter", New: newEssenceScatter,
	Programs:     []cardprogram.Program{{Reference: essenceScatterProgram, Resolve: resolveCounterTarget}},
	Verification: "mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/Essence_Scatter, TestProductionCountersDepartedTargetAndClone, TestProductionCountersRetargetedCopy; TestProductionBatchLiveAndSimulator; WotC 2026-03-20; CR 2026-09-25 608.2b, 701.6",
}

// Disallow can counter stack abilities independently of their departed sources.
var Disallow = cardset.Entry{
	OracleID: "88b51e15-6630-4e14-a6b8-db0aa12e34ef", Name: "Disallow", New: newDisallow,
	Programs:     []cardprogram.Program{{Reference: disallowProgram, Resolve: resolveCounterTarget}},
	Verification: "mtg/rules/production_counters_test.go: TestProductionCountersCastAndKinds/Disallow, TestProductionCountersDepartedTargetAndClone, TestProductionDisallowSourceDepartureAndManaAbility, TestProductionDisallowDelayedTrigger; TestProductionBatchLiveAndSimulator; WotC 2017-02-09; CR 2026-09-25 113.7a, 603.7b, 605, 608.2b, 701.6",
}

func newAnOfferYouCantRefuse() *game.CardDef {
	const text = `Counter target noncreature spell. Its controller creates two Treasure tokens. (They're artifacts with "{T}, Sacrifice this token: Add one mana of any color.")`
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "An Offer You Can't Refuse", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: anOfferProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}, ExcludedSpellCardTypes: []types.Card{types.Creature}},
			}}}.Ability()),
		},
	}
}

func newCounterspell() *game.CardDef {
	const text = "Counter target spell."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Counterspell", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.U, cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: counterspellProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}},
			}}}.Ability()),
		},
	}
}

func newNegate() *game.CardDef {
	const text = "Counter target noncreature spell."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Negate", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: negateProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}, ExcludedSpellCardTypes: []types.Card{types.Creature}},
			}}}.Ability()),
		},
	}
}

func newEssenceScatter() *game.CardDef {
	const text = "Counter target creature spell."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Essence Scatter", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: essenceScatterProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell}, SpellCardTypes: []types.Card{types.Creature}},
			}}}.Ability()),
		},
	}
}

func newDisallow() *game.CardDef {
	const text = "Counter target spell, activated ability, or triggered ability. (Mana abilities can't be targeted.)"
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Disallow", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.U, cost.U}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Blue},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: disallowProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowStackObject,
				Predicate: game.TargetPredicate{StackObjectKinds: []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility}},
			}}}.Ability()),
		},
	}
}

func resolveCounterTarget(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.CounterObject{Object: game.TargetStackObjectReference(0)})
	return err
}

func resolveAnOfferYouCantRefuse(ctx cardprogram.Context) error {
	controller, available, err := ctx.View().Player(game.ObjectControllerReference(game.TargetStackObjectReference(0)))
	if err != nil || !available {
		return err
	}
	if err := resolveCounterTarget(ctx); err != nil {
		return err
	}
	_, err = ctx.CreateTokens(&game.CardDef{CardFace: game.CardFace{
		Name: "Treasure", Types: []types.Card{types.Artifact}, Subtypes: []types.Sub{types.Treasure},
		ManaAbilities: []game.ManaAbility{game.TapSacrificeAnyColorManaAbility("{T}, Sacrifice this token: Add one mana of any color.")},
	}}, cardprogram.TokenRecipient{Player: controller, Count: 2})
	return err
}
