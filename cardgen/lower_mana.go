package cardgen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// imprintLinkKey is the canonical link identifier connecting an exile-from-hand
// imprint effect (the publisher) to a "one mana of any of the exiled card's
// colors" mana ability (the reader) on the same face (Chrome Mox). Both sides
// lower to this fixed key so the imprinted card is found by the source
// permanent's object identity at activation/resolution. It is card-name-blind:
// any face matching both wordings uses the same key.
const imprintLinkKey = "imprint"

// lowerReminderManaAbility preserves a parenthesized reminder mana ability such
// as "({T}: Add {R} or {G}.)" and consumes other rules-free reminder abilities.
func lowerReminderManaAbility(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, *shared.Diagnostic) {
	unsupported := func() *shared.Diagnostic {
		return executableDiagnostic(
			ability,
			"unsupported reminder ability",
			"the executable source backend does not yet lower reminder abilities",
		)
	}
	innerDocument, innerDiags, ok := syntax.ReminderInner()
	if !ok {
		return abilityLowering{}, unsupported()
	}
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	innerDiags = append(append([]shared.Diagnostic(nil), innerDiags...), compilerDiags...)
	if len(innerComp.Abilities) == 1 && isSemanticManaAbility(innerComp.Abilities[0]) {
		if len(innerDiags) != 0 ||
			len(innerComp.Syntax.Abilities) != 1 ||
			innerComp.Abilities[0].Kind != compiler.AbilityActivated {
			return abilityLowering{}, unsupported()
		}
		manaAbility, diagnostic := lowerManaAbility(
			"",
			innerComp.Abilities[0],
			&innerComp.Syntax.Abilities[0],
		)
		if diagnostic != nil {
			return abilityLowering{}, unsupported()
		}
		// The compiled reminder ability has no independent semantic elements;
		// all content is filtered as parenthesized. The consumed counts are all
		// zero, matching the empty CompiledAbility fields.
		return abilityLowering{
			manaAbility: opt.Val(manaAbility),
			consumed:    semanticConsumption{},
			sourceSpans: []shared.Span{ability.Span},
		}, nil
	}

	// Non-mana reminder abilities carry no semantic content beyond their
	// parenthesized explanation.
	return abilityLowering{
		sourceSpans: []shared.Span{ability.Span},
	}, nil
}

// lowerManaAbility lowers an activated mana ability into a game.ManaAbility.
// It accepts the same supported cost shapes as ordinary activated abilities,
// plus supported fixed-symbol, choice, and any-color mana output bodies. A
// single fixed self-damage rider ("<CARDNAME> deals N damage to you") or a
// fixed life-gain rider ("You gain N life") may accompany the add-mana effect,
// modelling painlands, the painland Talismans, and similar self-damaging or
// life-gaining mana sources; the lowered content already carries the matching
// source-dealt Damage or GainLife instruction. Unrecognised costs and bodies
// remain fail-closed.
func lowerManaAbility(
	cardName string,
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (game.ManaAbility, *shared.Diagnostic) {
	if len(ability.Content.Modes) != 0 {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activation modes",
			"the Payment Planner cannot safely choose modes for a mana ability",
		)
	}
	loweredAbility := ability
	var gatedContent opt.V[game.AbilityContent]
	if stripped, gated, ok := tronConditionalManaContent(ability); ok {
		loweredAbility = stripped
		gatedContent = opt.Val(gated)
	}
	shell, diagnostic := lowerActivationShell(cardName, loweredAbility, syntax)
	if diagnostic != nil {
		return game.ManaAbility{}, diagnostic
	}
	if len(shell.semanticContent.Effects) < 1 || len(shell.semanticContent.Effects) > 2 ||
		shell.semanticContent.Effects[0].Kind != compiler.EffectAddMana ||
		shell.semanticContent.Effects[0].Negated ||
		len(shell.semanticContent.Keywords) != 0 ||
		len(shell.semanticContent.Targets) != 0 ||
		(len(shell.semanticContent.Effects) == 2 &&
			!isSelfDamageToControllerRider(&shell.semanticContent.Effects[1]) &&
			!isGainLifeToControllerRider(&shell.semanticContent.Effects[1]) &&
			!isSourceStunRider(&shell.semanticContent.Effects[1]) &&
			!isManaSpendRider(&shell.semanticContent.Effects[1])) {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana effect",
			"the executable source backend supports only exact non-targeting add-mana content, optionally with a fixed self-damage or life-gain rider, in mana abilities",
		)
	}
	if shell.semanticContent.Effects[0].HasUnrecognizedSibling {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported mana symbol",
			"the executable source backend cannot lower this add-mana content",
		)
	}
	supportedZone := shell.zoneOfFunction == zone.Battlefield ||
		(shell.zoneOfFunction == zone.Hand && handManaAbilityCostSupported(shell.manaCost, shell.additionalCosts))
	if !supportedZone {
		return game.ManaAbility{}, executableDiagnostic(
			ability,
			"unsupported activation zone",
			"the Payment Planner supports mana abilities only on the battlefield or, with a self-exile-from-hand cost, in the hand",
		)
	}

	functionZone := shell.zoneOfFunction
	if functionZone == zone.Battlefield {
		functionZone = zone.None
	}

	content := shell.content
	if gatedContent.Exists {
		content = gatedContent.Val
	}
	return game.ManaAbility{
		Text:                shell.text,
		ManaCost:            shell.manaCost,
		Content:             content,
		Timing:              shell.timing,
		ActivationCondition: shell.activationCondition,
		AdditionalCosts:     shell.additionalCosts,
		ZoneOfFunction:      functionZone,
	}, nil
}

