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

// renderStaticAbility renders a static ability, honouring a caller-supplied
// hint that the ability is already published as a package-level var. Without a
// hint it is exactly the generated renderer.
func (r Renderer) renderStaticAbility(ctx *renderCtx, body *game.StaticAbility, hint *staticVarHint) (string, error) {
	if hint != nil && hint.VarName != "" {
		return hint.VarName, nil
	}
	return r.renderGameStaticAbility(ctx, *body)
}

// constructorStaticAbility spells a static ability as a game constructor call
// when one produces exactly this value. See constructorRenderers in
// cardgen/cmd/genrender for why this hangs off the generated renderer.
func (r Renderer) constructorStaticAbility(ctx *renderCtx, v game.StaticAbility) (string, bool, error) {
	body := &v
	if prot, ok := game.StaticBodyProtectionKeyword(body); ok {
		if s, err := r.renderProtectionStaticAbility(ctx, body, prot); s != "" || err != nil {
			return s, true, err
		}
	}
	if hexproof, ok := game.StaticBodyHexproofFromKeyword(body); ok && len(hexproof.FromColors) > 0 &&
		reflect.DeepEqual(*body, game.HexproofFromColorsStaticAbility(hexproof.FromColors...)) {
		renderedColors, err := renderColorArguments(ctx, hexproof.FromColors)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.HexproofFromColorsStaticAbility(%s)", renderedColors), true, nil
	}
	if enchant, ok := game.StaticBodyEnchantKeyword(body); ok && enchant.Reanimates &&
		reflect.DeepEqual(*body, game.ReanimationEnchantStaticAbility(&enchant.Target)) {
		renderedTarget, err := r.renderTargetSpec(ctx, &enchant.Target)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.ReanimationEnchantStaticAbility(&%s)", renderedTarget), true, nil
	}
	if target, ok := game.StaticBodyEnchantTarget(body); ok &&
		reflect.DeepEqual(*body, game.EnchantStaticAbility(&target)) {
		renderedTarget, err := r.renderTargetSpec(ctx, &target)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.EnchantStaticAbility(&%s)", renderedTarget), true, nil
	}
	if bestow, ok := game.StaticBodyBestow(body); ok &&
		reflect.DeepEqual(*body, game.BestowStaticAbility(bestow.Cost, &bestow.Target)) {
		renderedMana, err := r.renderManaCost(ctx, bestow.Cost)
		if err != nil {
			return "", false, err
		}
		renderedTarget, err := r.renderTargetSpec(ctx, &bestow.Target)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.BestowStaticAbility(%s, &%s)", renderedMana, renderedTarget), true, nil
	}
	if offspring, ok := game.StaticBodyOffspring(body); ok &&
		reflect.DeepEqual(*body, game.OffspringStaticAbility(offspring.Cost)) {
		renderedMana, err := r.renderManaCost(ctx, offspring.Cost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.OffspringStaticAbility(%s)", renderedMana), true, nil
	}
	if manaCost, additionalCosts, ok := game.StaticBodyWardCosts(body); ok &&
		len(additionalCosts) > 0 &&
		reflect.DeepEqual(*body, game.WardStaticAbilityWithCosts(manaCost, additionalCosts)) {
		renderedMana, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		renderedAdditional, err := r.renderAdditionalCosts(ctx, additionalCosts)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.WardStaticAbilityWithCosts(%s, %s)", renderedMana, renderedAdditional), true, nil
	}
	if manaCost, ok := game.StaticBodyWardCost(body); ok &&
		reflect.DeepEqual(*body, game.WardStaticAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.WardStaticAbility(%s)", renderedCost), true, nil
	}
	if count, ok := game.StaticBodyDredgeCount(body); ok &&
		reflect.DeepEqual(*body, game.DredgeStaticAbility(count)) {
		return fmt.Sprintf("game.DredgeStaticAbility(%d)", count), true, nil
	}
	return "", false, nil
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
	case prot.ChosenColor:
		if reflect.DeepEqual(*body, game.ProtectionFromChosenColorStaticAbility()) {
			return "game.ProtectionFromChosenColorStaticAbility()", nil
		}
	case prot.CommanderIdentityComplement:
		if reflect.DeepEqual(*body, game.ProtectionFromNonCommanderIdentityColorsStaticAbility()) {
			return "game.ProtectionFromNonCommanderIdentityColorsStaticAbility()", nil
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
	return r.renderGameContinuousEffect(ctx, *effect)
}

// validateContinuousEffectLayerFields fails closed when an effect carries fields
// that do not belong to its layer, keeping rendering layer-faithful.
func validateContinuousEffectLayerFields(effect *game.ContinuousEffect) error {
	hasPTDelta := effect.PowerDelta != 0 ||
		effect.ToughnessDelta != 0 ||
		effect.PowerDeltaDynamic.Exists ||
		effect.ToughnessDeltaDynamic.Exists
	hasKeywords := len(effect.AddKeywords) > 0 || len(effect.RemoveKeywords) > 0
	keywordOnAbility := errors.New("render: keyword fields require the ability layer")
	ptOnNonPT := errors.New("render: power/toughness fields require a power/toughness layer")
	switch effect.Layer {
	case game.LayerControl:
		if hasPTDelta {
			return ptOnNonPT
		}
		if hasKeywords {
			return keywordOnAbility
		}
	case game.LayerAbility:
		if hasPTDelta {
			return ptOnNonPT
		}
		if effect.RemoveAllAbilities &&
			(len(effect.AddKeywords) > 0 || len(effect.AddAbilities) > 0) {
			return errors.New("render: remove-all-abilities effect cannot also add abilities or keywords")
		}
	case game.LayerPowerToughnessModify:
		if hasKeywords {
			return keywordOnAbility
		}
	case game.LayerPowerToughnessSet:
		if hasKeywords {
			return keywordOnAbility
		}
		if hasPTDelta {
			return errors.New("render: power/toughness delta fields require the modify layer")
		}
		setsPower := effect.SetPower.Exists || effect.SetPowerDynamic.Exists
		setsToughness := effect.SetToughness.Exists || effect.SetToughnessDynamic.Exists
		if !setsPower || !setsToughness {
			return errors.New("render: base power/toughness layer requires set power and toughness")
		}
		if effect.SetPower.Exists && effect.SetPowerDynamic.Exists {
			return errors.New("render: base power layer cannot set both a fixed and dynamic power")
		}
		if effect.SetToughness.Exists && effect.SetToughnessDynamic.Exists {
			return errors.New("render: base toughness layer cannot set both a fixed and dynamic toughness")
		}
	case game.LayerColor:
		if hasKeywords {
			return keywordOnAbility
		}
		if len(effect.SetColors) == 0 && len(effect.AddColors) == 0 && !effect.SetColorless {
			return errors.New("render: color layer requires set or add colors")
		}
		if len(effect.SetColors) > 0 && len(effect.AddColors) > 0 {
			return errors.New("render: color layer cannot both set and add colors")
		}
		if effect.SetColorless && (len(effect.SetColors) > 0 || len(effect.AddColors) > 0) {
			return errors.New("render: colorless set cannot also set or add colors")
		}
	case game.LayerType:
		if hasKeywords {
			return keywordOnAbility
		}
		if len(effect.AddTypes) == 0 && len(effect.AddSubtypes) == 0 &&
			len(effect.SetTypes) == 0 && len(effect.SetSubtypes) == 0 &&
			len(effect.AddSupertypes) == 0 && len(effect.SetSupertypes) == 0 &&
			len(effect.RemoveTypes) == 0 && len(effect.RemoveSubtypes) == 0 &&
			len(effect.RemoveSupertypes) == 0 &&
			effect.AddSubtypeFromEntryChoice == "" && effect.SetSubtypeFromSourceChoice == "" &&
			!effect.AddEveryCreatureType && !effect.AddEveryBasicLandType {
			return errors.New("render: type layer requires set, added, or removed types or subtypes")
		}
	case game.LayerText:
		if hasPTDelta {
			return ptOnNonPT
		}
		if hasKeywords {
			return keywordOnAbility
		}
		if effect.SetName == "" && effect.SetNameFromSourceChoice == "" && effect.TextFrom == "" && effect.TextTo == "" {
			return errors.New("render: text layer requires a name or text change")
		}
	case game.LayerPowerToughnessSwitch:
		if hasPTDelta {
			return errors.New("render: power/toughness delta fields require the modify layer")
		}
		if hasKeywords {
			return keywordOnAbility
		}
		if effect.SetPower.Exists || effect.SetToughness.Exists ||
			effect.SetPowerDynamic.Exists || effect.SetToughnessDynamic.Exists {
			return errors.New("render: power/toughness switch layer cannot set power or toughness")
		}
	default:
	}
	return nil
}

func renderContinuousLayer(layer game.ContinuousLayer) (string, error) {
	switch layer {
	case game.LayerControl:
		return "game.LayerControl", nil
	case game.LayerText:
		return "game.LayerText", nil
	case game.LayerAbility:
		return "game.LayerAbility", nil
	case game.LayerPowerToughnessModify:
		return "game.LayerPowerToughnessModify", nil
	case game.LayerPowerToughnessSet:
		return "game.LayerPowerToughnessSet", nil
	case game.LayerPowerToughnessSwitch:
		return "game.LayerPowerToughnessSwitch", nil
	case game.LayerColor:
		return "game.LayerColor", nil
	case game.LayerType:
		return "game.LayerType", nil
	default:
		return "", fmt.Errorf("render: unsupported continuous layer %d", layer)
	}
}

func (r Renderer) renderRuleEffect(ctx *renderCtx, effect *game.RuleEffect) (string, error) {
	return r.renderGameRuleEffect(ctx, *effect)
}

func renderRuleEffectKind(kind game.RuleEffectKind) (string, error) {
	switch kind {
	case game.RuleEffectCantBlock:
		return "game.RuleEffectCantBlock", nil
	case game.RuleEffectCantAttack:
		return "game.RuleEffectCantAttack", nil
	case game.RuleEffectCantBeCountered:
		return "game.RuleEffectCantBeCountered", nil
	case game.RuleEffectCantBeBlocked:
		return "game.RuleEffectCantBeBlocked", nil
	case game.RuleEffectCantBeBlockedByCreaturesWith:
		return "game.RuleEffectCantBeBlockedByCreaturesWith", nil
	case game.RuleEffectCantBeBlockedExceptBy:
		return "game.RuleEffectCantBeBlockedExceptBy", nil
	case game.RuleEffectCantBeTargetedByControllerOpponents:
		return "game.RuleEffectCantBeTargetedByControllerOpponents", nil
	case game.RuleEffectSkipExtraTurns:
		return "game.RuleEffectSkipExtraTurns", nil
	case game.RuleEffectAssignCombatDamageUsingToughness:
		return "game.RuleEffectAssignCombatDamageUsingToughness", nil
	case game.RuleEffectCanBlockOnlyCreaturesWith:
		return "game.RuleEffectCanBlockOnlyCreaturesWith", nil
	case game.RuleEffectCanBlockAdditional:
		return "game.RuleEffectCanBlockAdditional", nil
	case game.RuleEffectDamageDoesntCauseLifeLoss:
		return "game.RuleEffectDamageDoesntCauseLifeLoss", nil
	case game.RuleEffectRedirectDamageToSource:
		return "game.RuleEffectRedirectDamageToSource", nil
	case game.RuleEffectCantBeBlockedByMoreThanOne:
		return "game.RuleEffectCantBeBlockedByMoreThanOne", nil
	case game.RuleEffectMustAttack:
		return "game.RuleEffectMustAttack", nil
	case game.RuleEffectCanAttackAsThoughDefender:
		return "game.RuleEffectCanAttackAsThoughDefender", nil
	case game.RuleEffectMustBeBlocked:
		return "game.RuleEffectMustBeBlocked", nil
	case game.RuleEffectMustBeBlockedByAllAble:
		return "game.RuleEffectMustBeBlockedByAllAble", nil
	case game.RuleEffectAssignCombatDamageAsThoughUnblocked:
		return "game.RuleEffectAssignCombatDamageAsThoughUnblocked", nil
	case game.RuleEffectDoesntUntap:
		return "game.RuleEffectDoesntUntap", nil
	case game.RuleEffectCantTransform:
		return "game.RuleEffectCantTransform", nil
	case game.RuleEffectCostModifier:
		return "game.RuleEffectCostModifier", nil
	case game.RuleEffectGrantHandCardAbility:
		return "game.RuleEffectGrantHandCardAbility", nil
	case game.RuleEffectGrantGraveyardCardKeyword:
		return "game.RuleEffectGrantGraveyardCardKeyword", nil
	case game.RuleEffectGrantSpellKeyword:
		return "game.RuleEffectGrantSpellKeyword", nil
	case game.RuleEffectAscend:
		return "game.RuleEffectAscend", nil
	case game.RuleEffectPlayerProtection:
		return "game.RuleEffectPlayerProtection", nil
	case game.RuleEffectPlayerHexproof:
		return "game.RuleEffectPlayerHexproof", nil
	case game.RuleEffectPlayerShroud:
		return "game.RuleEffectPlayerShroud", nil
	case game.RuleEffectAttackTax:
		return "game.RuleEffectAttackTax", nil
	case game.RuleEffectLifeTotalCantChange:
		return "game.RuleEffectLifeTotalCantChange", nil
	case game.RuleEffectAdditionalTriggerForChosenCreatureType:
		return "game.RuleEffectAdditionalTriggerForChosenCreatureType", nil
	case game.RuleEffectAdditionalLandPlays:
		return "game.RuleEffectAdditionalLandPlays", nil
	case game.RuleEffectCantCastSpells:
		return "game.RuleEffectCantCastSpells", nil
	case game.RuleEffectCantCastFromZones:
		return "game.RuleEffectCantCastFromZones", nil
	case game.RuleEffectCantEnterFromZones:
		return "game.RuleEffectCantEnterFromZones", nil
	case game.RuleEffectCantActivateAbilities:
		return "game.RuleEffectCantActivateAbilities", nil
	case game.RuleEffectCantActivateAbilitiesOfPermanent:
		return "game.RuleEffectCantActivateAbilitiesOfPermanent", nil
	case game.RuleEffectAdditionalTriggerForEnteringPermanent:
		return "game.RuleEffectAdditionalTriggerForEnteringPermanent", nil
	case game.RuleEffectAdditionalTriggerForControlledPermanent:
		return "game.RuleEffectAdditionalTriggerForControlledPermanent", nil
	case game.RuleEffectAdditionalTriggerForRoomAbility:
		return "game.RuleEffectAdditionalTriggerForRoomAbility", nil
	case game.RuleEffectSuppressOpponentEnteringTriggers:
		return "game.RuleEffectSuppressOpponentEnteringTriggers", nil
	case game.RuleEffectControlOpponentSearches:
		return "game.RuleEffectControlOpponentSearches", nil
	case game.RuleEffectExileOpponentSearchFinds:
		return "game.RuleEffectExileOpponentSearchFinds", nil
	case game.RuleEffectAttackTaxPerCreature:
		return "game.RuleEffectAttackTaxPerCreature", nil
	case game.RuleEffectManaProductionMultiplier:
		return "game.RuleEffectManaProductionMultiplier", nil
	case game.RuleEffectUntapDuringOtherPlayersUntapStep:
		return "game.RuleEffectUntapDuringOtherPlayersUntapStep", nil
	case game.RuleEffectCastSpellsAsThoughFlash:
		return "game.RuleEffectCastSpellsAsThoughFlash", nil
	case game.RuleEffectPlayLandsFromZone:
		return "game.RuleEffectPlayLandsFromZone", nil
	case game.RuleEffectPlayWithTopCardRevealed:
		return "game.RuleEffectPlayWithTopCardRevealed", nil
	case game.RuleEffectLookAtTopCardAnyTime:
		return "game.RuleEffectLookAtTopCardAnyTime", nil
	case game.RuleEffectCastSpellsFromZone:
		return "game.RuleEffectCastSpellsFromZone", nil
	case game.RuleEffectCastFromZone:
		return "game.RuleEffectCastFromZone", nil
	case game.RuleEffectNoMaximumHandSize:
		return "game.RuleEffectNoMaximumHandSize", nil
	case game.RuleEffectLegendRuleDoesNotApply:
		return "game.RuleEffectLegendRuleDoesNotApply", nil
	case game.RuleEffectSkipDrawStep:
		return "game.RuleEffectSkipDrawStep", nil
	case game.RuleEffectPayLifeForColoredMana:
		return "game.RuleEffectPayLifeForColoredMana", nil
	case game.RuleEffectPayLifeForCommanderTax:
		return "game.RuleEffectPayLifeForCommanderTax", nil
	case game.RuleEffectDrawLimitPerTurn:
		return "game.RuleEffectDrawLimitPerTurn", nil
	case game.RuleEffectCastLimitPerTurn:
		return "game.RuleEffectCastLimitPerTurn", nil
	case game.RuleEffectGoaded:
		return "game.RuleEffectGoaded", nil
	case game.RuleEffectCantBeSacrificed:
		return "game.RuleEffectCantBeSacrificed", nil
	case game.RuleEffectCastLinkedExileForFree:
		return "game.RuleEffectCastLinkedExileForFree", nil
	case game.RuleEffectCombatDamageCantBePrevented:
		return "game.RuleEffectCombatDamageCantBePrevented", nil
	default:
		return "", fmt.Errorf("render: unsupported rule effect kind %d", kind)
	}
}

func (r Renderer) renderCostModifier(ctx *renderCtx, modifier game.CostModifier) (string, error) {
	return r.renderGameCostModifier(ctx, modifier)
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
