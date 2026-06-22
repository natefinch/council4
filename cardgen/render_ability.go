package cardgen

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (r Renderer) renderActivatedAbility(ctx *renderCtx, ability *game.ActivatedAbility) (string, error) {
	if manaCost, ok := game.ActivatedBodyEquipCost(ability); ok &&
		reflect.DeepEqual(*ability, game.EquipActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EquipActivatedAbility(%s)", renderedCost), nil
	}
	if rendered, ok, err := r.renderEquipRestrictedAbility(ctx, ability); ok {
		return rendered, err
	}
	if manaCost, ok := game.ActivatedBodyCyclingCost(ability); ok &&
		reflect.DeepEqual(*ability, game.CyclingActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.CyclingActivatedAbility(%s)", renderedCost), nil
	}
	if manaCost, ok := game.ActivatedBodyScavengeCost(ability); ok &&
		reflect.DeepEqual(*ability, game.ScavengeActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.ScavengeActivatedAbility(%s)", renderedCost), nil
	}
	if manaCost, ok := game.ActivatedBodyUnearthCost(ability); ok &&
		reflect.DeepEqual(*ability, game.UnearthActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.UnearthActivatedAbility(%s)", renderedCost), nil
	}
	if power, ok := game.ActivatedBodySaddlePower(ability); ok &&
		reflect.DeepEqual(*ability, game.SaddleActivatedAbility(power)) {
		return fmt.Sprintf("game.SaddleActivatedAbility(%d)", power), nil
	}
	if manaCost, subtypes, ok := game.ActivatedBodyEternalizeParams(ability); ok &&
		reflect.DeepEqual(*ability, game.EternalizeActivatedBody(manaCost, subtypes...)) {
		return r.renderEternalizeFamilyAbility(ctx, "game.EternalizeActivatedBody", manaCost, subtypes)
	}
	if manaCost, subtypes, ok := game.ActivatedBodyEmbalmParams(ability); ok &&
		reflect.DeepEqual(*ability, game.EmbalmActivatedBody(manaCost, subtypes...)) {
		return r.renderEternalizeFamilyAbility(ctx, "game.EmbalmActivatedBody", manaCost, subtypes)
	}

	var fields []string
	if ability.ManaCost.Exists {
		ctx.need(importOpt)
		manaCostLit, err := r.renderManaCost(ctx, ability.ManaCost.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ManaCost: opt.Val(%s),", manaCostLit))
	}
	if len(ability.AdditionalCosts) > 0 {
		rendered, err := r.renderAdditionalCosts(ctx, ability.AdditionalCosts)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AdditionalCosts: %s,", rendered))
	}
	if len(ability.CostModifiers) > 0 {
		rendered := make([]string, 0, len(ability.CostModifiers))
		for i := range ability.CostModifiers {
			value, err := r.renderCostModifier(ctx, ability.CostModifiers[i])
			if err != nil {
				return "", err
			}
			rendered = append(rendered, value)
		}
		fields = append(fields, fmt.Sprintf("CostModifiers: []game.CostModifier{%s},", strings.Join(rendered, ", ")))
	}
	if ability.ZoneOfFunction != zone.None {
		ctx.need(importZone)
		zoneLiteral, err := renderZone(ability.ZoneOfFunction)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ZoneOfFunction: %s,", zoneLiteral))
	}
	if ability.Timing != game.NoTimingRestriction {
		timing, err := renderTimingRestriction(ability.Timing)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Timing: %s,", timing))
	}
	if len(ability.KeywordAbilities) > 0 {
		elements := make([]string, 0, len(ability.KeywordAbilities))
		for _, keyword := range ability.KeywordAbilities {
			rendered, err := r.renderKeywordAbility(ctx, keyword)
			if err != nil {
				return "", err
			}
			elements = append(elements, rendered+",")
		}
		fields = append(fields, sliceField("KeywordAbilities", "game.KeywordAbility", elements))
	}
	if ability.ActivationCondition.Exists {
		condition, err := r.renderControllerControlsCondition(ctx, &ability.ActivationCondition.Val, "activated ability")
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ActivationCondition: opt.Val(%s),", condition))
	}
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Content: %s,", content))
	return structLit("game.ActivatedAbility", fields), nil
}

func (r Renderer) renderEquipRestrictedAbility(
	ctx *renderCtx,
	ability *game.ActivatedAbility,
) (rendered string, matched bool, err error) {
	manaCost, ok := game.ActivatedBodyEquipCost(ability)
	if !ok {
		return "", false, nil
	}
	supertypes, subtypes, ok := equipRestrictionTypes(ability)
	if !ok || (len(supertypes) == 0 && len(subtypes) == 0) {
		return "", false, nil
	}
	if !reflect.DeepEqual(*ability, game.EquipRestrictedActivatedAbility(manaCost, supertypes, subtypes)) {
		return "", false, nil
	}
	renderedCost, err := r.renderManaCost(ctx, manaCost)
	if err != nil {
		return "", true, err
	}
	superLit, err := renderSupertypeSlice(ctx, supertypes)
	if err != nil {
		return "", true, err
	}
	subLit, err := renderSubtypeSlice(ctx, subtypes)
	if err != nil {
		return "", true, err
	}
	return fmt.Sprintf(
		"game.EquipRestrictedActivatedAbility(%s, %s, %s)",
		renderedCost, superLit, subLit,
	), true, nil
}

func equipRestrictionTypes(ability *game.ActivatedAbility) ([]types.Super, []types.Sub, bool) {
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Targets) != 1 {
		return nil, nil, false
	}
	predicate := ability.Content.Modes[0].Targets[0].Predicate
	return predicate.Supertypes, predicate.Subtypes, true
}

