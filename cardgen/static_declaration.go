package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerStaticDeclarations is the only semantic Static Declaration to runtime
// static-value lowering path.
func lowerStaticDeclarations(
	ability oracle.CompiledAbility,
) (abilityLowering, bool, *oracle.Diagnostic) {
	if ability.Kind != oracle.AbilityStatic || ability.Static == nil || len(ability.Static.Declarations) == 0 {
		return abilityLowering{}, false, nil
	}
	if ability.Static.Blocker != oracle.StaticDeclarationBlockerNone {
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
	conditionSpan := oracle.Span{}
	hasCondition := declarations[0].Condition != nil
	for _, declaration := range declarations {
		if (declaration.Condition != nil) != hasCondition ||
			(declaration.Condition != nil && conditionSpan != (oracle.Span{}) && declaration.Condition.Span != conditionSpan) {
			return abilityLowering{}, true, staticDeclarationDiagnostic(
				ability,
				"unsupported static declaration condition",
				"all declarations in one static ability must have the same supported condition",
			)
		}
		if declaration.Condition != nil && conditionSpan == (oracle.Span{}) {
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
			case oracle.StaticDeclarationContinuous:
				ok = appendStaticContinuousDeclaration(&body, declaration)
			case oracle.StaticDeclarationRule:
				ok = appendStaticRuleDeclaration(&body, declaration)
			case oracle.StaticDeclarationCostModifier:
				ok = appendStaticCostModifierDeclaration(&body, declaration)
			case oracle.StaticDeclarationCardAbilityGrant:
				ok = appendStaticCardAbilityGrantDeclaration(&body, declaration)
				if ok {
					body.Text = declaration.CardGrant.Text
				}
			default:
				ok = false
			}
		}
		if !ok {
			return abilityLowering{}, true, staticDeclarationDiagnostic(
				ability,
				"unsupported static declaration operation",
				"the recognized static declaration operation is not representable by the runtime static-value vocabulary",
			)
		}
	}
	if len(declarations) == 1 {
		varName = canonicalStaticDeclarationVarName(&body)
	}
	spans := make([]oracle.Span, 0, len(declarations))
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

func lowerStaticDeclarationBlocker(ability oracle.CompiledAbility) *oracle.Diagnostic {
	if ability.Static == nil {
		return nil
	}
	switch ability.Static.Blocker {
	case oracle.StaticDeclarationBlockerHistoricCardSelection:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration group",
			"historic card predicates are not supported by the executable source backend",
		)
	case oracle.StaticDeclarationBlockerCondition:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration condition",
			"the static declaration has an unsupported or ambiguously scoped condition",
		)
	case oracle.StaticDeclarationBlockerDuration:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration duration",
			"the static declaration has a duration that is not valid for a source-derived static value",
		)
	case oracle.StaticDeclarationBlockerGroup:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration group",
			"the static declaration affected group is unsupported or ambiguous",
		)
	case oracle.StaticDeclarationBlockerOperation:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration operation",
			"the static declaration operation or its exact syntax is not representable",
		)
	case oracle.StaticDeclarationBlockerShell:
		return staticDeclarationDiagnostic(
			ability,
			"unsupported static declaration shell",
			"the static declaration shell carries unsupported additional semantics",
		)
	default:
		return nil
	}
}

func staticDeclarationPayloadValid(declaration oracle.StaticDeclaration) bool {
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
	case oracle.StaticDeclarationContinuous:
		return declaration.Continuous != nil
	case oracle.StaticDeclarationRule:
		return declaration.Rule != nil
	case oracle.StaticDeclarationCostModifier:
		return declaration.Cost != nil
	case oracle.StaticDeclarationCardAbilityGrant:
		return declaration.CardGrant != nil
	default:
		return false
	}
}

func appendStaticContinuousDeclaration(body *game.StaticAbility, declaration oracle.StaticDeclaration) bool {
	effect, ok := lowerStaticContinuousDeclaration(declaration)
	if !ok {
		return false
	}
	body.ContinuousEffects = append(body.ContinuousEffects, effect)
	return true
}

