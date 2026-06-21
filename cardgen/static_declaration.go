package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerStaticDeclarations is the only semantic Static Declaration to runtime
// static-value lowering path.
func lowerStaticDeclarations(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	if ability.Kind != compiler.AbilityStatic || ability.Static == nil || len(ability.Static.Declarations) == 0 {
		return abilityLowering{}, false, nil
	}
	if ability.Static.Blocker != compiler.StaticDeclarationBlockerNone {
		return abilityLowering{}, true, lowerStaticDeclarationBlocker(ability)
	}
	declarations := ability.Static.Declarations

	if lowered, ok, diagnostic := lowerStaticCharacteristicPowerToughness(ability, syntax); ok {
		return lowered, true, diagnostic
	}

	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		!rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return abilityLowering{}, true, staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration shell",
			"the recognized static declarations require an otherwise empty static ability shell",
		)
	}
	body := game.StaticAbility{Text: ability.Text}
	varName := ""
	conditionSpan := shared.Span{}
	hasCondition := declarations[0].Condition != nil
	for _, declaration := range declarations {
		if (declaration.Condition != nil) != hasCondition ||
			(declaration.Condition != nil && conditionSpan != (shared.Span{}) && declaration.Condition.Span != conditionSpan) {
			return abilityLowering{}, true, staticDeclarationDiagnostic(
				ability,
				"unsupported static declaration condition",
				"all declarations in one static ability must have the same supported condition",
			)
		}
		if declaration.Condition != nil && conditionSpan == (shared.Span{}) {
			if declaration.Condition.SourceInGraveyard {
				body.ZoneOfFunction = zone.Graveyard
			}
			condition, ok := lowerCondition(*declaration.Condition, conditionContextStatic)
			if !ok {
				return abilityLowering{}, true, staticDeclarationDiagnostic(
					ability,
					"unsupported static declaration condition",
					"the recognized static declaration condition is not representable in a static runtime ability",
				)
			}
			body.Condition = opt.Val(condition)
			conditionSpan = declaration.Condition.Span
		}
		var ok bool
		if !staticDeclarationPayloadValid(declaration) {
			ok = false
		} else {
			switch declaration.Kind {
			case compiler.StaticDeclarationContinuous:
				ok = appendStaticContinuousDeclaration(&body, declaration)
			case compiler.StaticDeclarationRule:
				ok = appendStaticRuleDeclaration(&body, declaration)
			case compiler.StaticDeclarationCostModifier:
				ok = appendStaticCostModifierDeclaration(&body, declaration)
			case compiler.StaticDeclarationCardAbilityGrant:
				ok = appendStaticCardAbilityGrantDeclaration(&body, declaration)
				if ok {
					body.Text = declaration.CardGrant.Text
				}
			case compiler.StaticDeclarationPlayerRule:
				ok = appendStaticPlayerRuleDeclaration(&body, declaration)
			case compiler.StaticDeclarationOpponentActionRestriction:
				ok = appendStaticOpponentActionRestrictionDeclaration(&body, declaration)
			case compiler.StaticDeclarationEnterBattlefieldRestriction:
				ok = appendStaticEnterBattlefieldRestrictionDeclaration(&body, declaration)
			case compiler.StaticDeclarationSpellUncounterable:
				ok = appendStaticSpellUncounterableDeclaration(&body, declaration)
			case compiler.StaticDeclarationEnteringTriggerMultiplier:
				ok = appendStaticEnteringTriggerMultiplierDeclaration(&body, declaration)
			case compiler.StaticDeclarationUntapStep:
				ok = appendStaticUntapStepDeclaration(&body, declaration)
			case compiler.StaticDeclarationCastAsThoughFlash:
				ok = appendStaticCastAsThoughFlashDeclaration(&body, declaration)
			default:
				ok = false
			}
		}
		if !ok {
			detail := "the recognized static declaration operation is not representable by the runtime static-value vocabulary"
			if declaration.Kind == compiler.StaticDeclarationCardAbilityGrant || strings.Contains(ability.Text, `have "`) {
				detail = "the static declaration operation or its exact syntax is not representable"
			}
			return abilityLowering{}, true, staticDeclarationDiagnostic(
				ability,
				"unsupported static declaration operation",
				detail,
			)
		}
	}
	if len(declarations) == 1 {
		varName = canonicalStaticDeclarationVarName(declarations[0])
	}
	spans := make([]shared.Span, 0, len(declarations)+len(syntax.Reminders))
	for _, declaration := range declarations {
		spans = append(spans, declaration.Span)
	}
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		staticAbilities: []loweredStaticAbility{{
			Body:    body,
			VarName: varName,
		}},
		consumed: semanticConsumption{
			conditions:   len(ability.Content.Conditions),
			effects:      len(ability.Content.Effects),
			keywords:     len(ability.Content.Keywords),
			references:   len(ability.Content.References),
			declarations: len(declarations),
		},
		sourceSpans: spans,
	}, true, nil
}

// lowerStaticCharacteristicPowerToughness lowers a characteristic-defining
// power/toughness declaration ("<source>'s power and toughness are each equal to
// <count>") into a face-level dynamic power and toughness. The declaration sets
// the source object's printed characteristic, so it produces no runtime static
// ability; the printed power and toughness are the `*` placeholders the runtime
// evaluates against the dynamic value.
func lowerStaticCharacteristicPowerToughness(
	ability compiler.CompiledAbility,
	syntax *parser.Ability,
) (abilityLowering, bool, *shared.Diagnostic) {
	declarations := ability.Static.Declarations
	if len(declarations) != 1 || declarations[0].Kind != compiler.StaticDeclarationCharacteristicPowerToughness {
		return abilityLowering{}, false, nil
	}
	declaration := declarations[0]
	if ability.Cost != nil ||
		ability.Trigger != nil ||
		len(ability.Content.Modes) != 0 ||
		len(ability.Content.Targets) != 0 ||
		!rulesFreeAbilityWordLabel(ability.AbilityWord) {
		return abilityLowering{}, true, staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration shell",
			"the recognized static declarations require an otherwise empty static ability shell",
		)
	}
	if !staticDeclarationPayloadValid(declaration) || declaration.Condition != nil {
		return abilityLowering{}, true, staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration operation",
			"the recognized static declaration operation is not representable by the runtime static-value vocabulary",
		)
	}
	value := game.DynamicValue{Kind: declaration.CharacteristicPT.Value}
	spans := make([]shared.Span, 0, 1+len(syntax.Reminders))
	spans = append(spans, declaration.Span)
	for _, reminder := range syntax.Reminders {
		spans = append(spans, reminder.Span)
	}
	return abilityLowering{
		dynamicPower:     opt.Val(value),
		dynamicToughness: opt.Val(value),
		consumed: semanticConsumption{
			conditions:   len(ability.Content.Conditions),
			effects:      len(ability.Content.Effects),
			keywords:     len(ability.Content.Keywords),
			references:   len(ability.Content.References),
			declarations: len(declarations),
		},
		sourceSpans: spans,
	}, true, nil
}

