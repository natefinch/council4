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
	spans := make([]shared.Span, 0, len(declarations))
	for _, declaration := range declarations {
		spans = append(spans, declaration.Span)
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
			if power := dynamicSignedQuantity(dynamic, declaration.Continuous.PowerDelta); power.IsDynamic() {
				effect.PowerDeltaDynamic = power.DynamicAmount()
			}
			if toughness := dynamicSignedQuantity(dynamic, declaration.Continuous.ToughnessDelta); toughness.IsDynamic() {
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
		effect.AddAbilities = []game.Ability{ability}
	case compiler.StaticContinuousChangeControl:
		if layer != game.LayerControl {
			return game.ContinuousEffect{}, false
		}
		effect.NewController = opt.Val(game.Player1)
	default:
		return game.ContinuousEffect{}, false
	}
	return effect, true
}

func lowerStaticContinuousLayer(layer compiler.StaticContinuousLayer) (game.ContinuousLayer, bool) {
	switch layer {
	case compiler.StaticLayerAbility:
		return game.LayerAbility, true
	case compiler.StaticLayerPowerToughnessModify:
		return game.LayerPowerToughnessModify, true
	case compiler.StaticLayerControl:
		return game.LayerControl, true
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
	if declaration.Group.Domain != compiler.StaticGroupSource {
		return false
	}
	if declaration.Rule.Domain != staticRuleDomain(declaration.Rule.Kind) {
		return false
	}
	kind, ok := lowerStaticRuleKind(declaration.Rule.Kind)
	if !ok {
		return false
	}

	functionZone, ok := lowerStaticZone(declaration.Rule.Zone)
	if !ok || (body.ZoneOfFunction != zone.None && body.ZoneOfFunction != functionZone) {
		return false
	}
	body.ZoneOfFunction = functionZone
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:           kind,
		AffectedSource: true,
	})
	return true
}

func staticRuleDomain(kind compiler.StaticRuleKind) compiler.StaticRuleDomain {
	switch kind {
	case compiler.StaticRuleMustAttack:
		return compiler.StaticRuleDomainAttack
	case compiler.StaticRuleCantBlock, compiler.StaticRuleCantBeBlocked:
		return compiler.StaticRuleDomainBlock
	case compiler.StaticRuleCantBeCountered:
		return compiler.StaticRuleDomainCountering
	default:
		return compiler.StaticRuleDomainUnknown
	}
}

func lowerStaticRuleKind(kind compiler.StaticRuleKind) (game.RuleEffectKind, bool) {
	switch kind {
	case compiler.StaticRuleCantBlock:
		return game.RuleEffectCantBlock, true
	case compiler.StaticRuleCantBeBlocked:
		return game.RuleEffectCantBeBlocked, true
	case compiler.StaticRuleMustAttack:
		return game.RuleEffectMustAttack, true
	case compiler.StaticRuleCantBeCountered:
		return game.RuleEffectCantBeCountered, true
	default:
		return game.RuleEffectNone, false
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
	result := game.Selection{
		Controller: lowerStaticController(selection.Controller),
		TokenOnly:  selection.TokenOnly,
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
	default:
		return "", false
	}
}

func canonicalStaticDeclarationVarName(declaration compiler.StaticDeclaration) string {
	if declaration.Kind != compiler.StaticDeclarationRule ||
		declaration.Rule == nil ||
		declaration.Condition != nil {
		return ""
	}
	switch declaration.Rule.Kind {
	case compiler.StaticRuleCantBlock:
		return "game.CantBlockStaticBody"
	case compiler.StaticRuleCantBeBlocked:
		return "game.CantBeBlockedStaticBody"
	case compiler.StaticRuleMustAttack:
		return "game.MustAttackStaticBody"
	case compiler.StaticRuleCantBeCountered:
		return "game.CantBeCounteredStaticBody"
	default:
		return ""
	}
}

func staticDeclarationDiagnostic(ability compiler.CompiledAbility, summary, detail string) *shared.Diagnostic {
	return executableDiagnostic(ability, summary, detail)
}