func lowerStaticContinuousDeclaration(declaration oracle.StaticDeclaration) (game.ContinuousEffect, bool) {
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
	case oracle.StaticContinuousModifyPowerToughness:
		if layer != game.LayerPowerToughnessModify {
			return game.ContinuousEffect{}, false
		}
		effect.PowerDelta = compiledSignedAmountValue(declaration.Continuous.PowerDelta)
		effect.ToughnessDelta = compiledSignedAmountValue(declaration.Continuous.ToughnessDelta)
		if declaration.Continuous.DynamicAmount.DynamicKind != oracle.DynamicAmountNone {
			dynamic, ok := lowerDynamicAmount(declaration.Continuous.DynamicAmount, game.SourcePermanentReference())
			if !ok || declaration.Continuous.DynamicAmount.DynamicKind == oracle.DynamicAmountSourcePower {
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
	case oracle.StaticContinuousGrantKeywords:
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
	default:
		return game.ContinuousEffect{}, false
	}
	return effect, true
}

func lowerStaticContinuousLayer(layer oracle.StaticContinuousLayer) (game.ContinuousLayer, bool) {
	switch layer {
	case oracle.StaticLayerAbility:
		return game.LayerAbility, true
	case oracle.StaticLayerPowerToughnessModify:
		return game.LayerPowerToughnessModify, true
	default:
		return 0, false
	}
}

func lowerStaticGrantedAbility(keywords []oracle.CompiledKeyword) (game.StaticAbility, bool) {
	if len(keywords) != 1 || keywords[0].Name != "Protection" {
		return game.StaticAbility{}, false
	}
	if colors, ok := oracleColors(keywords[0].Parameter); ok {
		return game.ProtectionFromColorsStaticAbility(colors...), true
	}
	protection, ok := parseProtectionParameter(keywords[0].Parameter)
	if !ok {
		return game.StaticAbility{}, false
	}
	return staticAbilityFromProtectionKeyword(protection, ""), true
}

func appendStaticRuleDeclaration(body *game.StaticAbility, declaration oracle.StaticDeclaration) bool {
	if declaration.Group.Domain != oracle.StaticGroupSource {
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

func staticRuleDomain(kind oracle.StaticRuleKind) oracle.StaticRuleDomain {
	switch kind {
	case oracle.StaticRuleMustAttack:
		return oracle.StaticRuleDomainAttack
	case oracle.StaticRuleCantBlock, oracle.StaticRuleCantBeBlocked:
		return oracle.StaticRuleDomainBlock
	case oracle.StaticRuleCantBeCountered:
		return oracle.StaticRuleDomainCountering
	default:
		return oracle.StaticRuleDomainUnknown
	}
}

func lowerStaticRuleKind(kind oracle.StaticRuleKind) (game.RuleEffectKind, bool) {
	switch kind {
	case oracle.StaticRuleCantBlock:
		return game.RuleEffectCantBlock, true
	case oracle.StaticRuleCantBeBlocked:
		return game.RuleEffectCantBeBlocked, true
	case oracle.StaticRuleMustAttack:
		return game.RuleEffectMustAttack, true
	case oracle.StaticRuleCantBeCountered:
		return game.RuleEffectCantBeCountered, true
	default:
		return game.RuleEffectNone, false
	}
}

func lowerStaticZone(value oracle.StaticZone) (zone.Type, bool) {
	switch value {
	case oracle.StaticZoneBattlefield:
		return zone.None, true
	case oracle.StaticZoneStack:
		return zone.Stack, true
	case oracle.StaticZoneHand:
		return zone.Hand, true
	default:
		return zone.None, false
	}
}

func appendStaticCostModifierDeclaration(body *game.StaticAbility, declaration oracle.StaticDeclaration) bool {
	if declaration.Group.Domain != oracle.StaticGroupControllerHandCards ||
		declaration.Cost.Kind != oracle.StaticCostModifierAbility ||
		declaration.Cost.AbilityKeyword != "Cycling" {
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

func appendStaticCardAbilityGrantDeclaration(body *game.StaticAbility, declaration oracle.StaticDeclaration) bool {
	if declaration.Group.Domain != oracle.StaticGroupControllerHandCards ||
		declaration.CardGrant.Keyword.Name != "Cycling" {
		return false
	}
	selection, ok := lowerStaticSelection(declaration.Group.Selection)
	if !ok || selection.Empty() {
		return false
	}
	manaCost, err := parseManaCostValue(declaration.CardGrant.Keyword.Parameter)
	if err != nil || len(manaCost) == 0 {
		return false
	}
	body.RuleEffects = append(body.RuleEffects, game.RuleEffect{
		Kind:           game.RuleEffectGrantHandCardAbility,
		AffectedPlayer: game.PlayerYou,
		CardSelection:  selection,
		GrantedAbility: game.CyclingActivatedAbility(manaCost),
	})
	return true
}

type loweredStaticGroupReference struct {
	Group          game.GroupReference
	AffectedSource bool
}

func lowerStaticGroupReference(reference oracle.StaticGroupReference) (loweredStaticGroupReference, bool) {
	selection, ok := lowerStaticSelection(reference.Selection)
	if !ok {
		return loweredStaticGroupReference{}, false
	}
	switch reference.Domain {
	case oracle.StaticGroupSource:
		if !selection.Empty() || reference.ExcludeSource {
			return loweredStaticGroupReference{}, false
		}
		return loweredStaticGroupReference{AffectedSource: true}, true
	case oracle.StaticGroupAttachedObject:
		if !selection.Empty() || reference.ExcludeSource {
			return loweredStaticGroupReference{}, false
		}
		return loweredStaticGroupReference{Group: game.AttachedObjectGroup(game.SourcePermanentReference())}, true
	case oracle.StaticGroupBattlefield:
		if reference.ExcludeSource {
			return loweredStaticGroupReference{
				Group: game.BattlefieldGroupExcluding(selection, game.SourcePermanentReference()),
			}, true
		}
		return loweredStaticGroupReference{Group: game.BattlefieldGroup(selection)}, true
	case oracle.StaticGroupSourceControllerPermanents:
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

func lowerStaticSelection(selection oracle.StaticSelection) (game.Selection, bool) {
	result := game.Selection{
		Controller: lowerStaticController(selection.Controller),
		TokenOnly:  selection.TokenOnly,
	}
	if selection.Controller != oracle.ControllerAny && result.Controller == game.ControllerAny {
		return game.Selection{}, false
	}
	for _, cardType := range selection.RequiredTypes {
		value, ok := lowerStaticCardType(cardType)
		if !ok {
			return game.Selection{}, false
		}
		result.RequiredTypes = append(result.RequiredTypes, value)
	}
	for _, subtypeText := range selection.SubtypesAny {
		subtype, ok := knownCreatureSubtypeFromPlural(subtypeText)
		if !ok {
			return game.Selection{}, false
		}
		result.SubtypesAny = append(result.SubtypesAny, subtype)
	}
	return result, len(result.Validate()) == 0
}

func lowerStaticController(controller oracle.ControllerKind) game.ControllerRelation {
	switch controller {
	case oracle.ControllerYou:
		return game.ControllerYou
	case oracle.ControllerOpponent:
		return game.ControllerOpponent
	case oracle.ControllerNotYou:
		return game.ControllerNotYou
	default:
		return game.ControllerAny
	}
}

func lowerStaticCardType(cardType oracle.StaticCardType) (types.Card, bool) {
	switch cardType {
	case oracle.StaticCardTypeArtifact:
		return types.Artifact, true
	case oracle.StaticCardTypeCreature:
		return types.Creature, true
	case oracle.StaticCardTypeLand:
		return types.Land, true
	default:
		return "", false
	}
}

func canonicalStaticDeclarationVarName(body *game.StaticAbility) string {
	switch body.Text {
	case game.CantBlockStaticBody.Text:
		return "game.CantBlockStaticBody"
	case game.CantBeBlockedStaticBody.Text:
		return "game.CantBeBlockedStaticBody"
	case game.MustAttackStaticBody.Text:
		return "game.MustAttackStaticBody"
	case game.CantBeCounteredStaticBody.Text:
		return "game.CantBeCounteredStaticBody"
	default:
		return ""
	}
}

func staticDeclarationDiagnostic(ability oracle.CompiledAbility, summary, detail string) *oracle.Diagnostic {
	return executableDiagnostic(ability, summary, detail)
}