// handManaAbilityCostSupported reports whether a hand-activated mana ability has
// exactly the self-exile-from-hand cost shape — no mana and a single
// "Exile this card from your hand" cost — modelling Simian Spirit Guide and
// Elvish Spirit Guide. Other hand mana costs fail closed.
func handManaAbilityCostSupported(manaCost opt.V[cost.Mana], additionalCosts []cost.Additional) bool {
	if manaCost.Exists && len(manaCost.Val) != 0 {
		return false
	}
	return len(additionalCosts) == 1 &&
		additionalCosts[0].Kind == cost.AdditionalExileSource &&
		additionalCosts[0].Source == zone.Hand
}

// isSelfDamageToControllerRider reports whether effect is exactly a
// "<CARDNAME> deals N damage to you" rider, the only non-mana effect a mana
// ability may carry. It accepts only a fixed positive amount of source-dealt
// damage to the source's own controller with no target, no divided damage, and
// no additional damage riders, so unrelated deal-damage clauses cannot ride
// into a mana ability. This models the painlands ("This land deals 1 damage to
// you."), the painland Talismans, Ancient Tomb, and Tarnished Citadel, whose
// lowered content already carries the matching self-source Damage instruction.
func isSelfDamageToControllerRider(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectDealDamage &&
		effect.Exact &&
		!effect.Negated &&
		!effect.Optional &&
		!effect.Divided &&
		!effect.HasUnrecognizedSibling &&
		effect.DamageRecipientReference == parser.DamageRecipientReferenceYou &&
		len(effect.DamageRecipientSelectors) == 0 &&
		len(effect.Targets) == 0 &&
		!effect.HasSelfDamageRider &&
		!effect.HasSecondTargetDamageRider &&
		effect.TargetControllerDamageRiderRecipient == parser.DamageRecipientReferenceNone &&
		effect.Duration == compiler.DurationNone &&
		effect.DelayedTiming == 0 &&
		effect.Amount.Known &&
		!effect.Amount.VariableX &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		effect.Amount.Value >= 1
}

// isGainLifeToControllerRider reports whether effect is exactly a "You gain N
// life" rider, the life-gaining counterpart to isSelfDamageToControllerRider and
// the only other non-mana effect a mana ability may carry. It accepts only a
// fixed positive amount of life gained by the ability's own controller with no
// target, no references, and no duration, so unrelated gain-life clauses cannot
// ride into a mana ability. This models lands and rocks that gain life when
// tapped for mana (The Great Henge: "{T}: Add {G}{G}. You gain 2 life."), whose
// lowered content already carries the matching controller GainLife instruction.
func isGainLifeToControllerRider(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectGain &&
		effect.LifeObject &&
		effect.Context == parser.EffectContextController &&
		effect.Exact &&
		!effect.Negated &&
		!effect.Optional &&
		!effect.Divided &&
		!effect.HasUnrecognizedSibling &&
		len(effect.Targets) == 0 &&
		len(effect.References) == 0 &&
		effect.Duration == compiler.DurationNone &&
		effect.DelayedTiming == 0 &&
		effect.Amount.Known &&
		!effect.Amount.VariableX &&
		effect.Amount.DynamicKind == compiler.DynamicAmountNone &&
		effect.Amount.Value >= 1
}

// isSourceStunRider reports whether effect is exactly the self-source stun "This
// <permanent> doesn't untap during your next untap step." rider that the dual
// lands (Mogg Hollows, Rootwater Depths, and the rest of the cycle) append to a
// mana ability so the land stays tapped through its controller's next untap
// step. It accepts only the parser-exact negated next-untap-step clause whose
// subject is the source itself, with no target, duration, or delayed timing, so
// no other negated-untap clause can ride into a mana ability. The mana ability's
// lowered content already carries the matching source SkipNextUntap instruction.
func isSourceStunRider(effect *compiler.CompiledEffect) bool {
	return effect.Kind == compiler.EffectUntap &&
		effect.Negated &&
		effect.Exact &&
		!effect.Optional &&
		effect.Context == parser.EffectContextSource &&
		len(effect.Targets) == 0 &&
		effect.Duration == compiler.DurationNone &&
		effect.DelayedTiming == 0
}

// isManaSpendRider reports whether effect is one of the closed, fully modeled
// mana-spend rider shapes.
func isManaSpendRider(effect *compiler.CompiledEffect) bool {
	if effect.Kind != compiler.EffectManaSpendRider ||
		!effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.HasUnrecognizedSibling ||
		effect.ManaSpendRider == nil {
		return false
	}
	return isCommanderScryManaSpendRider(effect.ManaSpendRider) ||
		isChosenTypeManaSpendRider(effect.ManaSpendRider) ||
		isChosenTypeCastOrActivateManaSpendRider(effect.ManaSpendRider) ||
		isLegendarySpellManaSpendRider(effect.ManaSpendRider) ||
		isCreatureSpellHasteManaSpendRider(effect.ManaSpendRider)
}