func lowerStaticDeclarationBlocker(ability compiler.CompiledAbility) *shared.Diagnostic {
	if ability.Static == nil {
		return nil
	}
	switch ability.Static.Blocker {
	case compiler.StaticDeclarationBlockerHistoricCardSelection:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration group",
			"historic card predicates are not supported by the executable source backend",
		)
	case compiler.StaticDeclarationBlockerCondition:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration condition",
			"the static declaration has an unsupported or ambiguously scoped condition",
		)
	case compiler.StaticDeclarationBlockerDuration:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration duration",
			"the static declaration has a duration that is not valid for a source-derived static value",
		)
	case compiler.StaticDeclarationBlockerGroup:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration group",
			"the static declaration affected group is unsupported or ambiguous",
		)
	case compiler.StaticDeclarationBlockerOperation:
		detail := "the static declaration operation or its exact syntax is not representable"
		if staticDeclarationHasUnknownTypedSubtypeSubject(ability) && len(ability.Content.Keywords) != 0 && !strings.Contains(ability.Text, `have "`) {
			detail = "the recognized static declaration operation is not representable by the runtime static-value vocabulary"
		}
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration operation",
			detail,
		)
	case compiler.StaticDeclarationBlockerShell:
		detail := "the static declaration shell carries unsupported additional semantics"
		if !rulesFreeAbilityWordLabel(ability.AbilityWord) && staticDeclarationHasUnknownTypedSubtypeSubject(ability) {
			detail = "the recognized static declarations require an otherwise empty static ability shell"
		}
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration shell",
			detail,
		)
	default:
		return nil
	}
}

func staticDeclarationHasUnknownTypedSubtypeSubject(ability compiler.CompiledAbility) bool {
	return slices.ContainsFunc(ability.Content.Effects, func(effect compiler.CompiledEffect) bool {
		switch effect.StaticSubject {
		case compiler.StaticSubjectControlledCreatureSubtype, compiler.StaticSubjectOtherControlledCreatureSubtype:
			if effect.StaticSubjectSub() == types.Equipment || strings.EqualFold(effect.StaticSubjectSubtype(), "Equipment") {
				return false
			}
			return !effect.StaticSubjectSubKnown()
		default:
			return false
		}
	})
}

func staticDeclarationPayloadValid(declaration compiler.StaticDeclaration) bool {
	payloads := 0
	if declaration.Continuous != nil {
		payloads++
	}
	if declaration.Rule != nil {
		payloads++
	}
	if declaration.Cost != nil {
		payloads++
	}
	if declaration.CardGrant != nil {
		payloads++
	}
	if declaration.Player != nil {
		payloads++
	}
	if declaration.OpponentRestriction != nil {
		payloads++
	}
	if declaration.EnterRestriction != nil {
		payloads++
	}
	if declaration.SpellUncounterable != nil {
		payloads++
	}
	if declaration.EnteringMultiplier != nil {
		payloads++
	}
	if declaration.Untap != nil {
		payloads++
	}
	if declaration.CharacteristicPT != nil {
		payloads++
	}
	if declaration.CastAsThoughFlash != nil {
		payloads++
	}
	if payloads != 1 {
		return false
	}
	switch declaration.Kind {
	case compiler.StaticDeclarationContinuous:
		return declaration.Continuous != nil
	case compiler.StaticDeclarationRule:
		return declaration.Rule != nil
	case compiler.StaticDeclarationCostModifier:
		return declaration.Cost != nil
	case compiler.StaticDeclarationCardAbilityGrant:
		return declaration.CardGrant != nil
	case compiler.StaticDeclarationPlayerRule:
		return declaration.Player != nil
	case compiler.StaticDeclarationOpponentActionRestriction:
		return declaration.OpponentRestriction != nil
	case compiler.StaticDeclarationEnterBattlefieldRestriction:
		return declaration.EnterRestriction != nil
	case compiler.StaticDeclarationSpellUncounterable:
		return declaration.SpellUncounterable != nil
	case compiler.StaticDeclarationEnteringTriggerMultiplier:
		return declaration.EnteringMultiplier != nil
	case compiler.StaticDeclarationUntapStep:
		return declaration.Untap != nil
	case compiler.StaticDeclarationCharacteristicPowerToughness:
		return declaration.CharacteristicPT != nil
	case compiler.StaticDeclarationCastAsThoughFlash:
		return declaration.CastAsThoughFlash != nil
	default:
		return false
	}
}

func appendStaticContinuousDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	effect, ok := lowerStaticContinuousDeclaration(declaration)
	if !ok {
		return false
	}
	body.ContinuousEffects = append(body.ContinuousEffects, effect)
	return true
}