func renderSupertypeSlice(ctx *renderCtx, supertypes []types.Super) (string, error) {
	if len(supertypes) == 0 {
		return "nil", nil
	}
	ctx.need(importTypes)
	literals := make([]string, 0, len(supertypes))
	for _, st := range supertypes {
		lit, err := supertypeLiteral(st)
		if err != nil {
			return "", err
		}
		literals = append(literals, lit)
	}
	return fmt.Sprintf("[]types.Super{%s}", strings.Join(literals, ", ")), nil
}

func renderSubtypeSlice(ctx *renderCtx, subtypes []types.Sub) (string, error) {
	if len(subtypes) == 0 {
		return "nil", nil
	}
	arguments, err := renderSubtypeArguments(ctx, subtypes)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("[]types.Sub{%s}", arguments), nil
}

func (r Renderer) renderManaAbility(ctx *renderCtx, ability *game.ManaAbility) (string, error) {
	for _, manaColor := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C} {
		if !reflect.DeepEqual(*ability, game.TapManaAbility(manaColor)) {
			continue
		}
		ctx.need(importMana)
		colorLiteral, err := renderManaColor(manaColor)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.TapManaAbility(%s)", colorLiteral), nil
	}
	if colors, ok := tapManaChoiceColors(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaChoiceAbility(colors...)) {
		ctx.need(importMana)
		colorLiterals := make([]string, 0, len(colors))
		for _, manaColor := range colors {
			colorLiteral, err := renderManaColor(manaColor)
			if err != nil {
				return "", err
			}
			colorLiterals = append(colorLiterals, colorLiteral)
		}
		return fmt.Sprintf("game.TapManaChoiceAbility(%s)", strings.Join(colorLiterals, ", ")), nil
	}
	if colors, count, ok := tapManaChoiceCountColors(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaChoiceCountAbility(ability.Text, count, colors...)) {
		ctx.need(importMana)
		colorLiterals := make([]string, 0, len(colors))
		for _, manaColor := range colors {
			colorLiteral, err := renderManaColor(manaColor)
			if err != nil {
				return "", err
			}
			colorLiterals = append(colorLiterals, colorLiteral)
		}
		return fmt.Sprintf("game.TapManaChoiceCountAbility(%q, %d, %s)", ability.Text, count, strings.Join(colorLiterals, ", ")), nil
	}
	if reflect.DeepEqual(*ability, game.TapChosenColorManaAbility(ability.Text)) {
		return fmt.Sprintf("game.TapChosenColorManaAbility(%q)", ability.Text), nil
	}
	for _, fixed := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C} {
		if reflect.DeepEqual(*ability, game.TapFixedOrChosenColorManaAbility(ability.Text, fixed)) {
			ctx.need(importMana)
			colorLiteral, err := renderManaColor(fixed)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.TapFixedOrChosenColorManaAbility(%q, %s)", ability.Text, colorLiteral), nil
		}
	}
	if reflect.DeepEqual(*ability, game.TapManaCommanderIdentityAbility()) {
		return "game.TapManaCommanderIdentityAbility()", nil
	}
	for _, relation := range []game.PlayerRelation{game.PlayerYou, game.PlayerOpponent} {
		for _, includeColorless := range []bool{false, true} {
			if reflect.DeepEqual(*ability, game.TapManaLandsProduceAbility(relation, includeColorless)) {
				literal, err := renderPlayerRelation(relation)
				if err != nil {
					return "", err
				}
				return fmt.Sprintf("game.TapManaLandsProduceAbility(%s, %t)", literal, includeColorless), nil
			}
		}
	}
	if linkID, ok := linkedExileColorManaLinkID(ability); ok &&
		reflect.DeepEqual(*ability, game.TapLinkedExileColorManaAbility(linkID)) {
		return fmt.Sprintf("game.TapLinkedExileColorManaAbility(%q)", linkID), nil
	}
	if selection, ok := amongControlledColorsSelection(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaAmongControlledColorsAbility(ability.Text, selection)) {
		rendered, err := r.renderSelection(ctx, selection)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.TapManaAmongControlledColorsAbility(%q, %s)", ability.Text, rendered), nil
	}
	if selection, ok := eachControlledColorSelection(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaEachControlledColorAbility(ability.Text, selection)) {
		rendered, err := r.renderSelection(ctx, selection)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.TapManaEachControlledColorAbility(%q, %s)", ability.Text, rendered), nil
	}

	if game.IsTapSacrificeAnyOneColorManaAbility(ability) {
		_, count, ok := game.ManaAbilityChoiceOutput(ability)
		if ok {
			return fmt.Sprintf("game.TapSacrificeAnyOneColorManaAbility(%q, %d)", ability.Text, count), nil
		}
	}

	var fields []string
	if ability.ZoneOfFunction != zone.None {
		ctx.need(importZone)
		zoneLiteral, err := renderZone(ability.ZoneOfFunction)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ZoneOfFunction: %s,", zoneLiteral))
	}
	if ability.ManaCost.Exists {
		ctx.need(importOpt)
		manaCostLit, err := r.renderManaCost(ctx, ability.ManaCost.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ManaCost: opt.Val(%s),", manaCostLit))
	}
	if len(ability.AdditionalCosts) > 0 {
		rendered, err := r.renderAdditionalCosts(ctx, ability.AdditionalCosts)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AdditionalCosts: %s,", rendered))
	}
	if ability.Timing != game.NoTimingRestriction {
		timing, err := renderTimingRestriction(ability.Timing)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Timing: %s,", timing))
	}
	if ability.ActivationCondition.Exists {
		condition, err := r.renderControllerControlsCondition(ctx, &ability.ActivationCondition.Val, "mana ability")
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ActivationCondition: opt.Val(%s),", condition))
	}
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Content: %s,", content))
	return structLit("game.ManaAbility", fields), nil
}