func isCommanderScryManaSpendRider(rider *compiler.CompiledManaSpendRider) bool {
	return rider.Condition == parser.ManaSpendCastCommanderCreatureType &&
		rider.Effect == parser.ManaSpendRiderEffectScry &&
		!rider.Restricted &&
		rider.ScryAmount >= 1
}

// isChosenTypeManaSpendRider reports whether rider is the restricted "spend this
// mana only to cast a creature spell of the chosen type" rider, either bare
// (Unclaimed Territory, Pillar of Origins) or with the optional "and that spell
// can't be countered" effect (Cavern of Souls).
func isChosenTypeManaSpendRider(rider *compiler.CompiledManaSpendRider) bool {
	return rider.Condition == parser.ManaSpendCastChosenCreatureType &&
		(rider.Effect == parser.ManaSpendRiderEffectCantBeCountered ||
			rider.Effect == parser.ManaSpendRiderEffectUnknown) &&
		rider.Restricted &&
		rider.ScryAmount == 0
}

// isChosenTypeCastOrActivateManaSpendRider reports whether rider is the
// restricted "spend this mana only to cast a creature spell of the chosen type
// or activate an ability of a creature source of the chosen type" rider
// (Secluded Courtyard).
func isChosenTypeCastOrActivateManaSpendRider(rider *compiler.CompiledManaSpendRider) bool {
	return rider.Condition == parser.ManaSpendCastOrActivateChosenCreatureType &&
		rider.Effect == parser.ManaSpendRiderEffectUnknown &&
		rider.Restricted &&
		rider.ScryAmount == 0
}

// isLegendarySpellManaSpendRider reports whether rider is the restricted
// "spend this mana only to cast a legendary spell" rider, with or without the
// optional "and that spell can't be countered" effect (Delighted Halfling).
func isLegendarySpellManaSpendRider(rider *compiler.CompiledManaSpendRider) bool {
	return rider.Condition == parser.ManaSpendCastLegendarySpell &&
		(rider.Effect == parser.ManaSpendRiderEffectCantBeCountered ||
			rider.Effect == parser.ManaSpendRiderEffectUnknown) &&
		rider.Restricted &&
		rider.ScryAmount == 0
}

// isCreatureSpellHasteManaSpendRider reports whether rider is the unrestricted
// "if that mana is spent on a creature spell, it gains haste until end of turn"
// bonus rider (Arena of Glory, Generator Servant). It is unrestricted: the
// tagged mana may pay for anything, but a creature spell paid with it gains
// haste through end of turn.
func isCreatureSpellHasteManaSpendRider(rider *compiler.CompiledManaSpendRider) bool {
	return rider.Condition == parser.ManaSpendCastCreatureSpell &&
		rider.Effect == parser.ManaSpendRiderEffectGainsHasteUntilEndOfTurn &&
		!rider.Restricted &&
		rider.ScryAmount == 0
}

// lowerManaSpendRiderContent lowers a typed add-mana effect and its exact
// spend-linked rider without consulting source text.
func lowerManaSpendRiderContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	manaEffect := ctx.content.Effects[0]
	if ctx.optional ||
		manaEffect.Negated ||
		manaEffect.DelayedTiming != 0 ||
		manaEffect.Duration != compiler.DurationNone ||
		manaEffect.HasUnrecognizedSibling {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana effect",
			"the executable source backend supports mana-spend riders only on exact modeled add-mana effects",
		)
	}
	riderEffect := ctx.content.Effects[1].ManaSpendRider
	if isCommanderScryManaSpendRider(riderEffect) {
		if !manaEffect.Mana.CommanderIdentity {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported mana effect",
				"the commander-creature-type rider requires an exact commander-identity add-mana effect",
			)
		}
		rider := game.ManaSpendRider{
			Condition: game.ManaSpendCastCommanderCreatureType,
			Effect: game.Mode{Sequence: []game.Instruction{
				{
					Primitive: game.Scry{
						Amount: game.Fixed(riderEffect.ScryAmount),
						Player: game.ControllerReference(),
					},
				},
			}},
		}
		return game.TapManaCommanderIdentityWithSpendRiderAbility(ctx.text, rider).Content, nil
	}
	if isChosenTypeManaSpendRider(riderEffect) {
		if !manaEffect.Mana.AnyColor {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported mana effect",
				"the chosen-creature-type rider requires an exact any-color add-mana effect",
			)
		}
		rider := game.ManaSpendRider{
			Condition:         game.ManaSpendCastChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		}
		if riderEffect.Effect == parser.ManaSpendRiderEffectCantBeCountered {
			rider.SpellRuleEffect = game.RuleEffectCantBeCountered
		}
		return game.TapManaChoiceWithSpendRiderAbility(
			ctx.text,
			rider,
			mana.W, mana.U, mana.B, mana.R, mana.G,
		).Content, nil
	}
	if isChosenTypeCastOrActivateManaSpendRider(riderEffect) {
		if !manaEffect.Mana.AnyColor {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported mana effect",
				"the chosen-creature-type cast-or-activate rider requires an exact any-color add-mana effect",
			)
		}
		rider := game.ManaSpendRider{
			Condition:         game.ManaSpendCastOrActivateChosenCreatureType,
			Restriction:       game.ManaSpendRestrictedToCondition,
			ChosenSubtypeFrom: game.EntryTypeChoiceKey,
		}
		return game.TapManaChoiceWithSpendRiderAbility(
			ctx.text,
			rider,
			mana.W, mana.U, mana.B, mana.R, mana.G,
		).Content, nil
	}
	if isLegendarySpellManaSpendRider(riderEffect) {
		if !manaEffect.Mana.AnyColor {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported mana effect",
				"the legendary-spell rider requires an exact any-color add-mana effect",
			)
		}
		rider := game.ManaSpendRider{
			Condition:   game.ManaSpendCastLegendarySpell,
			Restriction: game.ManaSpendRestrictedToCondition,
		}
		if riderEffect.Effect == parser.ManaSpendRiderEffectCantBeCountered {
			rider.SpellRuleEffect = game.RuleEffectCantBeCountered
		}
		return game.TapManaChoiceWithSpendRiderAbility(
			ctx.text,
			rider,
			mana.W, mana.U, mana.B, mana.R, mana.G,
		).Content, nil
	}
	if isCreatureSpellHasteManaSpendRider(riderEffect) {
		content, ok := typedManaEffectContent(manaEffect.Mana)
		if !ok {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported mana effect",
				"the creature-spell haste rider requires an exact modeled add-mana effect",
			)
		}
		rider := game.ManaSpendRider{
			Condition:          game.ManaSpendCastCreatureSpell,
			SpellGainsKeywords: []game.Keyword{game.Haste},
		}
		if !attachManaSpendRider(&content, rider) {
			return game.AbilityContent{}, contentDiagnostic(
				ctx,
				"unsupported mana effect",
				"the creature-spell haste rider requires an exact add-mana instruction to tag",
			)
		}
		return content, nil
	}
	return game.AbilityContent{}, contentDiagnostic(
		ctx,
		"unsupported mana effect",
		"the executable source backend cannot lower this mana-spend rider",
	)
}