func lowerStaticContinuousDeclaration(declaration compiler.StaticDeclaration) (game.ContinuousEffect, bool) {
	layer, ok := lowerStaticContinuousLayer(declaration.Continuous.Layer)
	if !ok {
		return game.ContinuousEffect{}, false
	}
	group, ok := lowerStaticGroupReference(declaration.Group)
	if !ok {
		return game.ContinuousEffect{}, false
	}
	effect := game.ContinuousEffect{
		Layer:          layer,
		AffectedSource: group.AffectedSource,
		Group:          group.Group,
	}
	switch declaration.Continuous.Operation {
	case compiler.StaticContinuousModifyPowerToughness:
		if layer != game.LayerPowerToughnessModify {
			return game.ContinuousEffect{}, false
		}
		effect.PowerDelta = compiledSignedAmountValue(declaration.Continuous.PowerDelta)
		effect.ToughnessDelta = compiledSignedAmountValue(declaration.Continuous.ToughnessDelta)
		if declaration.Continuous.DynamicAmount.DynamicKind != compiler.DynamicAmountNone {
			dynamic, ok := lowerDynamicAmount(declaration.Continuous.DynamicAmount, game.SourcePermanentReference())
			if !ok || declaration.Continuous.DynamicAmount.DynamicKind == compiler.DynamicAmountSourcePower {
				return game.ContinuousEffect{}, false
			}
			effect.PowerDelta = 0
			effect.ToughnessDelta = 0
			if power := dynamicSignedQuantity(&dynamic, declaration.Continuous.PowerDelta); power.IsDynamic() {
				effect.PowerDeltaDynamic = power.DynamicAmount()
			}
			if toughness := dynamicSignedQuantity(&dynamic, declaration.Continuous.ToughnessDelta); toughness.IsDynamic() {
				effect.ToughnessDeltaDynamic = toughness.DynamicAmount()
			}
		}
	case compiler.StaticContinuousGrantKeywords:
		if layer != game.LayerAbility {
			return game.ContinuousEffect{}, false
		}
		if keywords, ok := mixedStaticKeywords(declaration.Continuous.Keywords); ok && len(keywords) > 0 {
			effect.AddKeywords = keywords
			return effect, true
		}
		ability, ok := lowerStaticGrantedAbility(declaration.Continuous.Keywords)
		if !ok {
			return game.ContinuousEffect{}, false
		}
		effect.AddAbilities = []game.Ability{&ability}
	case compiler.StaticContinuousGrantManaAbility:
		if layer != game.LayerAbility || declaration.Continuous.GrantedMana == nil {
			return game.ContinuousEffect{}, false
		}
		ability, ok := lowerStaticGrantedManaAbility(declaration.Continuous.GrantedMana)
		if !ok {
			return game.ContinuousEffect{}, false
		}
		effect.AddAbilities = []game.Ability{&ability}
	case compiler.StaticContinuousGrantAbility:
		if layer != game.LayerAbility || declaration.Continuous.GrantedAbility == nil {
			return game.ContinuousEffect{}, false
		}
		ability, ok := lowerStaticGrantedQuotedAbility(declaration.Continuous.GrantedAbility)
		if !ok {
			return game.ContinuousEffect{}, false
		}
		effect.AddAbilities = []game.Ability{ability}
	case compiler.StaticContinuousChangeControl:
		if layer != game.LayerControl {
			return game.ContinuousEffect{}, false
		}
		effect.NewController = opt.Val(game.Player1)
	case compiler.StaticContinuousSetBasePowerToughness:
		if layer != game.LayerPowerToughnessSet {
			return game.ContinuousEffect{}, false
		}
		effect.SetPower = opt.Val(game.PT{Value: declaration.Continuous.SetPower})
		effect.SetToughness = opt.Val(game.PT{Value: declaration.Continuous.SetToughness})
	case compiler.StaticContinuousAddColors, compiler.StaticContinuousSetColors:
		if layer != game.LayerColor {
			return game.ContinuousEffect{}, false
		}
		if declaration.Continuous.SetColorless {
			if declaration.Continuous.Operation != compiler.StaticContinuousSetColors ||
				len(declaration.Continuous.Colors) != 0 {
				return game.ContinuousEffect{}, false
			}
			effect.SetColorless = true
			return effect, true
		}
		if len(declaration.Continuous.Colors) == 0 {
			return game.ContinuousEffect{}, false
		}
		colors := slices.Clone(declaration.Continuous.Colors)
		if declaration.Continuous.Operation == compiler.StaticContinuousAddColors {
			effect.AddColors = colors
		} else {
			effect.SetColors = colors
		}
	case compiler.StaticContinuousAddTypes:
		if layer != game.LayerType {
			return game.ContinuousEffect{}, false
		}
		cardTypes, subtypes, ok := lowerStaticAddedTypes(declaration.Continuous)
		if !ok {
			return game.ContinuousEffect{}, false
		}
		effect.AddTypes = cardTypes
		effect.AddSubtypes = subtypes
	case compiler.StaticContinuousAddSubtypeFromEntryChoice:
		if layer != game.LayerType {
			return game.ContinuousEffect{}, false
		}
		effect.AddSubtypeFromEntryChoice = game.EntryTypeChoiceKey
	case compiler.StaticContinuousSetTypes, compiler.StaticContinuousSetSubtypes:
		if layer != game.LayerType {
			return game.ContinuousEffect{}, false
		}
		cardTypes, subtypes, ok := lowerStaticSetTypes(declaration.Continuous)
		if !ok {
			return game.ContinuousEffect{}, false
		}
		effect.SetTypes = cardTypes
		effect.SetSubtypes = subtypes
	case compiler.StaticContinuousRemoveAllAbilities:
		if layer != game.LayerAbility {
			return game.ContinuousEffect{}, false
		}
		effect.RemoveAllAbilities = true
	default:
		return game.ContinuousEffect{}, false
	}
	return effect, true
}

// lowerStaticGrantedManaAbility builds the runtime mana ability conferred by a
// permanent-ability grant from the closed typed forms the compiler recognized:
// the bare tap-for-one-mana-of-any-color ability and the Treasure-style
// sacrifice ability that adds N mana of one chosen color.
func lowerStaticGrantedManaAbility(granted *compiler.StaticGrantedManaAbility) (game.ManaAbility, bool) {
	if !granted.TapCost {
		return game.ManaAbility{}, false
	}
	switch {
	case granted.AnyColor:
		if granted.Amount != 1 || granted.Sacrifice || granted.AnyOneColor {
			return game.ManaAbility{}, false
		}
		return game.TapAnyColorManaAbility(), true
	case granted.AnyOneColor:
		if granted.Amount < 2 || !granted.Sacrifice {
			return game.ManaAbility{}, false
		}
		return game.TapSacrificeAnyOneColorManaAbility(granted.Text, granted.Amount), true
	case granted.Colorless:
		if granted.Amount != 1 || granted.Sacrifice || granted.AnyColor {
			return game.ManaAbility{}, false
		}
		return game.TapManaAbility(mana.C), true
	default:
		return game.ManaAbility{}, false
	}
}

// lowerStaticGrantedQuotedAbility compiles and lowers a quoted triggered or
// activated ability conferred by a static grant ("Equipped creature has
// '<quoted ability>'."). The parser parsed the quoted body once; this recursive
// compile + lower mirrors the reminder-mana-ability pattern and produces the
// runtime ability attached as a granted ability of the continuous effect.
func lowerStaticGrantedQuotedAbility(granted *parser.StaticGrantedAbilitySyntax) (game.Ability, bool) {
	innerDocument, innerDiags := granted.Inner()
	if len(innerDiags) != 0 {
		return nil, false
	}
	innerComp, compilerDiags := compiler.Compile(innerDocument, compiler.Context{})
	if len(compilerDiags) != 0 ||
		len(innerComp.Abilities) != 1 ||
		len(innerComp.Syntax.Abilities) != 1 {
		return nil, false
	}
	lowered, diagnostic := lowerExecutableAbility("", false, nil, innerComp.Abilities[0], &innerComp.Syntax.Abilities[0])
	if diagnostic != nil {
		return nil, false
	}
	switch {
	case lowered.triggeredAbility.Exists:
		ability := lowered.triggeredAbility.Val
		return &ability, true
	case lowered.activatedAbility.Exists:
		ability := lowered.activatedAbility.Val
		return &ability, true
	case lowered.manaAbility.Exists:
		ability := lowered.manaAbility.Val
		return &ability, true
	default:
		return nil, false
	}
}

func lowerStaticAddedTypes(continuous *compiler.StaticContinuousDeclaration) ([]types.Card, []types.Sub, bool) {
	cardTypes := make([]types.Card, 0, len(continuous.AddTypes))
	for _, cardType := range continuous.AddTypes {
		value, ok := lowerStaticCardType(cardType)
		if !ok {
			return nil, nil, false
		}
		cardTypes = append(cardTypes, value)
	}
	subtypes := slices.Clone(continuous.AddSubtypes)
	if len(cardTypes) == 0 && len(subtypes) == 0 {
		return nil, nil, false
	}
	return cardTypes, subtypes, true
}