func renderTimingRestriction(timing game.TimingRestriction) (string, error) {
	switch timing {
	case game.NoTimingRestriction:
		return "game.NoTimingRestriction", nil
	case game.SorceryOnly:
		return "game.SorceryOnly", nil
	case game.OncePerTurn:
		return "game.OncePerTurn", nil
	case game.SorceryOncePerTurn:
		return "game.SorceryOncePerTurn", nil
	case game.DuringCombat:
		return "game.DuringCombat", nil
	case game.DuringUpkeep:
		return "game.DuringUpkeep", nil
	case game.DuringYourTurn:
		return "game.DuringYourTurn", nil
	default:
		return "", fmt.Errorf("unsupported timing restriction %d", timing)
	}
}

func tapManaChoiceColors(ability *game.ManaAbility) ([]mana.Color, bool) {
	colors, amount, ok := game.ManaAbilityChoiceOutput(ability)
	return colors, ok && amount == 1
}

// tapManaChoiceCountColors extracts the color choices and produced count from a
// mana ability that adds N mana (N >= 2) of a single chosen color, so the
// ability can render back to game.TapManaChoiceCountAbility (Gilded Lotus's
// "Add three mana of any one color."). It rejects the single-mana choice, which
// renders to game.TapManaChoiceAbility instead.
func tapManaChoiceCountColors(ability *game.ManaAbility) ([]mana.Color, int, bool) {
	colors, amount, ok := game.ManaAbilityChoiceOutput(ability)
	return colors, amount, ok && amount >= 2
}

// linkedExileColorManaLinkID extracts the imprint link identifier from a mana
// ability whose single mana-color choice draws on a linked exiled card's colors,
// so the ability can render back to game.TapLinkedExileColorManaAbility(linkID).
func linkedExileColorManaLinkID(ability *game.ManaAbility) (string, bool) {
	if len(ability.Content.Modes) != 1 {
		return "", false
	}
	for i := range ability.Content.Modes[0].Sequence {
		choose, ok := ability.Content.Modes[0].Sequence[i].Primitive.(game.Choose)
		if !ok {
			continue
		}
		if choose.Choice.Kind == game.ResolutionChoiceMana &&
			choose.Choice.ColorSource == game.ResolutionChoiceColorSourceLinkedExileColors {
			return choose.Choice.LinkID, true
		}
	}
	return "", false
}