// attachManaSpendRider tags every add-mana instruction in content's single mode
// with rider so each produced unit carries the spend-linked semantics. It
// reports false (fail closed) when content is not a single mode whose entire
// sequence is add-mana instructions, so a rider cannot ride onto content it
// cannot fully tag.
func attachManaSpendRider(content *game.AbilityContent, rider game.ManaSpendRider) bool {
	if len(content.Modes) != 1 {
		return false
	}
	seq := content.Modes[0].Sequence
	if len(seq) == 0 {
		return false
	}
	for i := range seq {
		add, ok := seq[i].Primitive.(game.AddMana)
		if !ok {
			return false
		}
		add.SpendRider = opt.Val(rider)
		seq[i].Primitive = add
	}
	return true
}

func lowerAddManaContent(ctx contentCtx) (game.AbilityContent, *shared.Diagnostic) {
	effect := ctx.content.Effects[0]
	if content, ok := lowerTriggerLandProducedMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerTargetOpponentHandMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerControlledCountMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerSourceCounterCountMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerChosenColorCountMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerChosenColorSourceCounterMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerAnyOneColorDynamicMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerReferencedControllerAddMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerReferencedPlayerAddMana(ctx); ok {
		return content, nil
	}
	if content, ok := lowerEachColorAmongControlledMana(ctx); ok {
		return content, nil
	}
	if !effect.Mana.LegacyBodyExact && (effect.Mana.AnyColor || effect.Mana.CommanderIdentity || effect.Mana.LandsProduce || effect.Mana.ColorsAmongControlled || len(effect.Mana.Symbols) != 0) {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana symbol",
			"the executable source backend cannot lower this add-mana content",
		)
	}
	if ctx.optional ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		ctx.content.Unconsumed() {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana effect",
			"the executable source backend supports only exact unconditional add-mana content",
		)
	}
	content, ok := typedManaEffectContent(effect.Mana)
	if !ok {
		return game.AbilityContent{}, contentDiagnostic(
			ctx,
			"unsupported mana symbol",
			"the executable source backend cannot lower this add-mana content",
		)
	}
	return content, nil
}

func lowerTargetOpponentHandMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 1 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	selector := effect.Amount.Selector()
	selection, characteristicsOK := dynamicCountCharacteristics(selector)
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.Amount.DynamicKind != compiler.DynamicAmountCount ||
		effect.Amount.DynamicForm != compiler.DynamicAmountForEach ||
		effect.Amount.Multiplier != 1 ||
		selector.Kind != compiler.SelectorCard ||
		selector.Controller != compiler.ControllerOpponent ||
		selector.Zone != zone.Hand ||
		!characteristicsOK ||
		!selection.Empty() ||
		!effect.Mana.ColorsKnown ||
		len(effect.Mana.Colors) != 1 ||
		effect.Mana.Colors[0] != mana.R ||
		effect.Mana.Choice ||
		effect.Mana.AnyColor {
		return game.AbilityContent{}, false
	}
	target, ok := playerTargetSpec(ctx.content.Targets[0])
	if !ok || target.Predicate.Player != game.PlayerOpponent {
		return game.AbilityContent{}, false
	}
	player := game.TargetPlayerReference(0)
	return game.Mode{
		Targets: []game.TargetSpec{target},
		Sequence: []game.Instruction{{Primitive: game.AddMana{
			Amount: game.Dynamic(game.DynamicAmount{
				Kind:       game.DynamicAmountCountCardsInZone,
				Multiplier: 1,
				Player:     &player,
				CardZone:   zone.Hand,
				Selection:  &game.Selection{},
			}),
			ManaColor: mana.R,
		}}},
	}.Ability(), true
}