func lowerStaticSetTypes(continuous *compiler.StaticContinuousDeclaration) ([]types.Card, []types.Sub, bool) {
	cardTypes := make([]types.Card, 0, len(continuous.SetTypes))
	for _, cardType := range continuous.SetTypes {
		value, ok := lowerStaticCardType(cardType)
		if !ok {
			return nil, nil, false
		}
		cardTypes = append(cardTypes, value)
	}
	subtypes := slices.Clone(continuous.SetSubtypes)
	if len(cardTypes) == 0 && len(subtypes) == 0 {
		return nil, nil, false
	}
	return cardTypes, subtypes, true
}

func lowerStaticContinuousLayer(layer compiler.StaticContinuousLayer) (game.ContinuousLayer, bool) {
	switch layer {
	case compiler.StaticLayerAbility:
		return game.LayerAbility, true
	case compiler.StaticLayerPowerToughnessModify:
		return game.LayerPowerToughnessModify, true
	case compiler.StaticLayerControl:
		return game.LayerControl, true
	case compiler.StaticLayerPowerToughnessSet:
		return game.LayerPowerToughnessSet, true
	case compiler.StaticLayerColor:
		return game.LayerColor, true
	case compiler.StaticLayerType:
		return game.LayerType, true
	default:
		return 0, false
	}
}

func lowerStaticGrantedAbility(keywords []compiler.CompiledKeyword) (game.StaticAbility, bool) {
	if len(keywords) != 1 {
		return game.StaticAbility{}, false
	}
	switch keywords[0].Kind {
	case parser.KeywordProtection:
		if !keywords[0].ProtectionKnown {
			return game.StaticAbility{}, false
		}
		return staticAbilityFromProtectionKeyword(keywords[0].Protection, ""), true
	case parser.KeywordWard:
		if keywords[0].ParameterKind != parser.KeywordParameterManaCost || len(keywords[0].ManaCost) == 0 {
			return game.StaticAbility{}, false
		}
		return game.WardStaticAbility(slices.Clone(keywords[0].ManaCost)), true
	default:
		return game.StaticAbility{}, false
	}
}

func appendStaticRuleDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	var affectedSource, affectedAttached bool
	switch declaration.Group.Domain {
	case compiler.StaticGroupSource:
		affectedSource = declaration.Rule.Kind != compiler.StaticRuleAdditionalTriggerForChosenCreatureType
	case compiler.StaticGroupAttachedObject:
		affectedAttached = true
	default:
		return false
	}
	if declaration.Rule.Domain != staticRuleDomain(declaration.Rule.Kind) {
		return false
	}
	effects, ok := lowerStaticRuleEffects(declaration.Rule.Kind)
	if !ok {
		return false
	}
	if declaration.Rule.Kind == compiler.StaticRuleCantBeBlockedByCreaturesWith {
		restriction, ok := lowerStaticBlockerRestriction(declaration.Rule.Blocker)
		if !ok {
			return false
		}
		for i := range effects {
			effects[i].BlockerRestriction = restriction
		}
	}
	functionZone, ok := lowerStaticZone(declaration.Rule.Zone)
	if !ok || (body.ZoneOfFunction != zone.None && body.ZoneOfFunction != functionZone) {
		return false
	}
	body.ZoneOfFunction = functionZone
	for i := range effects {
		effects[i].AffectedSource = affectedSource
		effects[i].AffectedAttached = affectedAttached
		body.RuleEffects = append(body.RuleEffects, effects[i])
	}
	return true
}

func staticRuleDomain(kind compiler.StaticRuleKind) compiler.StaticRuleDomain {
	switch kind {
	case compiler.StaticRuleCantAttack, compiler.StaticRuleMustAttack, compiler.StaticRuleCantAttackYou:
		return compiler.StaticRuleDomainAttack
	case compiler.StaticRuleCantBlock, compiler.StaticRuleCantBeBlocked, compiler.StaticRuleMustBeBlocked,
		compiler.StaticRuleCantBeBlockedByMoreThanOne, compiler.StaticRuleCantBeBlockedByCreaturesWith:
		return compiler.StaticRuleDomainBlock
	case compiler.StaticRuleCantBeCountered:
		return compiler.StaticRuleDomainCountering
	case compiler.StaticRuleCantAttackOrBlock:
		return compiler.StaticRuleDomainAttackBlock
	case compiler.StaticRuleDoesntUntap:
		return compiler.StaticRuleDomainUntap
	case compiler.StaticRuleAdditionalTriggerForChosenCreatureType:
		return compiler.StaticRuleDomainTrigger
	default:
		return compiler.StaticRuleDomainUnknown
	}
}

// lowerStaticRuleEffects lowers a static rule kind into one or more runtime rule
// effect templates (Kind and any rule-specific fields, but not the affected
// subject). The compound "can't attack or block" expands into separate
// can't-attack and can't-block effects; the defender-scoped "can't attack you or
// planeswalkers you control" carries a DefendingPlayer restriction.
// appendStaticPlayerRuleDeclaration lowers a player-scoped static rule into a
// controller-scoped runtime rule effect on the static ability body.
func appendStaticPlayerRuleDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.Player == nil {
		return false
	}
	switch declaration.Player.Kind {
	case compiler.StaticPlayerRuleNoMaximumHandSize:
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectNoMaximumHandSize,
			AffectedPlayer: game.PlayerYou,
		})
		return true
	case compiler.StaticPlayerRuleAttackTax:
		if declaration.Player.AttackTaxGeneric <= 0 {
			return false
		}
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:             game.RuleEffectAttackTax,
			AffectedPlayer:   game.PlayerYou,
			AttackTaxGeneric: declaration.Player.AttackTaxGeneric,
		})
		return true
	case compiler.StaticPlayerRuleAdditionalLandPlays:
		if declaration.Player.AdditionalLandPlays <= 0 {
			return false
		}
		affected := game.PlayerYou
		if declaration.Player.AffectsAllPlayers {
			affected = game.PlayerAny
		}
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:                game.RuleEffectAdditionalLandPlays,
			AffectedPlayer:      affected,
			AdditionalLandPlays: declaration.Player.AdditionalLandPlays,
		})
		return true
	case compiler.StaticPlayerRulePlayLandsFromGraveyard:
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectPlayLandsFromZone,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Graveyard,
			PermanentTypes: []types.Card{types.Land},
		})
		return true
	case compiler.StaticPlayerRulePlayLandsFromLibraryTop:
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectPlayLandsFromZone,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Library,
			PermanentTypes: []types.Card{types.Land},
			TopCardOnly:    true,
		})
		return true
	case compiler.StaticPlayerRulePlayWithTopCardRevealed:
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectPlayWithTopCardRevealed,
			AffectedPlayer: game.PlayerYou,
		})
		return true
	case compiler.StaticPlayerRuleLookAtTopCardAnyTime:
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectLookAtTopCardAnyTime,
			AffectedPlayer: game.PlayerYou,
		})
		return true
	case compiler.StaticPlayerRuleCastSpellsFromLibraryTop:
		var spellTypes []types.Card
		if len(declaration.Player.SpellTypes) > 0 {
			spellTypes = append([]types.Card(nil), declaration.Player.SpellTypes...)
		}
		effect := game.RuleEffect{
			Kind:           game.RuleEffectCastSpellsFromZone,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Library,
			SpellTypes:     spellTypes,
			SpellColorless: declaration.Player.CastColorless,
			TopCardOnly:    true,
		}
		if declaration.Player.CastChosenCreatureType {
			effect.SpellChosenSubtypeFrom = game.EntryTypeChoiceKey
		}
		body.RuleEffects = append(body.RuleEffects, effect)
		if declaration.Player.AlsoPlayLands {
			body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
				Kind:           game.RuleEffectPlayLandsFromZone,
				AffectedPlayer: game.PlayerYou,
				CastFromZone:   zone.Library,
				PermanentTypes: []types.Card{types.Land},
				TopCardOnly:    true,
			})
		}
		return true
	case compiler.StaticPlayerRuleCastThisFromGraveyard:
		body.ZoneOfFunction = zone.Graveyard
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectCastFromZone,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Graveyard,
			AffectedSource: true,
		})
		return true
	case compiler.StaticPlayerRuleCastThisFromExile:
		body.ZoneOfFunction = zone.Exile
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectCastFromZone,
			AffectedPlayer: game.PlayerYou,
			CastFromZone:   zone.Exile,
			AffectedSource: true,
		})
		return true
	default:
		return false
	}
}