// amongControlledColorsSelection extracts the permanent filter from a mana
// ability whose single mana-color choice draws on the colors of permanents the
// controller controls, so the ability can render back to
// game.TapManaAmongControlledColorsAbility (Mox Amber, Plaza of Heroes).
func amongControlledColorsSelection(ability *game.ManaAbility) (game.Selection, bool) {
	if len(ability.Content.Modes) != 1 {
		return game.Selection{}, false
	}
	for i := range ability.Content.Modes[0].Sequence {
		choose, ok := ability.Content.Modes[0].Sequence[i].Primitive.(game.Choose)
		if !ok {
			continue
		}
		if choose.Choice.Kind == game.ResolutionChoiceMana &&
			choose.Choice.ColorSource == game.ResolutionChoiceColorSourceControlledPermanentColors &&
			choose.Choice.Selection != nil {
			return *choose.Choice.Selection, true
		}
	}
	return game.Selection{}, false
}

// eachControlledColorSelection extracts the permanent filter from a mana ability
// that produces one mana of each color among the controller's permanents, so the
// ability can render back to game.TapManaEachControlledColorAbility (Bloom
// Tender). It matches a single AddMana instruction carrying an EachControlledColor
// selection.
func eachControlledColorSelection(ability *game.ManaAbility) (game.Selection, bool) {
	if len(ability.Content.Modes) != 1 {
		return game.Selection{}, false
	}
	for i := range ability.Content.Modes[0].Sequence {
		add, ok := ability.Content.Modes[0].Sequence[i].Primitive.(game.AddMana)
		if !ok {
			continue
		}
		if add.EachControlledColor != nil {
			return *add.EachControlledColor, true
		}
	}
	return game.Selection{}, false
}