// lowerControlledCountMana lowers an "Add <mana> for each <permanent> you
// control" body (Cabal Coffers, Gaea's Cradle, Serra's Sanctum) into an AddMana
// instruction whose amount is a dynamic battlefield permanent count. It accepts
// only a single fixed produced color scaled by a recognized battlefield count
// selector; choice, any-color, and non-battlefield counts fail closed so an
// unmodeled wording cannot lower to a mislabeled ability.
func lowerControlledCountMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.Amount.DynamicKind != compiler.DynamicAmountCount ||
		effect.Amount.DynamicForm != compiler.DynamicAmountForEach ||
		!effect.Mana.ColorsKnown ||
		len(effect.Mana.Colors) != 1 ||
		effect.Mana.Choice ||
		effect.Mana.AnyColor {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
	if !ok || dynamic.Kind != game.DynamicAmountCountSelector {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: game.AddMana{
			Amount:    game.Dynamic(dynamic),
			ManaColor: effect.Mana.Colors[0],
		}}},
	}.Ability(), true
}

// lowerEachColorAmongControlledMana lowers a "For each color among <permanents>
// you control, add one mana of that color" body (Bloom Tender) into an AddMana
// instruction that produces one mana of each color among the controller's
// matching permanents. It reuses the among-controlled permanent filter, which
// accepts a bare "permanents you control" group as well as a narrowed one, and
// fails closed on any filter facet the executable backend cannot represent.
func lowerEachColorAmongControlledMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		!effect.Mana.EachColorAmongControlled ||
		effect.Mana.ColorsAmongSelector == nil {
		return game.AbilityContent{}, false
	}
	selection, ok := colorsAmongControlledSelection(*effect.Mana.ColorsAmongSelector)
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.TapManaEachControlledColorAbility("", selection).Content, true
}

