package cardgen

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/natefinch/council4/mtg/game"
)

func staticHintAt(hints faceRenderHints, i int) *staticVarHint {
	if i < len(hints.StaticVarNames) {
		return &hints.StaticVarNames[i]
	}
	return nil
}

func (r Renderer) renderStaticAbility(ctx *renderCtx, body *game.StaticAbility, hint *staticVarHint) (string, error) {
	if hint != nil && hint.VarName != "" {
		return hint.VarName, nil
	}
	if prot, ok := game.StaticBodyProtectionKeyword(body); ok {
		if s, err := r.renderProtectionStaticAbility(ctx, body, prot); s != "" || err != nil {
			return s, err
		}
	}
	if target, ok := game.StaticBodyEnchantTarget(body); ok &&
		reflect.DeepEqual(*body, game.EnchantStaticAbility(&target)) {
		renderedTarget, err := r.renderTargetSpec(ctx, &target)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EnchantStaticAbility(&%s)", renderedTarget), nil
	}
	if manaCost, ok := game.StaticBodyWardCost(body); ok &&
		reflect.DeepEqual(*body, game.WardStaticAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.WardStaticAbility(%s)", renderedCost), nil
	}
	var fields []string
	if body.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %s,", renderText(body.Text)))
	}
	if len(body.KeywordAbilities) > 0 {
		elements := make([]string, 0, len(body.KeywordAbilities))
		for _, keyword := range body.KeywordAbilities {
			rendered, err := r.renderKeywordAbility(ctx, keyword)
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("KeywordAbilities", "game.KeywordAbility", elements))
	}
	if body.Condition.Exists {
		rendered, err := r.renderStaticAbilityCondition(ctx, &body.Condition.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Condition: opt.Val(%s),", rendered))
	}
	if len(body.ContinuousEffects) > 0 {
		elements := make([]string, 0, len(body.ContinuousEffects))
		for i := range body.ContinuousEffects {
			rendered, err := r.renderContinuousEffect(ctx, &body.ContinuousEffects[i])
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("ContinuousEffects", "game.ContinuousEffect", elements))
	}
	if len(body.RuleEffects) > 0 {
		elements := make([]string, 0, len(body.RuleEffects))
		for i := range body.RuleEffects {
			rendered, err := r.renderRuleEffect(ctx, &body.RuleEffects[i])
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("RuleEffects", "game.RuleEffect", elements))
	}
	return structLit("game.StaticAbility", fields), nil
}

// renderProtectionStaticAbility renders a ProtectionKeyword static ability as
// a factory call if it matches the canonical factory form. Returns ("", nil)
// when the body does not match any canonical factory, leaving the caller to
// fall through to the generic struct-literal renderer.
func (Renderer) renderProtectionStaticAbility(ctx *renderCtx, body *game.StaticAbility, prot game.ProtectionKeyword) (string, error) {
	switch {
	case prot.Everything:
		if reflect.DeepEqual(*body, game.ProtectionFromEverythingStaticAbility()) {
			return "game.ProtectionFromEverythingStaticAbility()", nil
		}
	case prot.EachColor:
		if reflect.DeepEqual(*body, game.ProtectionFromEachColorStaticAbility()) {
			return "game.ProtectionFromEachColorStaticAbility()", nil
		}
	case prot.Multicolored:
		if reflect.DeepEqual(*body, game.ProtectionFromMulticoloredStaticAbility()) {
			return "game.ProtectionFromMulticoloredStaticAbility()", nil
		}
	case prot.Monocolored:
		if reflect.DeepEqual(*body, game.ProtectionFromMonocoloredStaticAbility()) {
			return "game.ProtectionFromMonocoloredStaticAbility()", nil
		}
	case len(prot.FromTypes) > 0:
		renderedTypes, err := renderCardTypeArguments(ctx, prot.FromTypes)
		if err != nil {
			return "", err
		}
		if reflect.DeepEqual(*body, game.ProtectionFromTypesStaticAbility(prot.FromTypes...)) {
			return fmt.Sprintf("game.ProtectionFromTypesStaticAbility(%s)", renderedTypes), nil
		}
	case len(prot.FromSubtypes) > 0:
		renderedSubtypes, err := renderSubtypeArguments(ctx, prot.FromSubtypes)
		if err != nil {
			return "", err
		}
		if reflect.DeepEqual(*body, game.ProtectionFromSubtypesStaticAbility(prot.FromSubtypes...)) {
			return fmt.Sprintf("game.ProtectionFromSubtypesStaticAbility(%s)", renderedSubtypes), nil
		}
	case len(prot.FromColors) > 0:
		renderedColors, err := renderColorArguments(ctx, prot.FromColors)
		if err != nil {
			return "", err
		}
		if reflect.DeepEqual(*body, game.ProtectionFromColorsStaticAbility(prot.FromColors...)) {
			return fmt.Sprintf("game.ProtectionFromColorsStaticAbility(%s)", renderedColors), nil
		}
	default:
		// Unknown predicate combination — fall through to generic rendering.
	}
	return "", nil
}

