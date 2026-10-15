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
	bedevilProgram    = game.ProgramReference{ID: "bceecc64-96f1-4e7b-8904-0aef90377764/spell/0", Version: "1"}
	terminateProgram  = game.ProgramReference{ID: "6257c2fd-005f-41e3-8a72-af76df1eb134/spell/0", Version: "1"}
	vindicateProgram  = game.ProgramReference{ID: "63c1ac21-e3d8-40c2-8c09-3f31c52992ef/spell/0", Version: "1"}
	disenchantProgram = game.ProgramReference{ID: "a7e97fa9-4b72-4548-b854-5be5f18a6f1a/spell/0", Version: "1"}
	naturalizeProgram = game.ProgramReference{ID: "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728/spell/0", Version: "1"}
)

// Bedevil destroys one member of the printed artifact/creature/planeswalker union.
var Bedevil = cardset.Entry{
	OracleID: "bceecc64-96f1-4e7b-8904-0aef90377764", Name: "Bedevil", New: newBedevil,
	Programs:     []cardprogram.Program{{Reference: bedevilProgram, Resolve: resolveDestroyTarget}},
	Verification: "mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Bedevil, TestProductionDestructionProtectionAndReplacements/Bedevil, TestProductionDestructionIdentityAndCopies/Bedevil; TestProductionBatchLiveAndSimulator; CR 2026-09-25 608.2b, 701.8",
}

// Terminate prevents regeneration, not shield or umbra-armor replacement.
var Terminate = cardset.Entry{
	OracleID: "6257c2fd-005f-41e3-8a72-af76df1eb134", Name: "Terminate", New: newTerminate,
	Programs:     []cardprogram.Program{{Reference: terminateProgram, Resolve: resolveDestroyWithoutRegeneration}},
	Verification: "mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Terminate, TestProductionDestructionProtectionAndReplacements/Terminate, TestProductionDestructionIdentityAndCopies/Terminate; TestProductionBatchLiveAndSimulator; CR 2026-09-25 608.2b, 701.8, 701.19",
}

// Vindicate retains sorcery timing while targeting any permanent type.
var Vindicate = cardset.Entry{
	OracleID: "63c1ac21-e3d8-40c2-8c09-3f31c52992ef", Name: "Vindicate", New: newVindicate,
	Programs:     []cardprogram.Program{{Reference: vindicateProgram, Resolve: resolveDestroyTarget}},
	Verification: "mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Vindicate, TestProductionDestructionProtectionAndReplacements/Vindicate, TestProductionDestructionIdentityAndCopies/Vindicate; TestProductionBatchLiveAndSimulator; CR 2026-09-25 307, 608.2b, 701.8",
}

// Disenchant keeps a target legal when it changes between artifact and enchantment.
var Disenchant = cardset.Entry{
	OracleID: "a7e97fa9-4b72-4548-b854-5be5f18a6f1a", Name: "Disenchant", New: newDisenchant,
	Programs:     []cardprogram.Program{{Reference: disenchantProgram, Resolve: resolveDestroyTarget}},
	Verification: "mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Disenchant, TestProductionDestructionProtectionAndReplacements/Disenchant, TestProductionDestructionIdentityAndCopies/Disenchant; TestProductionBatchLiveAndSimulator; WotC ruling 2004-10-04; CR 2026-09-25 608.2b, 701.8",
}

// Naturalize uses the same nonmodal artifact-or-enchantment union as Disenchant.
var Naturalize = cardset.Entry{
	OracleID: "bdb3ca68-ec1f-4e16-81cc-d23f8f52c728", Name: "Naturalize", New: newNaturalize,
	Programs:     []cardprogram.Program{{Reference: naturalizeProgram, Resolve: resolveDestroyTarget}},
	Verification: "mtg/rules/production_destruction_test.go: TestProductionDestructionCastAndTypes/Naturalize, TestProductionDestructionProtectionAndReplacements/Naturalize, TestProductionDestructionIdentityAndCopies/Naturalize; TestProductionBatchLiveAndSimulator; CR 2026-09-25 608.2b, 701.8",
}

func newBedevil() *game.CardDef {
	const text = "Destroy target artifact, creature, or planeswalker."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Bedevil", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.B, cost.B, cost.R}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Black, color.Red},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: bedevilProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Creature, types.Planeswalker}}),
			}}}.Ability()),
		},
	}
}

func newTerminate() *game.CardDef {
	const text = "Destroy target creature. It can't be regenerated."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Terminate", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.B, cost.R}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Black, color.Red},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: terminateProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
			}}}.Ability()),
		},
	}
}

func newVindicate() *game.CardDef {
	const text = "Destroy target permanent."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name: "Vindicate", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W, cost.B}),
			Types: []types.Card{types.Sorcery}, Colors: []color.Color{color.White, color.Black},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: vindicateProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
			}}}.Ability()),
		},
	}
}

func newDisenchant() *game.CardDef {
	const text = "Destroy target artifact or enchantment."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Disenchant", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.W}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.White},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: disenchantProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
			}}}.Ability()),
		},
	}
}

func newNaturalize() *game.CardDef {
	const text = "Destroy target artifact or enchantment."
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Naturalize", OracleText: text, ManaCost: opt.Val(cost.Mana{cost.O(1), cost.G}),
			Types: []types.Card{types.Instant}, Colors: []color.Color{color.Green},
			SpellAbility: opt.Val(game.Mode{Text: text, Program: naturalizeProgram, Targets: []game.TargetSpec{{
				MinTargets: 1, MaxTargets: 1, Allow: game.TargetAllowPermanent,
				Selection: opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
			}}}.Ability()),
		},
	}
}

func resolveDestroyTarget(ctx cardprogram.Context) error {
	_, err := ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0)})
	return err
}

func resolveDestroyWithoutRegeneration(ctx cardprogram.Context) error {
	_, err := ctx.Destroy(game.Destroy{Object: game.TargetPermanentReference(0), PreventRegeneration: true})
	return err
}