// <this permanent>" body (Everflowing Chalice) into an AddMana instruction whose
// amount is the number of counters of one kind on the source permanent. It
// accepts only a single fixed produced color scaled by a recognized counter
// kind, and tolerates the lone self reference naming the source permanent ("this
// artifact"); choice, any-color, and other reference shapes fail closed so an
// unmodeled wording cannot lower to a mislabeled ability.
func lowerSourceCounterCountMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!singleSelfReference(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourceCounterCount ||
		effect.Amount.DynamicForm != compiler.DynamicAmountForEach ||
		!effect.Mana.ColorsKnown ||
		len(effect.Mana.Colors) != 1 ||
		effect.Mana.Choice ||
		effect.Mana.AnyColor {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
	if !ok || dynamic.Kind != game.DynamicAmountObjectCounters {
		return game.AbilityContent{}, false
	}
	return game.Mode{
		Sequence: []game.Instruction{{Primitive: game.AddMana{
			Amount:    game.Dynamic(dynamic),
			ManaColor: effect.Mana.Colors[0],
		}}},
	}.Ability(), true
}

// lowerChosenColorCountMana lowers a "Choose a color. Add an amount of mana of
// that color equal to <dynamic count>" body (Three Tree City) into a Choose plus
// an AddMana instruction whose color is the chosen color and whose amount is a
// dynamic battlefield permanent count. It accepts only a single "equal to"
// battlefield count of multiplier one; choice, fixed-color, multiplied, and
// non-battlefield amounts fail closed so an unmodeled wording cannot lower to a
// mislabeled ability.
func lowerChosenColorCountMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 0 {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		!effect.Mana.ChosenColorDynamic ||
		effect.Amount.DynamicKind != compiler.DynamicAmountCount ||
		effect.Amount.DynamicForm != compiler.DynamicAmountEqual ||
		effect.Amount.Multiplier != 1 {
		return game.AbilityContent{}, false
	}
	selection, ok := dynamicAmountSelection(effect.Amount.Selector())
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.TapManaChosenColorCountAbility("", selection).Content, true
}

// lowerChosenColorSourceCounterMana lowers a "Choose a color. Add one mana of
// that color for each <kind> counter on <this permanent>" body (Astral
// Cornucopia) into a Choose-a-color plus an AddMana instruction whose color is
// the chosen color and whose amount is the number of counters of one kind on the
// source permanent. It accepts only a single source-counter ForEach amount that
// lowers to an object-counter dynamic; other amounts fail closed so an unmodeled
// wording cannot lower to a mislabeled ability.
func lowerChosenColorSourceCounterMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		!singleSelfReference(ctx.content.References) {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		!effect.Mana.ChosenColorDynamic ||
		effect.Amount.DynamicKind != compiler.DynamicAmountSourceCounterCount ||
		effect.Amount.DynamicForm != compiler.DynamicAmountForEach {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
	if !ok || dynamic.Kind != game.DynamicAmountObjectCounters {
		return game.AbilityContent{}, false
	}
	return game.TapManaChosenColorDynamicAbility("", dynamic).Content, true
}

// lowerAnyOneColorDynamicMana lowers an "Add X mana of any one color, where X is
// <dynamic amount>" body (Kami of Whispered Hopes: "...this creature's power.")
// into a Choose-a-color plus an AddMana instruction whose color is the chosen
// color and whose amount is the dynamic value. The dynamic amount is lowered
// generically (source power/toughness, devotion, a permanent count, and so on),
// so any amount lowerDynamicAmount supports unlocks this mana ability; an
// unsupported amount fails closed.
func lowerAnyOneColorDynamicMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		(len(ctx.content.References) != 0 && !singleSelfReference(ctx.content.References)) {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if !effect.Exact ||
		effect.Negated ||
		effect.Optional ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone ||
		effect.Context != parser.EffectContextController ||
		!effect.Mana.AnyOneColorDynamic {
		return game.AbilityContent{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, game.SourcePermanentReference())
	if !ok {
		return game.AbilityContent{}, false
	}
	return game.TapManaChosenColorDynamicAbility("", dynamic).Content, true
}

// lowerTriggerLandProducedMana lowers the mana-doubler body of a tapped-for-mana
// trigger that adds "one mana of any type that land produced" (Mirari's Wake,
// Zendikar Resurgent). The "that land" pronoun binds to the triggering permanent
// (a single event-permanent reference), but the produced mana is read at
// resolution from the triggering tap's recorded production rather than from the
// referenced object, so the reference is consumed here. The mana goes to the
// ability's controller ("add ...", EffectContextController).
func lowerTriggerLandProducedMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 {
		return game.AbilityContent{}, false
	}
	for _, reference := range ctx.content.References {
		if reference.Binding != compiler.ReferenceBindingEventPermanent {
			return game.AbilityContent{}, false
		}
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectAddMana ||
		!effect.Mana.TriggerLandProducedType ||
		effect.Context != parser.EffectContextController ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone {
		return game.AbilityContent{}, false
	}
	return game.TriggerLandProducedManaContent(), true
}

// lowerReferencedControllerAddMana lowers a triggered-ability body that adds
// fixed mana to the controller of the permanent that fired the trigger ("its
// controller adds an additional {G}", Wild Growth and the mana-additional aura
// family). The "its" pronoun binds to the triggering permanent, so the produced
// mana is routed to ObjectControllerReference(EventPermanentReference()) rather
// than the ability's controller. Exact fixed-color output is supported, as well
// as "one mana of the chosen color" routed through the source permanent's
// entry-time color choice (Utopia Sprawl: "Whenever enchanted Forest is tapped
// for mana, its controller adds an additional one mana of the chosen color.").
func lowerReferencedControllerAddMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 1 ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingEventPermanent {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectAddMana ||
		effect.Context != parser.EffectContextReferencedObjectController ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone {
		return game.AbilityContent{}, false
	}
	manaEffect := effect.Mana
	recipient := game.ObjectControllerReference(game.EventPermanentReference())
	if manaEffect.ChosenColor &&
		!manaEffect.ChosenColorFixedKnown &&
		!manaEffect.ChosenColorDevotion &&
		!manaEffect.ChosenColorDynamic {
		return game.Mode{Sequence: []game.Instruction{{Primitive: game.AddMana{
			Amount:          game.Fixed(1),
			EntryChoiceFrom: game.EntryColorChoiceKey,
			Player:          opt.Val(recipient),
		}}}}.Ability(), true
	}
	if manaEffect.AnyColor ||
		manaEffect.CommanderIdentity ||
		manaEffect.LandsProduce ||
		manaEffect.Choice ||
		manaEffect.FilterPair ||
		manaEffect.ChosenColor ||
		manaEffect.LinkedExileColors ||
		!manaEffect.ColorsKnown ||
		len(manaEffect.Colors) == 0 {
		return game.AbilityContent{}, false
	}
	recipient = game.ObjectControllerReference(game.EventPermanentReference())
	seq := make([]game.Instruction, 0, len(manaEffect.Colors))
	for _, c := range manaEffect.Colors {
		seq = append(seq, game.Instruction{Primitive: game.AddMana{
			Amount:    game.Fixed(1),
			ManaColor: c,
			Player:    opt.Val(recipient),
		}})
	}
	return game.Mode{Sequence: seq}.Ability(), true
}

// lowerReferencedPlayerAddMana lowers a "that player adds <mana>" add-mana body
// whose recipient is the player named by the triggering event rather than the
// ability's controller, as in the generic "Whenever a player taps a land for
// mana, that player adds an additional {U}" tapped-for-mana trigger (High Tide,
// Bubbling Muck). The lone "that player" reference binds to the triggering
// event's player (ReferenceBindingEventPlayer), which the runtime resolves to
// the controller of the tapped permanent. It accepts only fixed known produced
// colors; dynamic, any-color, produced-type, and choice mana shapes fail closed.
func lowerReferencedPlayerAddMana(ctx contentCtx) (game.AbilityContent, bool) {
	if ctx.optional ||
		len(ctx.content.Effects) != 1 ||
		len(ctx.content.Targets) != 0 ||
		len(ctx.content.Conditions) != 0 ||
		len(ctx.content.Keywords) != 0 ||
		len(ctx.content.Modes) != 0 ||
		len(ctx.content.References) != 1 ||
		ctx.content.References[0].Kind != compiler.ReferenceThatPlayer ||
		ctx.content.References[0].Binding != compiler.ReferenceBindingEventPlayer {
		return game.AbilityContent{}, false
	}
	effect := ctx.content.Effects[0]
	if effect.Kind != compiler.EffectAddMana ||
		effect.Context != parser.EffectContextReferencedPlayer ||
		effect.Negated ||
		effect.DelayedTiming != 0 ||
		effect.Duration != compiler.DurationNone {
		return game.AbilityContent{}, false
	}
	manaEffect := effect.Mana
	if manaEffect.AnyColor ||
		manaEffect.CommanderIdentity ||
		manaEffect.LandsProduce ||
		manaEffect.Choice ||
		manaEffect.FilterPair ||
		manaEffect.ChosenColor ||
		manaEffect.LinkedExileColors ||
		!manaEffect.ColorsKnown ||
		len(manaEffect.Colors) == 0 {
		return game.AbilityContent{}, false
	}
	recipient := game.EventPlayerReference()
	seq := make([]game.Instruction, 0, len(manaEffect.Colors))
	for _, c := range manaEffect.Colors {
		seq = append(seq, game.Instruction{Primitive: game.AddMana{
			Amount:    game.Fixed(1),
			ManaColor: c,
			Player:    opt.Val(recipient),
		}})
	}
	return game.Mode{Sequence: seq}.Ability(), true
}

func typedManaEffectContent(effect compiler.CompiledEffectMana) (game.AbilityContent, bool) {
	if effect.TriggerLandProducedType {
		return game.TriggerLandProducedManaContent(), true
	}
	if effect.FilterPair {
		if len(effect.FilterColors) != 2 {
			return game.AbilityContent{}, false
		}
		return game.TwoColorFilterManaAbility(effect.FilterColors[0], effect.FilterColors[1]).Content, true
	}
	if effect.ChosenColor {
		if effect.ChosenColorFixedKnown {
			return game.TapFixedOrChosenColorManaAbility("", effect.ChosenColorFixed).Content, true
		}
		return game.TapChosenColorManaAbility("").Content, true
	}
	if effect.ChosenColorDevotion {
		return game.TapManaChosenColorDevotionAbility("").Content, true
	}
	if effect.CommanderIdentity {
		return game.TapManaCommanderIdentityAbility().Content, true
	}
	if effect.LandsProduce {
		relation, ok := landsProduceRelation(effect.LandsProduceScope)
		if !ok {
			return game.AbilityContent{}, false
		}
		return game.TapManaLandsProduceAbility(relation, effect.LandsProduceAnyType).Content, true
	}
	if effect.LinkedExileColors {
		return game.TapLinkedExileColorManaAbility(imprintLinkKey).Content, true
	}
	if effect.ColorsAmongControlled {
		if effect.ColorsAmongSelector == nil {
			return game.AbilityContent{}, false
		}
		selection, ok := colorsAmongControlledSelection(*effect.ColorsAmongSelector)
		if !ok {
			return game.AbilityContent{}, false
		}
		return game.TapManaAmongControlledColorsAbility("", selection).Content, true
	}
	if effect.AnyColor {
		if effect.AnyColorCount >= 2 {
			return game.TapManaChoiceCountAbility("", effect.AnyColorCount, mana.W, mana.U, mana.B, mana.R, mana.G).Content, true
		}
		return game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G).Content, true
	}
	if !effect.ColorsKnown {
		return game.AbilityContent{}, false
	}
	colors := effect.Colors
	if effect.Choice && len(colors) >= 2 {
		return game.TapManaChoiceAbility(colors...).Content, true
	}
	if !effect.Choice && len(colors) > 0 {
		return manaFixedContent(colors), true
	}
	return game.AbilityContent{}, false
}

// DO-NOT-COPY(filter): routes a single typed Kind through the disjunctive
// RequiredTypesAny union (not the canonical RequiredTypes), a representation the
// canonical projector does not reproduce; prefer SelectionForSelectorMasked for
// new code. (retire: #1393)
//
// colorsAmongControlledSelection builds the runtime permanent filter for a "one
// mana of any color among <permanents> you control" mana ability from its
// compiled selector. It accepts only a "you control" battlefield group and maps
// the selector's type union (or its single typed Kind), supertypes, subtypes,
// and color filters onto a Selection. It fails closed on any selector facet the
// executable backend cannot represent exactly (a foreign controller, a non-
// permanent Kind, excluded types/supertypes/colors, keyword, combat, tapped, or
// numeric qualifiers) so an unmodeled wording cannot lower to a mislabeled
// ability.
func colorsAmongControlledSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.Controller != compiler.ControllerYou ||
		selector.All || selector.Another || selector.Other ||
		selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.ExcludedKeyword != parser.KeywordUnknown ||
		selector.Zone != zone.None ||
		len(selector.ExcludedTypes()) != 0 ||
		len(selector.ExcludedSupertypes()) != 0 ||
		len(selector.ExcludedColors()) != 0 {
		return game.Selection{}, false
	}
	requiredTypes, ok := colorsAmongRequiredTypes(selector)
	if !ok {
		return game.Selection{}, false
	}
	selection := game.Selection{
		Controller:       game.ControllerYou,
		RequiredTypesAny: requiredTypes,
		Colorless:        selector.Colorless,
		Multicolored:     selector.Multicolored,
	}
	if supertypes := selector.Supertypes(); len(supertypes) > 0 {
		selection.Supertypes = append([]types.Super(nil), supertypes...)
	}
	if subtypes := selector.SubtypesAny(); len(subtypes) > 0 {
		selection.SubtypesAny = append([]types.Sub(nil), subtypes...)
	}
	if colors := selector.ColorsAny(); len(colors) > 0 {
		selection.ColorsAny = append([]color.Color(nil), colors...)
	}
	return selection, true
}

// colorsAmongRequiredTypes resolves the card-type filter of an among-controlled
// mana selector. A disjunctive type union ("creatures and planeswalkers") is
// carried verbatim; a single typed Kind ("creatures") contributes its one type;
// the catch-all permanent and any Kinds carry no type filter (every permanent
// the controller controls qualifies). It fails closed on a Kind that is not a
// permanent characteristic.
func colorsAmongRequiredTypes(selector compiler.CompiledSelector) ([]types.Card, bool) {
	if union := selector.RequiredTypesAny(); len(union) > 0 {
		return append([]types.Card(nil), union...), true
	}
	switch selector.Kind {
	case compiler.SelectorAny, compiler.SelectorPermanent:
		return nil, true
	case compiler.SelectorArtifact:
		return []types.Card{types.Artifact}, true
	case compiler.SelectorCreature:
		return []types.Card{types.Creature}, true
	case compiler.SelectorEnchantment:
		return []types.Card{types.Enchantment}, true
	case compiler.SelectorLand:
		return []types.Card{types.Land}, true
	case compiler.SelectorPlaneswalker:
		return []types.Card{types.Planeswalker}, true
	default:
		return nil, false
	}
}

// manaFixedContent builds AbilityContent that adds one mana of each color in
// the given order. For a single color this produces a single AddMana
// instruction identical to game.TapManaAbility.
func manaFixedContent(colors []mana.Color) game.AbilityContent {
	seq := make([]game.Instruction, 0, len(colors))
	for _, c := range colors {
		seq = append(seq, game.Instruction{
			Primitive: game.AddMana{
				Amount:    game.Fixed(1),
				ManaColor: c,
			},
		})
	}
	return game.Mode{Sequence: seq}.Ability()
}

// landsProduceRelation maps a parsed lands-produce scope to the game player
// relation its mana ability scopes to. It fails closed on any unrecognized
// scope so an unmodeled wording cannot lower to a mislabeled ability.
func landsProduceRelation(scope parser.ManaLandsProduceScope) (game.PlayerRelation, bool) {
	switch scope {
	case parser.ManaLandsProduceYou:
		return game.PlayerYou, true
	case parser.ManaLandsProduceOpponent:
		return game.PlayerOpponent, true
	default:
		return game.PlayerAny, false
	}
}

func choiceTapManaAbility(colorNames []string) game.ManaAbility {
	colors := make([]mana.Color, 0, len(colorNames))
	for _, name := range colorNames {
		if manaColor, ok := manaColorValue(name); ok {
			colors = append(colors, manaColor)
		}
	}
	return game.TapManaChoiceAbility(colors...)
}

func manaCostHasVariableSymbol(manaCost cost.Mana) bool {
	for _, symbol := range manaCost {
		if symbol.Kind == cost.VariableSymbol {
			return true
		}
	}
	return false
}

func manaColorValue(name string) (mana.Color, bool) {
	switch name {
	case "W":
		return mana.W, true
	case "U":
		return mana.U, true
	case "B":
		return mana.B, true
	case "R":
		return mana.R, true
	case "G":
		return mana.G, true
	case "C":
		return mana.C, true
	default:
		return "", false
	}
}

// parseManaCostValue parses a Scryfall mana cost string (e.g., "{2}{W}") into a
// typed cost.Mana value. Empty input yields a nil cost.
func parseManaCostValue(s string) (cost.Mana, error) {
	if s == "" {
		return nil, nil
	}
	matches := manaSymbolRe.FindAllStringSubmatch(s, -1)
	if len(matches) == 0 {
		return nil, nil
	}
	out := make(cost.Mana, 0, len(matches))
	for _, match := range matches {
		symbol, err := parseManaSymbolValue(match[1])
		if err != nil {
			return nil, fmt.Errorf("unsupported mana symbol {%s} in cost %q: %w", match[1], s, err)
		}
		out = append(out, symbol)
	}
	return out, nil
}

func parseManaSymbolValue(sym string) (cost.Symbol, error) {
	switch sym {
	case "X":
		return cost.X, nil
	case "C":
		return cost.C, nil
	case "S":
		return cost.S, nil
	case "W":
		return cost.W, nil
	case "U":
		return cost.U, nil
	case "B":
		return cost.B, nil
	case "R":
		return cost.R, nil
	case "G":
		return cost.G, nil
	default:
	}
	if before, ok := strings.CutSuffix(sym, "/P"); ok {
		manaColor, ok := manaColorValue(before)
		if !ok {
			return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
		}
		return cost.PhyrexianMana(manaColor), nil
	}
	if strings.Contains(sym, "/") {
		parts := strings.SplitN(sym, "/", 2)
		if _, err := strconv.Atoi(parts[0]); err == nil {
			manaColor, ok := manaColorValue(parts[1])
			if !ok {
				return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
			}
			return cost.Twobrid(manaColor), nil
		}
		first, ok := manaColorValue(parts[0])
		second, ok2 := manaColorValue(parts[1])
		if !ok || !ok2 {
			return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
		}
		return cost.HybridMana(first, second), nil
	}
	if n, err := strconv.Atoi(sym); err == nil {
		return cost.O(n), nil
	}
	return cost.Symbol{}, fmt.Errorf("unsupported mana symbol: %s", sym)
}