func (r Renderer) renderContinuousEffect(ctx *renderCtx, effect *game.ContinuousEffect) (string, error) {
	var fields []string
	if len(effect.RemoveKeywords) > 0 {
		return "", errors.New("render: unsupported ability-layer continuous effect fields")
	}
	if effect.AffectedSource && !effect.Group.Empty() {
		return "", errors.New("render: continuous effect cannot set both AffectedSource and Group")
	}
	switch effect.Layer {
	case game.LayerControl:
		if effect.PowerDelta != 0 ||
			effect.ToughnessDelta != 0 ||
			effect.PowerDeltaDynamic.Exists ||
			effect.ToughnessDeltaDynamic.Exists {
			return "", errors.New("render: power/toughness fields require a power/toughness layer")
		}
		if len(effect.AddKeywords) > 0 {
			return "", errors.New("render: keyword fields require the ability layer")
		}
	case game.LayerAbility:
		if effect.PowerDelta != 0 ||
			effect.ToughnessDelta != 0 ||
			effect.PowerDeltaDynamic.Exists ||
			effect.ToughnessDeltaDynamic.Exists {
			return "", errors.New("render: power/toughness fields require a power/toughness layer")
		}
	case game.LayerPowerToughnessModify:
		if len(effect.AddKeywords) > 0 {
			return "", errors.New("render: keyword fields require the ability layer")
		}
	default:
	}
	layerLit, err := renderContinuousLayer(effect.Layer)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Layer: %s,", layerLit))
	if effect.Layer == game.LayerControl && effect.NewController.Exists {
		ctx.need(importOpt)
		fields = append(fields, "NewController: opt.Val(game.Player1),")
	}
	if effect.AffectedSource {
		fields = append(fields, "AffectedSource: true,")
	}
	if effect.Group.Valid() {
		groupLit, err := r.renderGroupReference(ctx, effect.Group)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Group: %s,", groupLit))
	}
	if effect.PowerDelta != 0 {
		fields = append(fields, fmt.Sprintf("PowerDelta: %d,", effect.PowerDelta))
	}
	if effect.ToughnessDelta != 0 {
		fields = append(fields, fmt.Sprintf("ToughnessDelta: %d,", effect.ToughnessDelta))
	}
	if effect.PowerDeltaDynamic.Exists {
		dynamic, err := r.renderDynamicAmount(ctx, effect.PowerDeltaDynamic.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("PowerDeltaDynamic: opt.Val(%s),", dynamic))
	}
	if effect.ToughnessDeltaDynamic.Exists {
		dynamic, err := r.renderDynamicAmount(ctx, effect.ToughnessDeltaDynamic.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ToughnessDeltaDynamic: opt.Val(%s),", dynamic))
	}
	if len(effect.AddKeywords) > 0 {
		elements := make([]string, 0, len(effect.AddKeywords))
		for _, keyword := range effect.AddKeywords {
			literal, err := renderKeyword(keyword)
			if err != nil {
				return "", err
			}
			elements = append(elements, literal+",")
		}
		fields = append(fields, sliceField("AddKeywords", "game.Keyword", elements))
	}
	if len(effect.AddAbilities) > 0 {
		elements := make([]string, 0, len(effect.AddAbilities))
		for _, ability := range effect.AddAbilities {
			staticBody, ok := ability.(game.StaticAbility)
			if !ok {
				return "", fmt.Errorf("render: AddAbilities element is not a StaticAbility: %T", ability)
			}
			rendered, err := r.renderStaticAbility(ctx, &staticBody, nil)
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("AddAbilities", "game.Ability", elements))
	}
	return structLit("game.ContinuousEffect", fields), nil
}

func renderContinuousLayer(layer game.ContinuousLayer) (string, error) {
	switch layer {
	case game.LayerControl:
		return "game.LayerControl", nil
	case game.LayerAbility:
		return "game.LayerAbility", nil
	case game.LayerPowerToughnessModify:
		return "game.LayerPowerToughnessModify", nil
	default:
		return "", fmt.Errorf("render: unsupported continuous layer %d", layer)
	}
}

func (r Renderer) renderRuleEffect(ctx *renderCtx, effect *game.RuleEffect) (string, error) {
	kind, err := renderRuleEffectKind(effect.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if effect.AffectedSource {
		fields = append(fields, "AffectedSource: true,")
	}
	if effect.AffectedPlayer != game.PlayerAny {
		player, err := renderPlayerRelation(effect.AffectedPlayer)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AffectedPlayer: %s,", player))
	}
	if !effect.CardSelection.Empty() {
		selection, err := r.renderSelection(ctx, effect.CardSelection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CardSelection: %s,", selection))
	}
	if effect.Kind == game.RuleEffectGrantHandCardAbility {
		if !game.BodyHasKeyword(effect.GrantedAbility, game.Cycling) {
			return "", errors.New("render: hand-card ability grant must grant Cycling")
		}
		ability, err := r.renderActivatedAbility(ctx, &effect.GrantedAbility)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("GrantedAbility: %s,", ability))
	}
	if effect.Kind == game.RuleEffectCostModifier {
		modifier, err := r.renderCostModifier(ctx, effect.CostModifier)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CostModifier: %s,", modifier))
	}
	return structLit("game.RuleEffect", fields), nil
}

func renderRuleEffectKind(kind game.RuleEffectKind) (string, error) {
	switch kind {
	case game.RuleEffectCantBlock:
		return "game.RuleEffectCantBlock", nil
	case game.RuleEffectCantBeCountered:
		return "game.RuleEffectCantBeCountered", nil
	case game.RuleEffectCantBeBlocked:
		return "game.RuleEffectCantBeBlocked", nil
	case game.RuleEffectMustAttack:
		return "game.RuleEffectMustAttack", nil
	case game.RuleEffectCostModifier:
		return "game.RuleEffectCostModifier", nil
	case game.RuleEffectGrantHandCardAbility:
		return "game.RuleEffectGrantHandCardAbility", nil
	default:
		return "", fmt.Errorf("render: unsupported rule effect kind %d", kind)
	}
}

func (r Renderer) renderCostModifier(ctx *renderCtx, modifier game.CostModifier) (string, error) {
	kind, err := renderCostModifierKind(modifier.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if modifier.MatchCardType {
		cardType, err := cardTypeLiteral(modifier.CardType)
		if err != nil {
			return "", err
		}
		fields = append(fields, "MatchCardType: true,", fmt.Sprintf("CardType: %s,", cardType))
	}
	if modifier.AbilityKeyword != game.KeywordNone {
		keyword, err := renderKeyword(modifier.AbilityKeyword)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AbilityKeyword: %s,", keyword))
	}
	if modifier.GenericIncrease != 0 {
		fields = append(fields, fmt.Sprintf("GenericIncrease: %d,", modifier.GenericIncrease))
	}
	if modifier.GenericReduction != 0 {
		fields = append(fields, fmt.Sprintf("GenericReduction: %d,", modifier.GenericReduction))
	}
	if modifier.SetGeneric.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SetGeneric: opt.Val(%d),", modifier.SetGeneric.Val))
	}
	if modifier.SetManaCost.Exists {
		ctx.need(importOpt)
		manaCost, err := r.renderManaCost(ctx, modifier.SetManaCost.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SetManaCost: opt.Val(%s),", manaCost))
	}
	if modifier.MinimumGeneric != 0 {
		fields = append(fields, fmt.Sprintf("MinimumGeneric: %d,", modifier.MinimumGeneric))
	}
	if modifier.FirstCycleEachTurn {
		fields = append(fields, "FirstCycleEachTurn: true,")
	}
	return structLit("game.CostModifier", fields), nil
}

func renderCostModifierKind(kind game.CostModifierKind) (string, error) {
	switch kind {
	case game.CostModifierSpell:
		return "game.CostModifierSpell", nil
	case game.CostModifierAbility:
		return "game.CostModifierAbility", nil
	case game.CostModifierAttack:
		return "game.CostModifierAttack", nil
	default:
		return "", fmt.Errorf("render: unsupported cost modifier kind %d", kind)
	}
}