// appendStaticOpponentActionRestrictionDeclaration adds the continuous cast and
// activation prohibitions described by an opponent action restriction. "Your
// opponents"/"each opponent" affects every opponent of the controller; "players"
// and the passive voice affect every player.
func appendStaticOpponentActionRestrictionDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	restriction := declaration.OpponentRestriction
	if restriction == nil || (!restriction.RestrictCastSpells && len(restriction.ActivateTypes) == 0) {
		return false
	}
	affected := game.PlayerOpponent
	if restriction.AffectsAllPlayers {
		affected = game.PlayerAny
	}
	if restriction.RestrictCastSpells {
		zones, ok := lowerCastFromZones(restriction)
		if !ok {
			return false
		}
		if len(zones) > 0 {
			body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
				Kind:                           game.RuleEffectCantCastFromZones,
				AffectedPlayer:                 affected,
				CantCastFromZones:              zones,
				RestrictedDuringControllerTurn: restriction.DuringControllerTurn,
			})
		} else {
			body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
				Kind:                           game.RuleEffectCantCastSpells,
				AffectedPlayer:                 affected,
				RestrictedDuringControllerTurn: restriction.DuringControllerTurn,
			})
		}
	}
	if len(restriction.ActivateTypes) > 0 {
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:                           game.RuleEffectCantActivateAbilities,
			AffectedPlayer:                 affected,
			PermanentTypes:                 append([]types.Card(nil), restriction.ActivateTypes...),
			RestrictedDuringControllerTurn: restriction.DuringControllerTurn,
		})
	}
	return true
}

// lowerCastFromZones maps a cast prohibition's parser-owned zone scope onto the
// closed runtime zones. The "anywhere other than their hands" form expands to
// every non-hand cast zone; an explicit zone list maps each named zone. A cast
// prohibition without a zone scope returns no zones, signalling a full
// prohibition.
func lowerCastFromZones(restriction *compiler.StaticOpponentActionRestrictionDeclaration) ([]zone.Type, bool) {
	if restriction.CastOnlyFromHand {
		return []zone.Type{zone.Graveyard, zone.Exile, zone.Library, zone.Command}, true
	}
	if len(restriction.CastFromZones) == 0 {
		return nil, true
	}
	zones := make([]zone.Type, 0, len(restriction.CastFromZones))
	for _, kind := range restriction.CastFromZones {
		mapped, ok := lowerCastFromZone(kind)
		if !ok {
			return nil, false
		}
		zones = append(zones, mapped)
	}
	return zones, true
}

// lowerCastFromZone maps a single parser cast-zone kind onto its runtime zone.
func lowerCastFromZone(kind parser.StaticDeclarationCastZoneKind) (zone.Type, bool) {
	switch kind {
	case parser.StaticDeclarationCastZoneGraveyard:
		return zone.Graveyard, true
	case parser.StaticDeclarationCastZoneLibrary:
		return zone.Library, true
	case parser.StaticDeclarationCastZoneExile:
		return zone.Exile, true
	case parser.StaticDeclarationCastZoneCommand:
		return zone.Command, true
	default:
		return zone.None, false
	}
}

// appendStaticEnterBattlefieldRestrictionDeclaration lowers a "<filter> cards in
// <zones> can't enter the battlefield." declaration into a global
// RuleEffectCantEnterFromZones rule effect on the static ability body. The
// "creature" filter restricts only creature cards; "permanent" restricts every
// permanent card; "nonland permanent" restricts every permanent card except
// lands. The runtime collects the body as an active rule effect (it functions on
// the battlefield) and prevents matching cards from entering out of the listed
// zones.
func appendStaticEnterBattlefieldRestrictionDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	restriction := declaration.EnterRestriction
	if restriction == nil || len(restriction.FromZones) == 0 {
		return false
	}
	zones := make([]zone.Type, 0, len(restriction.FromZones))
	for _, kind := range restriction.FromZones {
		mapped, ok := lowerCastFromZone(kind)
		if !ok {
			return false
		}
		zones = append(zones, mapped)
	}
	effect := game.RuleEffect{
		Kind:           game.RuleEffectCantEnterFromZones,
		EnterFromZones: zones,
	}
	switch restriction.Filter {
	case parser.StaticDeclarationEnterFilterCreature:
		effect.PermanentTypes = []types.Card{types.Creature}
	case parser.StaticDeclarationEnterFilterPermanent:
	case parser.StaticDeclarationEnterFilterNonlandPermanent:
		effect.EnterExcludeLandCards = true
	default:
		return false
	}
	body.RuleEffects = append(body.RuleEffects, effect)
	return true
}

// appendStaticSpellUncounterableDeclaration lowers a "[<type>] spells you control
// can't be countered." declaration into a controller-scoped can't-be-countered
// rule effect on the static ability body. The body functions on the battlefield
// (no Stack zone), so the runtime collects it as an active rule effect and stops
// counters targeting matching spells the controller casts.
func appendStaticSpellUncounterableDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.SpellUncounterable == nil ||
		declaration.Group.Domain != compiler.StaticGroupControllerSpells {
		return false
	}
	spellTypes := make([]types.Card, 0, len(declaration.SpellUncounterable.SpellTypes))
	for _, spellType := range declaration.SpellUncounterable.SpellTypes {
		cardType, ok := lowerStaticCardType(spellType)
		if !ok {
			return false
		}
		spellTypes = append(spellTypes, cardType)
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:               game.RuleEffectCantBeCountered,
		AffectedController: game.ControllerYou,
		SpellTypes:         spellTypes,
	})
	return true
}