func (r Renderer) renderTriggeredAbility(ctx *renderCtx, ability *game.TriggeredAbility) (string, error) {
	if keyword, ok := game.BodyKeywordAbility(ability, game.CumulativeUpkeep); ok {
		if cumulative, ok := keyword.(game.CumulativeUpkeepKeyword); ok &&
			reflect.DeepEqual(*ability, game.CumulativeUpkeepTriggeredAbility(cumulative.Cost)) {
			renderedCost, err := r.renderManaCost(ctx, cumulative.Cost)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.CumulativeUpkeepTriggeredAbility(%s)", renderedCost), nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Fabricate); ok {
		if fabricate, ok := keyword.(game.FabricateKeyword); ok &&
			reflect.DeepEqual(*ability, game.FabricateTriggeredAbility(fabricate.Count)) {
			return fmt.Sprintf("game.FabricateTriggeredAbility(%d)", fabricate.Count), nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Soulshift); ok {
		if soulshift, ok := keyword.(game.SoulshiftKeyword); ok &&
			reflect.DeepEqual(*ability, game.SoulshiftTriggeredAbility(soulshift.Count)) {
			return fmt.Sprintf("game.SoulshiftTriggeredAbility(%d)", soulshift.Count), nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Rampage); ok {
		if rampage, ok := keyword.(game.RampageKeyword); ok &&
			reflect.DeepEqual(*ability, game.RampageTriggeredAbility(rampage.Count)) {
			return fmt.Sprintf("game.RampageTriggeredAbility(%d)", rampage.Count), nil
		}
	}
	if reflect.DeepEqual(*ability, game.UndyingTriggeredBody) {
		return "game.UndyingTriggeredBody", nil
	}
	if reflect.DeepEqual(*ability, game.PersistTriggeredBody) {
		return "game.PersistTriggeredBody", nil
	}
	if reflect.DeepEqual(*ability, game.DethroneTriggeredBody) {
		return "game.DethroneTriggeredBody", nil
	}
	if reflect.DeepEqual(*ability, game.FlankingTriggeredBody) {
		return "game.FlankingTriggeredBody", nil
	}
	if reflect.DeepEqual(*ability, game.TrainingTriggeredBody) {
		return "game.TrainingTriggeredBody", nil
	}
	if reflect.DeepEqual(*ability, game.LivingWeaponTriggeredAbility()) {
		return "game.LivingWeaponTriggeredAbility()", nil
	}
	var fields []string
	trigger, err := r.renderTriggerCondition(ctx, &ability.Trigger)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Trigger: %s,", trigger))
	if ability.Optional {
		fields = append(fields, "Optional: true,")
	}
	if ability.MaxTriggersPerTurn != 0 {
		fields = append(fields, fmt.Sprintf("MaxTriggersPerTurn: %d,", ability.MaxTriggersPerTurn))
	}
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Content: %s,", content))
	return structLit("game.TriggeredAbility", fields), nil
}

func (r Renderer) renderChapterAbility(ctx *renderCtx, ability *game.ChapterAbility) (string, error) {
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	return structLit("game.ChapterAbility", []string{
		fmt.Sprintf("Text: %s,", renderText(ability.Text)),
		fmt.Sprintf("Chapters: %#v,", ability.Chapters),
		fmt.Sprintf("Content: %s,", content),
	}), nil
}

func (r Renderer) renderLoyaltyAbility(ctx *renderCtx, ability *game.LoyaltyAbility) (string, error) {
	var fields []string
	fields = append(fields, fmt.Sprintf("LoyaltyCost: %d,", ability.LoyaltyCost))
	content, err := r.renderAbilityContent(ctx, ability.Content)
	if err != nil {
		return "", err
	}
	fields = append(fields, fmt.Sprintf("Content: %s,", content))
	return structLit("game.LoyaltyAbility", fields), nil
}

func (r Renderer) renderTriggerCondition(ctx *renderCtx, trigger *game.TriggerCondition) (string, error) {
	triggerType, err := renderTriggerType(trigger.Type)
	if err != nil {
		return "", err
	}
	pattern, err := r.renderTriggerPattern(ctx, &trigger.Pattern)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Type: %s,", triggerType),
		fmt.Sprintf("Pattern: %s,", pattern),
	}
	if trigger.InterveningIf != "" {
		fields = append(fields, fmt.Sprintf("InterveningIf: %q,", trigger.InterveningIf))
	}
	if trigger.InterveningCondition.Exists {
		condition, err := r.renderControllerControlsCondition(ctx, &trigger.InterveningCondition.Val, "trigger intervening")
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("InterveningCondition: opt.Val(%s),", condition))
	}
	if trigger.InterveningIfEventPermanentHadNoCounterKind.Exists {
		kind, err := renderCounterKind(trigger.InterveningIfEventPermanentHadNoCounterKind.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("InterveningIfEventPermanentHadNoCounterKind: opt.Val(%s),", kind))
	}
	if trigger.InterveningIfEventPermanentHadCounters {
		fields = append(fields, "InterveningIfEventPermanentHadCounters: true,")
	}
	if trigger.InterveningIfEventPermanentWasKicked {
		fields = append(fields, "InterveningIfEventPermanentWasKicked: true,")
	}
	if trigger.InterveningIfEventPermanentWasCast {
		fields = append(fields, "InterveningIfEventPermanentWasCast: true,")
	}
	if trigger.InterveningIfEventPermanentWasCastByController {
		fields = append(fields, "InterveningIfEventPermanentWasCastByController: true,")
	}
	if trigger.InterveningIfEventPermanentEnteredOrCastFromGraveyard {
		fields = append(fields, "InterveningIfEventPermanentEnteredOrCastFromGraveyard: true,")
	}
	if trigger.InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard {
		fields = append(fields, "InterveningIfEventPermanentEnteredOrCastFromControllerGraveyard: true,")
	}
	return structLit("game.TriggerCondition", fields), nil
}

func (Renderer) renderTriggerPattern(ctx *renderCtx, pattern *game.TriggerPattern) (string, error) {
	if (pattern.Event == game.EventBeginningOfStep) != (pattern.Step != game.StepNone) {
		return "", errors.New("render: beginning-of-step trigger pattern must set exactly one supported step")
	}
	allowZoneChangeZones := pattern.Event == game.EventZoneChanged
	allowFromZone := pattern.MatchFromZone &&
		(pattern.Event == game.EventSpellCast || pattern.Event == game.EventPermanentEnteredBattlefield || allowZoneChangeZones) &&
		!pattern.MatchToZone
	if len(pattern.RequireCardTypes) != 0 ||
		len(pattern.ExcludeCardTypes) != 0 ||
		(pattern.MatchFromZone && !allowFromZone && !allowZoneChangeZones) ||
		(pattern.MatchToZone && !allowZoneChangeZones) ||
		(pattern.ExcludeToZone && !allowZoneChangeZones) ||
		(pattern.MatchToZone && pattern.ExcludeToZone) ||
		pattern.DamageRecipientCombatState != game.CombatStateAny ||
		pattern.SpellTargetsSource ||
		pattern.SpellTargetAllow != game.TargetAllowUnspecified ||
		pattern.SpellTargetPattern.Exists ||
		(pattern.RequireKickerPaid && pattern.Event != game.EventSpellCast) ||
		(pattern.RequireHistoric && pattern.Event != game.EventSpellCast) ||
		(pattern.MatchSpellCopy && pattern.Event != game.EventSpellCast) ||
		(pattern.RequireTappedForMana && pattern.Event != game.EventPermanentTapped) ||
		(pattern.ExcludeManaAbility && pattern.Event != game.EventAbilityActivated) ||
		(pattern.Event == game.EventAbilityActivated && !pattern.ExcludeManaAbility) ||
		(pattern.PlayerEventOrdinalThisTurn > 0 &&
			pattern.Event != game.EventCardDrawn &&
			pattern.Event != game.EventLifeGained &&
			pattern.Event != game.EventLifeLost &&
			pattern.Event != game.EventScry &&
			pattern.Event != game.EventSurveil &&
			pattern.Event != game.EventSpellCast) ||
		(pattern.RequireCombatDamage && pattern.RequireNonCombatDamage) ||
		(pattern.AttackAlone && pattern.Event != game.EventAttackerDeclared) ||
		(pattern.AttackWhileSaddled && pattern.Event != game.EventAttackerDeclared) ||
		(pattern.ExcludeFirstDrawInDrawStep && pattern.Event != game.EventCardDrawn) ||
		(pattern.AttackerCountAtLeast != 0 &&
			(pattern.Event != game.EventAttackerDeclared || !pattern.OneOrMore || pattern.AttackAlone || pattern.AttackerCountAtLeast < 2)) {
		return "", errors.New("render: unsupported trigger pattern fields")
	}
	if err := validateTriggerPatternCardSelection(pattern); err != nil {
		return "", err
	}
	event, err := renderEventKind(pattern.Event)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Event: %s,", event)}
	relationFields, err := renderTriggerPatternRelationFields(pattern)
	if err != nil {
		return "", err
	}
	fields = append(fields, relationFields...)
	zoneFields, err := renderTriggerPatternZoneFields(ctx, pattern)
	if err != nil {
		return "", err
	}
	fields = append(fields, zoneFields...)
	flagFields, err := renderTriggerPatternFlagFields(ctx, pattern)
	if err != nil {
		return "", err
	}
	fields = append(fields, flagFields...)
	selectionFields, err := renderTriggerPatternSelectionFields(ctx, pattern)
	if err != nil {
		return "", err
	}
	fields = append(fields, selectionFields...)
	return structLit("game.TriggerPattern", fields), nil
}

func renderTriggerPatternRelationFields(pattern *game.TriggerPattern) ([]string, error) {
	var fields []string
	if pattern.Source != game.TriggerSourceAny {
		source, err := renderTriggerSource(pattern.Source)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("Source: %s,", source))
	}
	if pattern.Controller != game.TriggerControllerAny {
		controller, err := renderTriggerController(pattern.Controller)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("Controller: %s,", controller))
	}
	if pattern.CauseController != game.TriggerControllerAny {
		controller, err := renderTriggerController(pattern.CauseController)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("CauseController: %s,", controller))
	}
	if pattern.Step != game.StepNone {
		step, err := renderStep(pattern.Step)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("Step: %s,", step))
	}
	if pattern.Subject != game.TriggerSubjectDefault {
		subject, err := renderTriggerSubject(pattern.Subject)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("Subject: %s,", subject))
	}
	if pattern.ExcludeSelf {
		fields = append(fields, "ExcludeSelf: true,")
	}
	if pattern.SubjectSelectionOrSelf {
		fields = append(fields, "SubjectSelectionOrSelf: true,")
	}
	if pattern.Player != game.TriggerPlayerAny {
		player, err := renderTriggerPlayer(pattern.Player)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("Player: %s,", player))
	}
	return fields, nil
}

