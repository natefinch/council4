package authored

import (
	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

const (
	reiterateOracleID = "38394ee3-6726-4f91-bc25-36bce0c6aab9"
	allayOracleID     = "664d982a-5cd8-4289-afb8-fa2e083042c9"
	capsizeOracleID   = "77637eff-2963-4402-88f3-ca346f762fc8"
)

var (
	reiterateProgram = game.ProgramReference{ID: reiterateOracleID + "/spell/0", Version: "1"}
	allayProgram     = game.ProgramReference{ID: allayOracleID + "/spell/0", Version: "1"}
	capsizeProgram   = game.ProgramReference{ID: capsizeOracleID + "/spell/0", Version: "1"}
)

const buybackReminderText = "Buyback {3} (You may pay an additional {3} as you cast this spell. If you do, put this card into your hand as it resolves.)"

const (
	reiterateOracleText = buybackReminderText + "\nCopy target instant or sorcery spell. You may choose new targets for the copy."
	allayOracleText     = buybackReminderText + "\nDestroy target enchantment."
	capsizeOracleText   = buybackReminderText + "\nReturn target permanent to its owner's hand."
)

var Reiterate = cardset.Entry{
	OracleID: reiterateOracleID,
	Name:     "Reiterate",
	New:      newReiterate,
	Programs: []cardprogram.Program{{Reference: reiterateProgram, Resolve: resolveTargetSpellCopy}},
	Verification: "mtg/rules/buyback_interaction_test.go: TestBuybackInteractionRegistryIdentityAndCosts; " +
		"TestBuybackInteractionReiterateCopiesPaidSpellChoicesAndClone; " +
		"TestBuybackInteractionReiteratePaidBuybackCopyHasNoPhysicalCard; " +
		"TestBuybackInteractionReiterateRejectsAbilityTargets; " +
		"TestBuybackInteractionReiterateAllIllegalTargetDoesNotBuyBack; " +
		"TestBuybackInteractionCastsPaymentsAndInstantTiming",
}

var Allay = cardset.Entry{
	OracleID: allayOracleID,
	Name:     "Allay",
	New:      newAllay,
	Programs: []cardprogram.Program{{Reference: allayProgram, Resolve: resolveDestroyTarget}},
	Verification: "mtg/rules/buyback_interaction_test.go: TestBuybackInteractionRegistryIdentityAndCosts; " +
		"TestBuybackInteractionAllayProtectionAndDestructionReplacement; " +
		"TestBuybackInteractionAllayCountered; TestBuybackInteractionTargetIncarnations; " +
		"TestBuybackInteractionCastsPaymentsAndInstantTiming",
}

var Capsize = cardset.Entry{
	OracleID: capsizeOracleID,
	Name:     "Capsize",
	New:      newCapsize,
	Programs: []cardprogram.Program{{Reference: capsizeProgram, Resolve: resolveCapsize}},
	Verification: "mtg/rules/buyback_interaction_test.go: TestBuybackInteractionRegistryIdentityAndCosts; " +
		"TestBuybackInteractionCapsizePermanentTypesTokensAndOwner; " +
		"TestBuybackInteractionCapsizeCommanderAndReplacement; TestBuybackInteractionTargetIncarnations; " +
		"TestBuybackInteractionCapsizeSimulatorAndClone; TestBuybackInteractionCastsPaymentsAndInstantTiming",
}

func buybackStatic() game.StaticAbility {
	return game.StaticAbility{
		Text:           "Buyback {3}",
		ZoneOfFunction: zone.Stack,
		KeywordAbilities: []game.KeywordAbility{
			game.BuybackKeyword{Cost: cost.Mana{cost.O(3)}},
		},
	}
}

func instantBuybackFace(name string, oracleText string, manaCost cost.Mana, cardColor color.Color) game.CardFace {
	return game.CardFace{
		Name:       name,
		OracleText: oracleText,
		ManaCost:   opt.Val(manaCost),
		Types:      []types.Card{types.Instant},
		Colors:     []color.Color{cardColor},
		StaticAbilities: []game.StaticAbility{
			buybackStatic(),
		},
	}
}

func newReiterate() *game.CardDef {
	target := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate: game.TargetPredicate{
			StackObjectKinds:  []game.StackObjectKind{game.StackSpell},
			SpellCardTypesAny: []types.Card{types.Instant, types.Sorcery},
		},
	}
	face := instantBuybackFace("Reiterate", reiterateOracleText, cost.Mana{cost.O(1), cost.R, cost.R}, color.Red)
	face.SpellAbility = opt.Val(game.Mode{
		Text:    "Copy target instant or sorcery spell. You may choose new targets for the copy.",
		Targets: []game.TargetSpec{target},
		Program: reiterateProgram,
	}.Ability())
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace:      face,
	}
}

func newAllay() *game.CardDef {
	target := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowPermanent,
		Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Enchantment}}),
	}
	face := instantBuybackFace("Allay", allayOracleText, cost.Mana{cost.O(1), cost.W}, color.White)
	face.SpellAbility = opt.Val(game.Mode{
		Text:    "Destroy target enchantment.",
		Targets: []game.TargetSpec{target},
		Program: allayProgram,
	}.Ability())
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace:      face,
	}
}

func newCapsize() *game.CardDef {
	face := instantBuybackFace("Capsize", capsizeOracleText, cost.Mana{cost.O(1), cost.U, cost.U}, color.Blue)
	face.SpellAbility = opt.Val(game.Mode{
		Text: "Return target permanent to its owner's hand.",
		Targets: []game.TargetSpec{{
			MinTargets: 1,
			MaxTargets: 1,
			Allow:      game.TargetAllowPermanent,
		}},
		Program: capsizeProgram,
	}.Ability())
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace:      face,
	}
}

func resolveCapsize(ctx cardprogram.Context) error {
	_, err := ctx.Execute(game.MovePermanent{
		Object:      game.TargetPermanentReference(0),
		Destination: zone.Hand,
	})
	return err
}