// appendStaticCastAsThoughFlashDeclaration lowers a "You may cast [<filter>]
// spells as though they had flash." declaration into a controller-scoped
// cast-as-though-flash rule effect on the static ability body. SpellTypes and
// SpellSubtypes carry the optional card-type and subtype filters; empty filters
// grant the permission for every spell. The body functions on the battlefield,
// so the runtime collects it as an active rule effect and lets the controller
// cast matching spells at instant speed.
func appendStaticCastAsThoughFlashDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.CastAsThoughFlash == nil ||
		declaration.Group.Domain != compiler.StaticGroupControllerSpells {
		return false
	}
	spellTypes := make([]types.Card, 0, len(declaration.CastAsThoughFlash.SpellTypes))
	for _, spellType := range declaration.CastAsThoughFlash.SpellTypes {
		cardType, ok := lowerStaticCardType(spellType)
		if !ok {
			return false
		}
		spellTypes = append(spellTypes, cardType)
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectCastSpellsAsThoughFlash,
		AffectedPlayer: game.PlayerYou,
		SpellTypes:     spellTypes,
		SpellSubtypes:  declaration.CastAsThoughFlash.SpellSubtypes,
	})
	return true
}

// appendStaticUntapStepDeclaration lowers an "Untap <group> you control during
// each other player's untap step." declaration into a controller-scoped extra-
// untap rule effect on the static ability body. The self form scopes the effect
// to the source permanent; the group forms filter the controller's permanents by
// card type. The runtime collects the body as an active rule effect (it
// functions on the battlefield) and untaps the matching permanents during each
// other player's untap step.
func appendStaticUntapStepDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.Untap == nil {
		return false
	}
	if declaration.Untap.Self {
		if declaration.Group.Domain != compiler.StaticGroupSource ||
			len(declaration.Untap.PermanentTypes) != 0 {
			return false
		}
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectUntapDuringOtherPlayersUntapStep,
			AffectedSource: true,
		})
		return true
	}
	if declaration.Group.Domain != compiler.StaticGroupSourceControllerPermanents {
		return false
	}
	permanentTypes := make([]types.Card, 0, len(declaration.Untap.PermanentTypes))
	for _, cardType := range declaration.Untap.PermanentTypes {
		value, ok := lowerStaticCardType(cardType)
		if !ok {
			return false
		}
		permanentTypes = append(permanentTypes, value)
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:               game.RuleEffectUntapDuringOtherPlayersUntapStep,
		AffectedController: game.ControllerYou,
		PermanentTypes:     permanentTypes,
	})
	return true
}

// appendStaticEnteringTriggerMultiplierDeclaration lowers an "If <filter>
// entering causes a triggered ability of a permanent you control to trigger,
// that ability triggers an additional time." declaration into a controller-scoped
// trigger-multiplier rule effect on the static ability body. The runtime collects
// it as an active rule effect and fires a matching triggered ability one extra
// time. PermanentTypes carries the entering permanent's type filter; an empty
// filter matches any entering permanent.
func appendStaticEnteringTriggerMultiplierDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.EnteringMultiplier == nil {
		return false
	}
	permanentTypes := append([]types.Card(nil), declaration.EnteringMultiplier.EnteringTypes...)
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectAdditionalTriggerForEnteringPermanent,
		PermanentTypes: permanentTypes,
	})
	return true
}

func lowerStaticRuleEffects(kind compiler.StaticRuleKind) ([]game.RuleEffect, bool) {
	switch kind {
	case compiler.StaticRuleCantAttackOrBlock:
		return []game.RuleEffect{
			{Kind: game.RuleEffectCantAttack},
			{Kind: game.RuleEffectCantBlock},
		}, true
	case compiler.StaticRuleCantAttackYou:
		return []game.RuleEffect{
			{Kind: game.RuleEffectCantAttack, DefendingPlayer: game.PlayerYou},
		}, true
	default:
		single, ok := lowerStaticRuleKind(kind)
		if !ok {
			return nil, false
		}
		return []game.RuleEffect{{Kind: single}}, true
	}
}

func lowerStaticRuleKind(kind compiler.StaticRuleKind) (game.RuleEffectKind, bool) {
	switch kind {
	case compiler.StaticRuleCantBlock:
		return game.RuleEffectCantBlock, true
	case compiler.StaticRuleCantBeBlocked:
		return game.RuleEffectCantBeBlocked, true
	case compiler.StaticRuleCantBeBlockedByMoreThanOne:
		return game.RuleEffectCantBeBlockedByMoreThanOne, true
	case compiler.StaticRuleCantBeBlockedByCreaturesWith:
		return game.RuleEffectCantBeBlockedByCreaturesWith, true
	case compiler.StaticRuleCantAttack:
		return game.RuleEffectCantAttack, true
	case compiler.StaticRuleMustAttack:
		return game.RuleEffectMustAttack, true
	case compiler.StaticRuleMustBeBlocked:
		return game.RuleEffectMustBeBlocked, true
	case compiler.StaticRuleCantBeCountered:
		return game.RuleEffectCantBeCountered, true
	case compiler.StaticRuleDoesntUntap:
		return game.RuleEffectDoesntUntap, true
	case compiler.StaticRuleAdditionalTriggerForChosenCreatureType:
		return game.RuleEffectAdditionalTriggerForChosenCreatureType, true
	default:
		return game.RuleEffectNone, false
	}
}

// lowerStaticBlockerRestriction lowers the compiler's closed blocker
// characteristic into the runtime BlockerRestriction carried by a restricted
// "can't be blocked by creatures with ..." rule effect.
func lowerStaticBlockerRestriction(restriction compiler.StaticBlockerRestriction) (game.BlockerRestriction, bool) {
	switch restriction.Kind {
	case compiler.StaticBlockerRestrictionFlying:
		return game.BlockerRestriction{Kind: game.BlockerRestrictionFlying}, true
	case compiler.StaticBlockerRestrictionPowerOrLess:
		return game.BlockerRestriction{Kind: game.BlockerRestrictionPowerLessOrEqual, Power: restriction.Amount}, true
	case compiler.StaticBlockerRestrictionPowerOrGreater:
		return game.BlockerRestriction{Kind: game.BlockerRestrictionPowerGreaterOrEqual, Power: restriction.Amount}, true
	case compiler.StaticBlockerRestrictionColor:
		return game.BlockerRestriction{Kind: game.BlockerRestrictionColor, Color: restriction.Color}, true
	case compiler.StaticBlockerRestrictionArtifact:
		return game.BlockerRestriction{Kind: game.BlockerRestrictionArtifact}, true
	default:
		return game.BlockerRestriction{}, false
	}
}