func renderTriggerPatternFlagFields(ctx *renderCtx, pattern *game.TriggerPattern) ([]string, error) {
	var fields []string
	if pattern.MatchFaceDown {
		fields = append(fields, "MatchFaceDown: true,", fmt.Sprintf("FaceDown: %t,", pattern.FaceDown))
	}
	if pattern.RequireKickerPaid {
		fields = append(fields, "RequireKickerPaid: true,")
	}
	if pattern.RequireHistoric {
		fields = append(fields, "RequireHistoric: true,")
	}
	if pattern.MatchSpellCopy {
		fields = append(fields, "MatchSpellCopy: true,")
	}
	if pattern.RequireTappedForMana {
		fields = append(fields, "RequireTappedForMana: true,")
	}
	if pattern.RequireProducedManaColor != "" {
		colorLiteral, err := renderManaColor(pattern.RequireProducedManaColor)
		if err != nil {
			return nil, err
		}
		ctx.need(importMana)
		fields = append(fields, fmt.Sprintf("RequireProducedManaColor: %s,", colorLiteral))
	}
	if pattern.UnionEvent != game.EventUnknown {
		unionEvent, err := renderEventKind(pattern.UnionEvent)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("UnionEvent: %s,", unionEvent))
	}
	if pattern.ExcludeManaAbility {
		fields = append(fields, "ExcludeManaAbility: true,")
	}
	if pattern.PlayerEventOrdinalThisTurn > 0 {
		fields = append(fields, fmt.Sprintf("PlayerEventOrdinalThisTurn: %d,", pattern.PlayerEventOrdinalThisTurn))
	}
	if pattern.ExcludeFirstDrawInDrawStep {
		fields = append(fields, "ExcludeFirstDrawInDrawStep: true,")
	}
	if pattern.MatchStackObjectKind {
		stackObjectKind, err := renderStackObjectKind(pattern.StackObjectKind)
		if err != nil {
			return nil, err
		}
		fields = append(fields, "MatchStackObjectKind: true,", fmt.Sprintf("StackObjectKind: %s,", stackObjectKind))
	}
	if len(pattern.RequirePermanentTypes) > 0 {
		rpt, err := renderTypesCardSlice(ctx, pattern.RequirePermanentTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("RequirePermanentTypes: %s,", rpt))
	}
	if len(pattern.ExcludePermanentTypes) > 0 {
		ept, err := renderTypesCardSlice(ctx, pattern.ExcludePermanentTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("ExcludePermanentTypes: %s,", ept))
	}
	if pattern.RequireNonToken {
		fields = append(fields, "RequireNonToken: true,")
	}
	if pattern.OneOrMore {
		fields = append(fields, "OneOrMore: true,")
	}
	if pattern.OneOrMorePerAttackTarget {
		fields = append(fields, "OneOrMorePerAttackTarget: true,")
	}
	if pattern.AttackAlone {
		fields = append(fields, "AttackAlone: true,")
	}
	if pattern.AttackWhileSaddled {
		fields = append(fields, "AttackWhileSaddled: true,")
	}
	if pattern.AttackerCountAtLeast != 0 {
		fields = append(fields, fmt.Sprintf("AttackerCountAtLeast: %d,", pattern.AttackerCountAtLeast))
	}
	if pattern.MatchCounterKind {
		kindFields, err := renderTriggerPatternCounterKind(ctx, pattern)
		if err != nil {
			return nil, err
		}
		fields = append(fields, kindFields...)
	}
	if pattern.RequireCombatDamage {
		fields = append(fields, "RequireCombatDamage: true,")
	}
	if pattern.RequireNonCombatDamage {
		fields = append(fields, "RequireNonCombatDamage: true,")
	}
	if pattern.DamageRecipientIsSource {
		fields = append(fields, "DamageRecipientIsSource: true,")
	}
	if pattern.AttackRecipient != game.AttackRecipientAny {
		recipient, err := renderAttackRecipient(pattern.AttackRecipient)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("AttackRecipient: %s,", recipient))
	}
	return fields, nil
}

