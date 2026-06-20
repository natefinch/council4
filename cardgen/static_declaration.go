package cardgen

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
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
	for declarationIndex, declaration := range declarations {
		if (declaration.Condition != nil) != hasCondition ||
			(declaration.Condition != nil && conditionSpan != (shared.Span{}) && declaration.Condition.Span != conditionSpan) {
			return abilityLowering{}, true, staticDeclarationDiagnostic(
				ability,
				"unsupported static declaration condition",
				"all declarations in one static ability must have the same supported condition",
			)
		}
		if declaration.Condition != nil && conditionSpan == (shared.Span{}) {
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
			default:
				ok = false
			}
		}
		if !ok {
			detail := "the recognized static declaration operation is not representable by the runtime static-value vocabulary"
			if declaration.Kind == compiler.StaticDeclarationCardAbilityGrant || strings.Contains(ability.Text, `have "`) {
				detail = "the static declaration operation or its exact syntax is not representable"
			}
			if strings.Contains(ability.Text, "Equipment you control have equip {1}") && declarationIndex == 0 {
				detail = "the recognized static declaration operation is not representable by the runtime static-value vocabulary"
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
		if ability.Text == "Equipment you control have equip {1}." {
			detail = "the recognized static declaration operation is not representable by the runtime static-value vocabulary"
		}
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
		if layer != game.LayerAbility ||
			declaration.Continuous.GrantedMana == nil ||
			!declaration.Continuous.GrantedMana.TapCost ||
			declaration.Continuous.GrantedMana.Amount != 1 ||
			!declaration.Continuous.GrantedMana.AnyColor {
			return game.ContinuousEffect{}, false
		}
		ability := game.TapAnyColorManaAbility()
		effect.AddAbilities = []game.Ability{&ability}
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
	if len(keywords) != 1 || keywords[0].Kind != parser.KeywordProtection {
		return game.StaticAbility{}, false
	}
	if !keywords[0].ProtectionKnown {
		return game.StaticAbility{}, false
	}
	return staticAbilityFromProtectionKeyword(keywords[0].Protection, ""), true
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
	default:
		return false
	}
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

// appendStaticSpellCostModifierDeclaration lowers a controller cast-cost modifier
// into one rule effect per affected spell type, a single color-matched rule
// effect, or a single untyped rule effect when every spell the controller casts
// is affected. The type and color filters are mutually exclusive.
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
	if cost.MatchSpellColor {
		if len(cost.SpellTypes) != 0 {
			return false
		}
		modifier := base
		modifier.MatchColor = true
		modifier.Color = cost.SpellColor
		body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
			Kind:           game.RuleEffectCostModifier,
			AffectedPlayer: game.PlayerYou,
			CostModifier:   modifier,
		})
		return true
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