func lowerStaticZone(value compiler.StaticZone) (zone.Type, bool) {
	switch value {
	case compiler.StaticZoneBattlefield:
		return zone.None, true
	case compiler.StaticZoneStack:
		return zone.Stack, true
	case compiler.StaticZoneHand:
		return zone.Hand, true
	default:
		return zone.None, false
	}
}

func appendStaticCostModifierDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.Cost.Kind == compiler.StaticCostModifierSpell {
		return appendStaticSpellCostModifierDeclaration(body, declaration)
	}
	if declaration.Group.Domain == compiler.StaticGroupControllerEquipment {
		return appendStaticEquipCostModifierDeclaration(body, declaration)
	}
	if declaration.Group.Domain != compiler.StaticGroupControllerHandCards ||
		declaration.Cost.Kind != compiler.StaticCostModifierAbility ||
		declaration.Cost.AbilityKeyword != parser.KeywordCycling {
		return false
	}
	modifier := game.CostModifier{
		Kind:               game.CostModifierAbility,
		AbilityKeyword:     game.Cycling,
		GenericReduction:   declaration.Cost.GenericReduction,
		FirstCycleEachTurn: declaration.Cost.FirstCycleEachTurn,
	}
	if declaration.Cost.ReplaceManaCost {
		manaCost, err := parseManaCostValue(declaration.Cost.SetManaCost)
		if err != nil {
			return false
		}
		modifier.SetManaCost = opt.Val(manaCost)
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectCostModifier,
		AffectedPlayer: game.PlayerYou,
		CostModifier:   modifier,
	})
	return true
}

// appendStaticEquipCostModifierDeclaration lowers "Equipment you control have
// equip {N}." into a controller-scoped rule effect that replaces the Equip
// activation cost of the controller's Equipment with {N}. The runtime matches the
// affected abilities by the Equip keyword, so the affected group narrows to the
// controller without an explicit battlefield selection.
func appendStaticEquipCostModifierDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.Cost.Kind != compiler.StaticCostModifierAbility || !declaration.Cost.ReplaceManaCost {
		return false
	}
	keyword, ok := runtimeKeyword(declaration.Cost.AbilityKeyword)
	if !ok || keyword != game.Equip {
		return false
	}
	manaCost, err := parseManaCostValue(declaration.Cost.SetManaCost)
	if err != nil {
		return false
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectCostModifier,
		AffectedPlayer: game.PlayerYou,
		CostModifier: game.CostModifier{
			Kind:           game.CostModifierAbility,
			AbilityKeyword: keyword,
			SetManaCost:    opt.Val(manaCost),
		},
	})
	return true
}

// appendStaticSpellCostModifierDeclaration lowers a controller cast-cost modifier
// into one rule effect per affected spell type, or a single rule effect when the
// affected spells are constrained only by color, subtype, or not at all. A color
// filter combines with a single card-type filter ("black creature spells"); the
// color-disjunction and subtype filters are each mutually exclusive with the
// card-type filter.
func appendStaticSpellCostModifierDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.Group.Domain != compiler.StaticGroupControllerSpells {
		return false
	}
	cost := declaration.Cost
	if (cost.GenericReduction == 0) == (cost.GenericIncrease == 0) {
		return false
	}
	base := game.CostModifier{
		Kind:             game.CostModifierSpell,
		GenericReduction: cost.GenericReduction,
		GenericIncrease:  cost.GenericIncrease,
	}
	if cost.ChosenSubtypeFromEntryChoice {
		base.ChosenSubtypeFromEntryChoice = true
	}
	if len(cost.SpellColors) != 0 {
		if cost.MatchSpellColor || len(cost.SpellTypes) != 0 || len(cost.SpellSubtypes) != 0 {
			return false
		}
		modifier := base
		modifier.MatchColors = slices.Clone(cost.SpellColors)
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectCostModifier,
			AffectedPlayer: game.PlayerYou,
			CostModifier:   modifier,
		})
		return true
	}
	if len(cost.SpellSubtypes) != 0 {
		if len(cost.SpellTypes) != 0 {
			return false
		}
		base.MatchSubtypes = slices.Clone(cost.SpellSubtypes)
	}
	if cost.MatchSpellColor {
		base.MatchColor = true
		base.Color = cost.SpellColor
	}
	if len(cost.SpellTypes) == 0 {
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectCostModifier,
			AffectedPlayer: game.PlayerYou,
			CostModifier:   base,
		})
		return true
	}
	for _, spellType := range cost.SpellTypes {
		cardType, ok := lowerStaticCardType(spellType)
		if !ok {
			return false
		}
		modifier := base
		modifier.MatchCardType = true
		modifier.CardType = cardType
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectCostModifier,
			AffectedPlayer: game.PlayerYou,
			CostModifier:   modifier,
		})
	}
	return true
}

func appendStaticCardAbilityGrantDeclaration(body *game.StaticAbility, declaration compiler.StaticDeclaration) bool {
	if declaration.Group.Domain != compiler.StaticGroupControllerHandCards ||
		declaration.CardGrant.Keyword.Kind != parser.KeywordCycling ||
		declaration.CardGrant.Keyword.ParameterKind != parser.KeywordParameterManaCost {
		return false
	}
	selection, ok := lowerStaticSelection(declaration.Group.Selection)
	if !ok || selection.Empty() {
		return false
	}
	if len(declaration.CardGrant.Keyword.ManaCost) == 0 {
		return false
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectGrantHandCardAbility,
		AffectedPlayer: game.PlayerYou,
		CardSelection:  selection,
		GrantedAbility: game.CyclingActivatedAbility(slices.Clone(declaration.CardGrant.Keyword.ManaCost)),
	})
	return true
}

type loweredStaticGroupReference struct {
	Group          game.GroupReference
	AffectedSource bool
}

func lowerStaticGroupReference(reference compiler.StaticGroupReference) (loweredStaticGroupReference, bool) {
	selection, ok := lowerStaticSelection(reference.Selection)
	if !ok {
		return loweredStaticGroupReference{}, false
	}
	switch reference.Domain {
	case compiler.StaticGroupSource:
		if !selection.Empty() || reference.ExcludeSource {
			return loweredStaticGroupReference{}, false
		}
		return loweredStaticGroupReference{AffectedSource: true}, true
	case compiler.StaticGroupAttachedObject:
		if !selection.Empty() || reference.ExcludeSource {
			return loweredStaticGroupReference{}, false
		}
		return loweredStaticGroupReference{Group: game.AttachedObjectGroup(game.SourcePermanentReference())}, true
	case compiler.StaticGroupBattlefield:
		if reference.ExcludeSource {
			return loweredStaticGroupReference{
				Group: game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference()),
			}, true
		}
		return loweredStaticGroupReference{Group: game.BattlefieldGroup(selection)}, true
	case compiler.StaticGroupSourceControllerPermanents:
		if reference.ExcludeSource {
			return loweredStaticGroupReference{
				Group: game.ObjectControlledGroupExcluding(
					game.SourcePermanentReference(),
					selection,
					game.SourcePermanentReference(),
				),
			}, true
		}
		return loweredStaticGroupReference{
			Group: game.ObjectControlledGroup(game.SourcePermanentReference(), selection),
		}, true
	default:
		return loweredStaticGroupReference{}, false
	}
}