func renderTriggerPatternSelectionFields(ctx *renderCtx, pattern *game.TriggerPattern) ([]string, error) {
	var fields []string
	if pattern.DamageRecipient != game.DamageRecipientNone ||
		len(pattern.DamageRecipientTypes) > 0 ||
		!pattern.DamageRecipientSelection.Empty() {
		damageFields, err := renderTriggerPatternDamageFields(ctx, pattern)
		if err != nil {
			return nil, err
		}
		fields = append(fields, damageFields...)
	}
	if !pattern.CardSelection.Empty() {
		cardFields, err := renderTriggerPatternCardSelectionFields(ctx, pattern)
		if err != nil {
			return nil, err
		}
		fields = append(fields, cardFields...)
	}
	if !pattern.SubjectSelection.Empty() {
		subjectFields, err := renderTriggerPatternSelection(ctx, "SubjectSelection", pattern.SubjectSelection)
		if err != nil {
			return nil, err
		}
		fields = append(fields, subjectFields...)
	}
	if !pattern.RelatedSubjectSelection.Empty() {
		relatedFields, err := renderTriggerPatternSelection(ctx, "RelatedSubjectSelection", pattern.RelatedSubjectSelection)
		if err != nil {
			return nil, err
		}
		fields = append(fields, relatedFields...)
	}
	if !pattern.AttackRecipientSelection.Empty() {
		attackFields, err := renderTriggerPatternSelection(ctx, "AttackRecipientSelection", pattern.AttackRecipientSelection)
		if err != nil {
			return nil, err
		}
		fields = append(fields, attackFields...)
	}
	if !pattern.DamageSourceSelection.Empty() {
		sourceFields, err := renderTriggerPatternSelection(ctx, "DamageSourceSelection", pattern.DamageSourceSelection)
		if err != nil {
			return nil, err
		}
		fields = append(fields, sourceFields...)
	}
	if !pattern.StepPlayerSourceAttachedSelection.Empty() {
		stepFields, err := renderTriggerPatternSelection(ctx, "StepPlayerSourceAttachedSelection", pattern.StepPlayerSourceAttachedSelection)
		if err != nil {
			return nil, err
		}
		fields = append(fields, stepFields...)
	}
	return fields, nil
}

func renderTriggerPatternZoneFields(ctx *renderCtx, pattern *game.TriggerPattern) ([]string, error) {
	var fields []string
	if pattern.MatchFromZone {
		fromZone, err := renderZone(pattern.FromZone)
		if err != nil {
			return nil, err
		}
		ctx.need(importZone)
		fields = append(fields, "MatchFromZone: true,", fmt.Sprintf("FromZone: %s,", fromZone))
	}
	if pattern.MatchToZone {
		toZone, err := renderZone(pattern.ToZone)
		if err != nil {
			return nil, err
		}
		ctx.need(importZone)
		fields = append(fields, "MatchToZone: true,", fmt.Sprintf("ToZone: %s,", toZone))
	}
	if pattern.ExcludeToZone {
		toZone, err := renderZone(pattern.ToZone)
		if err != nil {
			return nil, err
		}
		ctx.need(importZone)
		fields = append(fields, "ExcludeToZone: true,", fmt.Sprintf("ToZone: %s,", toZone))
	}
	return fields, nil
}

