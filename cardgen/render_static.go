package cardgen

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
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
	if count, ok := game.StaticBodyDredgeCount(body); ok &&
		reflect.DeepEqual(*body, game.DredgeStaticAbility(count)) {
		return fmt.Sprintf("game.DredgeStaticAbility(%d)", count), nil
	}
	var fields []string
	if body.ZoneOfFunction != zone.None {
		ctx.need(importZone)
		zoneLiteral, err := renderZone(body.ZoneOfFunction)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ZoneOfFunction: %s,", zoneLiteral))
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
	case prot.ChosenColor:
		if reflect.DeepEqual(*body, game.ProtectionFromChosenColorStaticAbility()) {
			return "game.ProtectionFromChosenColorStaticAbility()", nil
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
	if effect.AffectedSource && !effect.Group.Empty() {
		return "", errors.New("render: continuous effect cannot set both AffectedSource and Group")
	}
	if err := validateContinuousEffectLayerFields(effect); err != nil {
		return "", err
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
	powerToughnessFields, err := r.renderContinuousPowerToughnessFields(ctx, effect)
	if err != nil {
		return "", err
	}
	fields = append(fields, powerToughnessFields...)
	characteristicFields, err := renderContinuousCharacteristicFields(ctx, effect)
	if err != nil {
		return "", err
	}
	fields = append(fields, characteristicFields...)
	abilityFields, err := r.renderContinuousAbilityFields(ctx, effect)
	if err != nil {
		return "", err
	}
	fields = append(fields, abilityFields...)
	return structLit("game.ContinuousEffect", fields), nil
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
		if !effect.SetPower.Exists || !effect.SetToughness.Exists {
			return errors.New("render: base power/toughness layer requires set power and toughness")
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
			effect.AddSubtypeFromEntryChoice == "" {
			return errors.New("render: type layer requires set or added types or subtypes")
		}
	default:
	}
	return nil
}

// renderContinuousPowerToughnessFields renders the power/toughness delta fields
// in canonical order.
func (r Renderer) renderContinuousPowerToughnessFields(ctx *renderCtx, effect *game.ContinuousEffect) ([]string, error) {
	var fields []string
	if effect.PowerDelta != 0 {
		fields = append(fields, fmt.Sprintf("PowerDelta: %d,", effect.PowerDelta))
	}
	if effect.ToughnessDelta != 0 {
		fields = append(fields, fmt.Sprintf("ToughnessDelta: %d,", effect.ToughnessDelta))
	}
	if effect.PowerDeltaDynamic.Exists {
		dynamic, err := r.renderDynamicAmount(ctx, &effect.PowerDeltaDynamic.Val)
		if err != nil {
			return nil, err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("PowerDeltaDynamic: opt.Val(%s),", dynamic))
	}
	if effect.ToughnessDeltaDynamic.Exists {
		dynamic, err := r.renderDynamicAmount(ctx, &effect.ToughnessDeltaDynamic.Val)
		if err != nil {
			return nil, err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ToughnessDeltaDynamic: opt.Val(%s),", dynamic))
	}
	return fields, nil
}

// renderContinuousCharacteristicFields renders the base power/toughness, color,
// and type characteristic fields in canonical order.
func renderContinuousCharacteristicFields(ctx *renderCtx, effect *game.ContinuousEffect) ([]string, error) {
	var fields []string
	if effect.SetPower.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SetPower: opt.Val(%s),", renderPTValue(effect.SetPower.Val)))
	}
	if effect.SetToughness.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("SetToughness: opt.Val(%s),", renderPTValue(effect.SetToughness.Val)))
	}
	if len(effect.SetColors) > 0 {
		literals, err := colorValueLiterals(effect.SetColors)
		if err != nil {
			return nil, err
		}
		ctx.need(importColor)
		fields = append(fields, fmt.Sprintf("SetColors: []color.Color{%s},", literals))
	}
	if len(effect.AddColors) > 0 {
		literals, err := colorValueLiterals(effect.AddColors)
		if err != nil {
			return nil, err
		}
		ctx.need(importColor)
		fields = append(fields, fmt.Sprintf("AddColors: []color.Color{%s},", literals))
	}
	if effect.SetColorless {
		fields = append(fields, "SetColorless: true,")
	}
	if len(effect.SetTypes) > 0 {
		literal, err := renderTypesCardSlice(ctx, effect.SetTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("SetTypes: %s,", literal))
	}
	if len(effect.AddTypes) > 0 {
		literal, err := renderTypesCardSlice(ctx, effect.AddTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AddTypes: %s,", literal))
	}
	if len(effect.SetSubtypes) > 0 {
		ctx.need(importTypes)
		cardTypeStrings := make([]string, 0, len(effect.SetTypes))
		for _, t := range effect.SetTypes {
			cardTypeStrings = append(cardTypeStrings, string(t))
		}
		literals := make([]string, 0, len(effect.SetSubtypes))
		for _, sub := range effect.SetSubtypes {
			literals = append(literals, SubtypeToLiteral(string(sub), cardTypeStrings))
		}
		fields = append(fields, fmt.Sprintf("SetSubtypes: []types.Sub{%s},", strings.Join(literals, ", ")))
	}
	if len(effect.AddSubtypes) > 0 {
		ctx.need(importTypes)
		cardTypeStrings := make([]string, 0, len(effect.AddTypes))
		for _, t := range effect.AddTypes {
			cardTypeStrings = append(cardTypeStrings, string(t))
		}
		literals := make([]string, 0, len(effect.AddSubtypes))
		for _, sub := range effect.AddSubtypes {
			literals = append(literals, SubtypeToLiteral(string(sub), cardTypeStrings))
		}
		fields = append(fields, fmt.Sprintf("AddSubtypes: []types.Sub{%s},", strings.Join(literals, ", ")))
	}
	if effect.AddSubtypeFromEntryChoice != "" {
		if effect.AddSubtypeFromEntryChoice != game.EntryTypeChoiceKey {
			return nil, errors.New("render: unsupported entry-choice subtype key")
		}
		fields = append(fields, "AddSubtypeFromEntryChoice: game.EntryTypeChoiceKey,")
	}
	return fields, nil
}

// renderContinuousAbilityFields renders the granted keyword and ability fields.
func (r Renderer) renderContinuousAbilityFields(ctx *renderCtx, effect *game.ContinuousEffect) ([]string, error) {
	var fields []string
	if effect.RemoveAllAbilities {
		fields = append(fields, "RemoveAllAbilities: true,")
	}
	if len(effect.AddKeywords) > 0 {
		elements := make([]string, 0, len(effect.AddKeywords))
		for _, keyword := range effect.AddKeywords {
			literal, err := renderKeyword(keyword)
			if err != nil {
				return nil, err
			}
			elements = append(elements, literal+",")
		}
		fields = append(fields, sliceField("AddKeywords", "game.Keyword", elements))
	}
	if len(effect.RemoveKeywords) > 0 {
		elements := make([]string, 0, len(effect.RemoveKeywords))
		for _, keyword := range effect.RemoveKeywords {
			literal, err := renderKeyword(keyword)
			if err != nil {
				return nil, err
			}
			elements = append(elements, literal+",")
		}
		fields = append(fields, sliceField("RemoveKeywords", "game.Keyword", elements))
	}
	if len(effect.AddAbilities) > 0 {
		elements := make([]string, 0, len(effect.AddAbilities))
		for _, ability := range effect.AddAbilities {
			var rendered string
			var err error
			switch body := ability.(type) {
			case *game.StaticAbility:
				rendered, err = r.renderStaticAbility(ctx, body, nil)
			case *game.ManaAbility:
				rendered, err = r.renderManaAbility(ctx, body)
			case *game.TriggeredAbility:
				rendered, err = r.renderTriggeredAbility(ctx, body)
			case *game.ActivatedAbility:
				rendered, err = r.renderActivatedAbility(ctx, body)
			default:
				return nil, fmt.Errorf("render: unsupported AddAbilities element: %T", ability)
			}
			if err != nil {
				return nil, err
			}
			// AddAbilities is []game.Ability and Ability is implemented on
			// pointer receivers, so each element must be a pointer. new(expr)
			// addresses the rendered value, which works whether it renders as a
			// composite literal or a helper function call (unlike &expr, which
			// cannot address a function call result).
			elements = append(elements, "new("+rendered+"),")
		}
		fields = append(fields, sliceField("AddAbilities", "game.Ability", elements))
	}
	return fields, nil
}

func renderContinuousLayer(layer game.ContinuousLayer) (string, error) {
	switch layer {
	case game.LayerControl:
		return "game.LayerControl", nil
	case game.LayerAbility:
		return "game.LayerAbility", nil
	case game.LayerPowerToughnessModify:
		return "game.LayerPowerToughnessModify", nil
	case game.LayerPowerToughnessSet:
		return "game.LayerPowerToughnessSet", nil
	case game.LayerColor:
		return "game.LayerColor", nil
	case game.LayerType:
		return "game.LayerType", nil
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
	if effect.AffectedAttached {
		fields = append(fields, "AffectedAttached: true,")
	}
	if effect.AffectedPlayer != game.PlayerAny {
		player, err := renderPlayerRelation(effect.AffectedPlayer)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AffectedPlayer: %s,", player))
	}
	if effect.AffectedController != game.ControllerAny {
		controller, err := renderControllerRelation(effect.AffectedController)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AffectedController: %s,", controller))
	}
	if effect.DefendingPlayer != game.PlayerAny {
		player, err := renderPlayerRelation(effect.DefendingPlayer)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("DefendingPlayer: %s,", player))
	}
	if !effect.CardSelection.Empty() {
		selection, err := r.renderSelection(ctx, effect.CardSelection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CardSelection: %s,", selection))
	}
	if effect.Kind == game.RuleEffectGrantHandCardAbility {
		if !game.BodyHasKeyword(&effect.GrantedAbility, game.Cycling) {
			return "", errors.New("render: hand-card ability grant must grant Cycling")
		}
		ability, err := r.renderActivatedAbility(ctx, &effect.GrantedAbility)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("GrantedAbility: %s,", ability))
	}
	if effect.Kind == game.RuleEffectCantBeBlockedByCreaturesWith {
		restriction, err := renderBlockerRestriction(effect.BlockerRestriction)
		if err != nil {
			return "", err
		}
		if effect.BlockerRestriction.Color != "" {
			ctx.need(importColor)
		}
		fields = append(fields, fmt.Sprintf("BlockerRestriction: %s,", restriction))
	}
	if effect.Kind == game.RuleEffectCostModifier {
		modifier, err := r.renderCostModifier(ctx, effect.CostModifier)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CostModifier: %s,", modifier))
	}
	if effect.Kind == game.RuleEffectPlayerProtection {
		if !effect.Protection.Everything ||
			len(effect.Protection.FromColors) != 0 ||
			len(effect.Protection.FromTypes) != 0 ||
			len(effect.Protection.FromSubtypes) != 0 ||
			effect.Protection.Multicolored ||
			effect.Protection.Monocolored ||
			effect.Protection.EachColor {
			return "", errors.New("render: player protection supports only protection from everything")
		}
		fields = append(fields, "Protection: game.ProtectionKeyword{Everything: true},")
	}
	if effect.Kind == game.RuleEffectAttackTax {
		if effect.AttackTaxGeneric <= 0 {
			return "", errors.New("render: attack tax requires a positive generic amount")
		}
		fields = append(fields, fmt.Sprintf("AttackTaxGeneric: %d,", effect.AttackTaxGeneric))
	}
	if effect.Kind == game.RuleEffectAdditionalLandPlays {
		if effect.AdditionalLandPlays < 1 {
			return "", errors.New("render: additional land plays requires a positive count")
		}
		fields = append(fields, fmt.Sprintf("AdditionalLandPlays: %d,", effect.AdditionalLandPlays))
	}
	if effect.Kind == game.RuleEffectPlayLandsFromZone ||
		effect.Kind == game.RuleEffectCastSpellsFromZone ||
		effect.Kind == game.RuleEffectCastFromZone {
		ctx.need(importZone)
		castZone, err := renderZone(effect.CastFromZone)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CastFromZone: %s,", castZone))
	}
	castZones, err := renderRuleEffectZoneField(ctx, "CantCastFromZones", effect.CantCastFromZones)
	if err != nil {
		return "", err
	}
	fields = append(fields, castZones...)
	enterZones, err := renderRuleEffectZoneField(ctx, "EnterFromZones", effect.EnterFromZones)
	if err != nil {
		return "", err
	}
	fields = append(fields, enterZones...)
	if effect.EnterExcludeLandCards {
		fields = append(fields, "EnterExcludeLandCards: true,")
	}
	if effect.TopCardOnly {
		fields = append(fields, "TopCardOnly: true,")
	}
	if len(effect.PermanentTypes) > 0 {
		permanentTypes, err := renderTypesCardSlice(ctx, effect.PermanentTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("PermanentTypes: %s,", permanentTypes))
	}
	if len(effect.SpellTypes) > 0 {
		spellTypes, err := renderTypesCardSlice(ctx, effect.SpellTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SpellTypes: %s,", spellTypes))
	}
	if effect.SpellColorless {
		fields = append(fields, "SpellColorless: true,")
	}
	fields = append(fields, renderRuleEffectChosenSubtypeField(effect)...)
	if len(effect.SpellSubtypes) > 0 {
		spellSubtypes, err := renderSubtypeSlice(ctx, effect.SpellSubtypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SpellSubtypes: %s,", spellSubtypes))
	}
	if effect.RestrictedDuringControllerTurn {
		fields = append(fields, "RestrictedDuringControllerTurn: true,")
	}
	if effect.AppliesToNextSpellOnly {
		fields = append(fields, "AppliesToNextSpellOnly: true,")
	}
	return structLit("game.RuleEffect", fields), nil
}

// renderRuleEffectChosenSubtypeField renders the SpellChosenSubtypeFrom entry-
// choice key of a cast-from-zone permission narrowed to the source permanent's
// chosen creature subtype ("creature spells of the chosen type", Realmwalker),
// returning an empty slice when no chosen-type filter applies.
func renderRuleEffectChosenSubtypeField(effect *game.RuleEffect) []string {
	if effect.SpellChosenSubtypeFrom == "" {
		return nil
	}
	switch effect.SpellChosenSubtypeFrom {
	case game.EntryTypeChoiceKey:
		return []string{"SpellChosenSubtypeFrom: game.EntryTypeChoiceKey,"}
	default:
		return []string{fmt.Sprintf("SpellChosenSubtypeFrom: game.ChoiceKey(%q),", effect.SpellChosenSubtypeFrom)}
	}
}

// renderRuleEffectZoneField renders a []zone.Type rule-effect field as a single
// "Name: []zone.Type{...}," struct field, returning an empty slice when the zone
// list is empty so callers append nothing.
func renderRuleEffectZoneField(ctx *renderCtx, name string, zones []zone.Type) ([]string, error) {
	if len(zones) == 0 {
		return nil, nil
	}
	ctx.need(importZone)
	rendered := make([]string, 0, len(zones))
	for _, sourceZone := range zones {
		zoneLit, err := renderZone(sourceZone)
		if err != nil {
			return nil, err
		}
		rendered = append(rendered, zoneLit)
	}
	return []string{fmt.Sprintf("%s: []zone.Type{%s},", name, strings.Join(rendered, ", "))}, nil
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
	case game.RuleEffectCantBeBlockedByMoreThanOne:
		return "game.RuleEffectCantBeBlockedByMoreThanOne", nil
	case game.RuleEffectMustAttack:
		return "game.RuleEffectMustAttack", nil
	case game.RuleEffectMustBeBlocked:
		return "game.RuleEffectMustBeBlocked", nil
	case game.RuleEffectDoesntUntap:
		return "game.RuleEffectDoesntUntap", nil
	case game.RuleEffectCostModifier:
		return "game.RuleEffectCostModifier", nil
	case game.RuleEffectGrantHandCardAbility:
		return "game.RuleEffectGrantHandCardAbility", nil
	case game.RuleEffectPlayerProtection:
		return "game.RuleEffectPlayerProtection", nil
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
	case game.RuleEffectAdditionalTriggerForEnteringPermanent:
		return "game.RuleEffectAdditionalTriggerForEnteringPermanent", nil
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
	default:
		return "", fmt.Errorf("render: unsupported rule effect kind %d", kind)
	}
}

func renderBlockerRestriction(restriction game.BlockerRestriction) (string, error) {
	var kind string
	switch restriction.Kind {
	case game.BlockerRestrictionFlying:
		kind = "game.BlockerRestrictionFlying"
	case game.BlockerRestrictionPowerLessOrEqual:
		kind = "game.BlockerRestrictionPowerLessOrEqual"
	case game.BlockerRestrictionPowerGreaterOrEqual:
		kind = "game.BlockerRestrictionPowerGreaterOrEqual"
	case game.BlockerRestrictionColor:
		kind = "game.BlockerRestrictionColor"
	case game.BlockerRestrictionArtifact:
		kind = "game.BlockerRestrictionArtifact"
	default:
		return "", fmt.Errorf("render: unsupported blocker restriction kind %d", restriction.Kind)
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if restriction.Power != 0 {
		fields = append(fields, fmt.Sprintf("Power: %d,", restriction.Power))
	}
	if restriction.Color != "" {
		literal, err := colorValueToLiteral(restriction.Color)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Color: %s,", literal))
	}
	return structLit("game.BlockerRestriction", fields), nil
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
	if modifier.MatchColor {
		fields = append(fields, "MatchColor: true,")
		if modifier.Color != "" {
			colorLit, err := colorValueToLiteral(modifier.Color)
			if err != nil {
				return "", err
			}
			ctx.need(importColor)
			fields = append(fields, fmt.Sprintf("Color: %s,", colorLit))
		}
	}
	if len(modifier.MatchColors) != 0 {
		colorLits, err := colorValueLiterals(modifier.MatchColors)
		if err != nil {
			return "", err
		}
		ctx.need(importColor)
		fields = append(fields, fmt.Sprintf("MatchColors: []color.Color{%s},", colorLits))
	}
	if len(modifier.MatchSubtypes) != 0 {
		ctx.need(importTypes)
		cardTypeStrings := make([]string, 0, len(subtypeLiteralTypes))
		for typ := range subtypeLiteralTypes {
			cardTypeStrings = append(cardTypeStrings, typ)
		}
		literals := make([]string, 0, len(modifier.MatchSubtypes))
		for _, sub := range modifier.MatchSubtypes {
			literals = append(literals, SubtypeToLiteral(string(sub), cardTypeStrings))
		}
		fields = append(fields, fmt.Sprintf("MatchSubtypes: []types.Sub{%s},", strings.Join(literals, ", ")))
	}
	if modifier.ChosenSubtypeFromEntryChoice {
		fields = append(fields, "ChosenSubtypeFromEntryChoice: true,")
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
	if modifier.PerObjectReduction != 0 {
		fields = append(fields, fmt.Sprintf("PerObjectReduction: %d,", modifier.PerObjectReduction))
	}
	if modifier.CountSelection != nil && !modifier.CountSelection.Empty() {
		selection, err := r.renderSelection(ctx, *modifier.CountSelection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("CountSelection: &%s,", selection))
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