func lowerStaticSelection(selection compiler.StaticSelection) (game.Selection, bool) {
	combatState, ok := lowerStaticCombatState(selection.CombatState)
	if !ok {
		return game.Selection{}, false
	}
	tapState, ok := lowerStaticTapState(selection.TapState)
	if !ok {
		return game.Selection{}, false
	}
	result := game.Selection{
		Controller:   lowerStaticController(selection.Controller),
		CombatState:  combatState,
		Tapped:       tapState,
		TokenOnly:    selection.TokenOnly,
		NonToken:     selection.NonToken,
		Supertypes:   slices.Clone(selection.Supertypes),
		ColorsAny:    slices.Clone(selection.ColorsAny),
		Colorless:    selection.Colorless,
		Multicolored: selection.Multicolored,
	}
	if selection.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selection.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		result.Keyword = keyword
	}
	if selection.ExcludedKeyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selection.ExcludedKeyword)
		if !ok {
			return game.Selection{}, false
		}
		result.ExcludedKeyword = keyword
	}
	if selection.Controller != compiler.ControllerAny && result.Controller == game.ControllerAny {
		return game.Selection{}, false
	}
	for _, cardType := range selection.RequiredTypes {
		value, ok := lowerStaticCardType(cardType)
		if !ok {
			return game.Selection{}, false
		}
		result.RequiredTypes = append(result.RequiredTypes, value)
	}
	result.SubtypesAny = append(result.SubtypesAny, selection.SubtypesAny...)
	if selection.SubtypeFromEntryChoice {
		result.SubtypeChoice = game.SubtypeChoiceSourceEntry
	}
	if selection.MatchCounter {
		result.MatchCounter = true
		result.RequiredCounter = selection.RequiredCounter
	}
	return result, len(result.Validate()) == 0
}

func lowerStaticCombatState(state compiler.StaticCombatState) (game.CombatStateFilter, bool) {
	switch state {
	case compiler.StaticCombatStateAny:
		return game.CombatStateAny, true
	case compiler.StaticCombatStateAttacking:
		return game.CombatStateAttacking, true
	case compiler.StaticCombatStateBlocking:
		return game.CombatStateBlocking, true
	default:
		return game.CombatStateAny, false
	}
}

func lowerStaticTapState(state compiler.StaticTapState) (game.TriState, bool) {
	switch state {
	case compiler.StaticTapStateAny:
		return game.TriAny, true
	case compiler.StaticTapStateTapped:
		return game.TriTrue, true
	case compiler.StaticTapStateUntapped:
		return game.TriFalse, true
	default:
		return game.TriAny, false
	}
}

func lowerStaticController(controller compiler.ControllerKind) game.ControllerRelation {
	switch controller {
	case compiler.ControllerYou:
		return game.ControllerYou
	case compiler.ControllerOpponent:
		return game.ControllerOpponent
	case compiler.ControllerNotYou:
		return game.ControllerNotYou
	default:
		return game.ControllerAny
	}
}

func lowerStaticCardType(cardType compiler.StaticCardType) (types.Card, bool) {
	switch cardType {
	case compiler.StaticCardTypeArtifact:
		return types.Artifact, true
	case compiler.StaticCardTypeCreature:
		return types.Creature, true
	case compiler.StaticCardTypeLand:
		return types.Land, true
	case compiler.StaticCardTypeEnchantment:
		return types.Enchantment, true
	case compiler.StaticCardTypeInstant:
		return types.Instant, true
	case compiler.StaticCardTypeSorcery:
		return types.Sorcery, true
	default:
		return "", false
	}
}

func canonicalStaticDeclarationVarName(declaration compiler.StaticDeclaration) string {
	if declaration.Kind == compiler.StaticDeclarationPlayerRule &&
		declaration.Condition == nil &&
		declaration.Player != nil &&
		declaration.Player.Kind == compiler.StaticPlayerRuleNoMaximumHandSize {
		return "game.NoMaximumHandSizeStaticBody"
	}
	if declaration.Kind == compiler.StaticDeclarationPlayerRule &&
		declaration.Condition == nil &&
		declaration.Player != nil &&
		declaration.Player.Kind == compiler.StaticPlayerRulePlayLandsFromGraveyard {
		return "game.PlayLandsFromGraveyardStaticBody"
	}
	if declaration.Kind == compiler.StaticDeclarationPlayerRule &&
		declaration.Condition == nil &&
		declaration.Player != nil &&
		declaration.Player.Kind == compiler.StaticPlayerRulePlayLandsFromLibraryTop {
		return "game.PlayLandsFromLibraryTopStaticBody"
	}
	if declaration.Kind == compiler.StaticDeclarationPlayerRule &&
		declaration.Condition == nil &&
		declaration.Player != nil &&
		declaration.Player.Kind == compiler.StaticPlayerRulePlayWithTopCardRevealed {
		return "game.PlayWithTopCardRevealedStaticBody"
	}
	if declaration.Kind == compiler.StaticDeclarationPlayerRule &&
		declaration.Condition == nil &&
		declaration.Player != nil &&
		declaration.Player.Kind == compiler.StaticPlayerRuleLookAtTopCardAnyTime {
		return "game.LookAtTopCardAnyTimeStaticBody"
	}
	if declaration.Kind != compiler.StaticDeclarationRule ||
		declaration.Rule == nil ||
		declaration.Condition != nil ||
		declaration.Group.Domain != compiler.StaticGroupSource {
		return ""
	}
	switch declaration.Rule.Kind {
	case compiler.StaticRuleCantBlock:
		return "game.CantBlockStaticBody"
	case compiler.StaticRuleCantBeBlocked:
		return "game.CantBeBlockedStaticBody"
	case compiler.StaticRuleCantAttack:
		return "game.CantAttackStaticBody"
	case compiler.StaticRuleMustAttack:
		return "game.MustAttackStaticBody"
	case compiler.StaticRuleMustBeBlocked:
		return "game.MustBeBlockedStaticBody"
	case compiler.StaticRuleCantBeCountered:
		return "game.CantBeCounteredStaticBody"
	case compiler.StaticRuleCantAttackOrBlock:
		return "game.CantAttackOrBlockStaticBody"
	default:
		return ""
	}
}

func staticDeclarationDiagnostic(ability compiler.CompiledAbility, summary, detail string) *shared.Diagnostic {
	return executableDiagnostic(ability, summary, detail)
}