func renderStackObjectKind(kind game.StackObjectKind) (string, error) {
	switch kind {
	case game.StackSpell:
		return "game.StackSpell", nil
	case game.StackActivatedAbility:
		return "game.StackActivatedAbility", nil
	case game.StackTriggeredAbility:
		return "game.StackTriggeredAbility", nil
	default:
		return "", fmt.Errorf("render: unsupported stack object kind %d", kind)
	}
}

// validateTriggerPatternCardSelection validates CardSelection constraints for a
// TriggerPattern and returns an error if they are unsupported. Spell-cast
// triggers read full card characteristics from the event, while discard
// triggers can only filter the discarded card's types (CR 603.2).
func validateTriggerPatternCardSelection(pattern *game.TriggerPattern) error {
	if pattern.CardSelection.Empty() {
		return nil
	}
	switch pattern.Event {
	case game.EventSpellCast, game.EventCardDiscarded:
	default:
		return errors.New("render: CardSelection is only supported for EventSpellCast and EventCardDiscarded trigger patterns")
	}
	unsupported := pattern.CardSelection
	unsupported.RequiredTypes = nil
	unsupported.RequiredTypesAny = nil
	unsupported.ExcludedTypes = nil
	if pattern.Event == game.EventSpellCast {
		unsupported.Supertypes = nil
		unsupported.SubtypesAny = nil
		unsupported.SubtypeChoice = game.SubtypeChoiceWithoutEntry(unsupported.SubtypeChoice)
		unsupported.ColorsAny = nil
		unsupported.Colorless = false
		unsupported.Multicolored = false
		unsupported.ManaValue.Exists = false
	}
	if !unsupported.Empty() {
		return errors.New("render: unsupported CardSelection fields in trigger pattern")
	}
	return nil
}

// renderTriggerPatternCardSelectionFields renders the CardSelection field for a
// TriggerPattern and returns it as a slice of struct-literal field strings.
func renderTriggerPatternCardSelectionFields(ctx *renderCtx, pattern *game.TriggerPattern) ([]string, error) {
	sel, err := (Renderer{}).renderSelection(ctx, pattern.CardSelection)
	if err != nil {
		return nil, err
	}
	return []string{fmt.Sprintf("CardSelection: %s,", sel)}, nil
}

// renderTriggerPatternSelection renders a Selection-valued TriggerPattern
// field and returns it as a slice of struct-literal field strings.
func renderTriggerPatternSelection(ctx *renderCtx, field string, selection game.Selection) ([]string, error) {
	sel, err := (Renderer{}).renderSelection(ctx, selection)
	if err != nil {
		return nil, err
	}
	return []string{fmt.Sprintf("%s: %s,", field, sel)}, nil
}

// renderTriggerPatternCounterKind renders the MatchCounterKind and CounterKind
// fields for a TriggerPattern and appends the import requirement.
func renderTriggerPatternCounterKind(ctx *renderCtx, pattern *game.TriggerPattern) ([]string, error) {
	kind, err := renderCounterKind(pattern.CounterKind)
	if err != nil {
		return nil, fmt.Errorf("render: trigger pattern counter kind: %w", err)
	}
	ctx.need(importCounter)
	return []string{"MatchCounterKind: true,", fmt.Sprintf("CounterKind: %s,", kind)}, nil
}

// renderTriggerPatternDamageFields renders DamageRecipient and
// DamageRecipientTypes fields for a TriggerPattern.
func renderTriggerPatternDamageFields(ctx *renderCtx, pattern *game.TriggerPattern) ([]string, error) {
	var fields []string
	if pattern.DamageRecipient != game.DamageRecipientNone {
		recipient, err := renderDamageRecipient(pattern.DamageRecipient)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("DamageRecipient: %s,", recipient))
	}
	if len(pattern.DamageRecipientTypes) > 0 {
		recipientTypes, err := renderTypesCardSlice(ctx, pattern.DamageRecipientTypes)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("DamageRecipientTypes: %s,", recipientTypes))
	}
	if !pattern.DamageRecipientSelection.Empty() {
		selection, err := (Renderer{}).renderSelection(ctx, pattern.DamageRecipientSelection)
		if err != nil {
			return nil, err
		}
		fields = append(fields, fmt.Sprintf("DamageRecipientSelection: %s,", selection))
	}
	return fields, nil
}
