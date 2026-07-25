package cardgen

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

func (r Renderer) renderActivatedAbility(ctx *renderCtx, ability *game.ActivatedAbility) (string, error) {
	return r.renderGameActivatedAbility(ctx, *ability)
}

// constructorActivatedAbility spells an ActivatedAbility as a game constructor call when one
// produces exactly this value. See constructorRenderers in cardgen/cmd/genrender.
func (r Renderer) constructorActivatedAbility(ctx *renderCtx, v game.ActivatedAbility) (string, bool, error) {
	ability := &v

	if manaCost, ok := game.ActivatedBodyEquipCost(ability); ok &&
		reflect.DeepEqual(*ability, game.EquipActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.EquipActivatedAbility(%s)", renderedCost), true, nil
	}
	if manaCost, ok := game.ActivatedBodyEquipCost(ability); ok &&
		reflect.DeepEqual(*ability, game.EquipCommanderActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.EquipCommanderActivatedAbility(%s)", renderedCost), true, nil
	}
	if manaCost, ok := game.ActivatedBodyEquipCost(ability); ok && len(ability.CostModifiers) == 1 &&
		reflect.DeepEqual(*ability, game.EquipCostReductionActivatedAbility(manaCost, ability.CostModifiers[0])) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		renderedModifier, err := r.renderCostModifier(ctx, ability.CostModifiers[0])
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.EquipCostReductionActivatedAbility(%s, %s)", renderedCost, renderedModifier), true, nil
	}
	if rendered, ok, err := r.renderEquipRestrictedCostReductionAbility(ctx, ability); ok {
		return rendered, true, err
	}
	if rendered, ok, err := r.renderEquipRestrictedAbility(ctx, ability); ok {
		return rendered, true, err
	}
	if manaCost, ok := game.ActivatedBodyReconfigureCost(ability); ok &&
		reflect.DeepEqual(*ability, game.ReconfigureActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.ReconfigureActivatedAbility(%s)", renderedCost), true, nil
	}
	if manaCost, ok := game.ActivatedBodyCyclingCost(ability); ok &&
		reflect.DeepEqual(*ability, game.CyclingActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.CyclingActivatedAbility(%s)", renderedCost), true, nil
	}
	if manaCost, ok := game.ActivatedBodyScavengeCost(ability); ok &&
		reflect.DeepEqual(*ability, game.ScavengeActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.ScavengeActivatedAbility(%s)", renderedCost), true, nil
	}
	if manaCost, sourceManaValue, ok := game.ActivatedBodyTransmuteParams(ability); ok &&
		reflect.DeepEqual(*ability, game.TransmuteActivatedAbility(manaCost, sourceManaValue)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.TransmuteActivatedAbility(%s, %d)", renderedCost, sourceManaValue), true, nil
	}
	if manaCost, ok := game.ActivatedBodyUnearthCost(ability); ok &&
		reflect.DeepEqual(*ability, game.UnearthActivatedAbility(manaCost)) {
		renderedCost, err := r.renderManaCost(ctx, manaCost)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.UnearthActivatedAbility(%s)", renderedCost), true, nil
	}
	if power, ok := game.ActivatedBodySaddlePower(ability); ok &&
		reflect.DeepEqual(*ability, game.SaddleActivatedAbility(power)) {
		return fmt.Sprintf("game.SaddleActivatedAbility(%d)", power), true, nil
	}
	if power, ok := game.ActivatedBodyCrewPower(ability); ok &&
		reflect.DeepEqual(*ability, game.CrewActivatedAbility(power)) {
		return fmt.Sprintf("game.CrewActivatedAbility(%d)", power), true, nil
	}
	if manaCost, subtypes, ok := game.ActivatedBodyEternalizeParams(ability); ok &&
		reflect.DeepEqual(*ability, game.EternalizeActivatedBody(manaCost, subtypes...)) {
		return matchedRender(r.renderEternalizeFamilyAbility(ctx, "game.EternalizeActivatedBody", manaCost, subtypes))
	}
	if manaCost, subtypes, ok := game.ActivatedBodyEmbalmParams(ability); ok &&
		reflect.DeepEqual(*ability, game.EmbalmActivatedBody(manaCost, subtypes...)) {
		return matchedRender(r.renderEternalizeFamilyAbility(ctx, "game.EmbalmActivatedBody", manaCost, subtypes))
	}

	return "", false, nil
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

// renderEquipRestrictedCostReductionAbility renders a restricted Equip ability
// carrying a single cost modifier ("Equip Knight {2}. This ability costs {1}
// less to activate if ...") as the EquipRestrictedCostReductionActivatedAbility
// factory call. It mirrors renderEquipRestrictedAbility with the modifier
// appended.
func (r Renderer) renderEquipRestrictedCostReductionAbility(ctx *renderCtx, ability *game.ActivatedAbility) (string, bool, error) {
	manaCost, ok := game.ActivatedBodyEquipCost(ability)
	if !ok || len(ability.CostModifiers) != 1 {
		return "", false, nil
	}
	supertypes, subtypes, ok := equipRestrictionTypes(ability)
	if !ok || (len(supertypes) == 0 && len(subtypes) == 0) {
		return "", false, nil
	}
	if !reflect.DeepEqual(*ability, game.EquipRestrictedCostReductionActivatedAbility(manaCost, supertypes, subtypes, ability.CostModifiers[0])) {
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
	renderedModifier, err := r.renderCostModifier(ctx, ability.CostModifiers[0])
	if err != nil {
		return "", true, err
	}
	return fmt.Sprintf(
		"game.EquipRestrictedCostReductionActivatedAbility(%s, %s, %s, %s)",
		renderedCost, superLit, subLit, renderedModifier,
	), true, nil
}

func equipRestrictionTypes(ability *game.ActivatedAbility) ([]types.Super, []types.Sub, bool) {
	if len(ability.Content.Modes) != 1 || len(ability.Content.Modes[0].Targets) != 1 {
		return nil, nil, false
	}
	target := ability.Content.Modes[0].Targets[0]
	if !target.Selection.Exists {
		return nil, nil, false
	}
	selection := target.Selection.Val
	return selection.Supertypes, selection.SubtypesAny, true
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
	return r.renderGameManaAbility(ctx, *ability)
}

// constructorManaAbility spells a ManaAbility as a game constructor call when one
// produces exactly this value. See constructorRenderers in cardgen/cmd/genrender.
func (r Renderer) constructorManaAbility(ctx *renderCtx, v game.ManaAbility) (string, bool, error) {
	ability := &v

	for _, manaColor := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C} {
		if !reflect.DeepEqual(*ability, game.TapManaAbility(manaColor)) {
			continue
		}
		ctx.need(importMana)
		colorLiteral, err := renderManaColor(manaColor)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.TapManaAbility(%s)", colorLiteral), true, nil
	}
	if colors, ok := tapManaChoiceColors(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaChoiceAbility(colors...)) {
		ctx.need(importMana)
		colorLiterals := make([]string, 0, len(colors))
		for _, manaColor := range colors {
			colorLiteral, err := renderManaColor(manaColor)
			if err != nil {
				return "", false, err
			}
			colorLiterals = append(colorLiterals, colorLiteral)
		}
		return fmt.Sprintf("game.TapManaChoiceAbility(%s)", strings.Join(colorLiterals, ", ")), true, nil
	}
	if colors, count, ok := tapManaChoiceCountColors(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaChoiceCountAbility(ability.Text, count, colors...)) {
		ctx.need(importMana)
		colorLiterals := make([]string, 0, len(colors))
		for _, manaColor := range colors {
			colorLiteral, err := renderManaColor(manaColor)
			if err != nil {
				return "", false, err
			}
			colorLiterals = append(colorLiterals, colorLiteral)
		}
		return fmt.Sprintf("game.TapManaChoiceCountAbility(%q, %d, %s)", ability.Text, count, strings.Join(colorLiterals, ", ")), true, nil
	}
	if reflect.DeepEqual(*ability, game.TapChosenColorManaAbility(ability.Text)) {
		return fmt.Sprintf("game.TapChosenColorManaAbility(%q)", ability.Text), true, nil
	}
	for _, fixed := range []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G, mana.C} {
		if reflect.DeepEqual(*ability, game.TapFixedOrChosenColorManaAbility(ability.Text, fixed)) {
			ctx.need(importMana)
			colorLiteral, err := renderManaColor(fixed)
			if err != nil {
				return "", false, err
			}
			return fmt.Sprintf("game.TapFixedOrChosenColorManaAbility(%q, %s)", ability.Text, colorLiteral), true, nil
		}
	}
	if reflect.DeepEqual(*ability, game.TapManaCommanderIdentityAbility()) {
		return "game.TapManaCommanderIdentityAbility()", true, nil
	}
	for _, relation := range []game.PlayerRelation{game.PlayerYou, game.PlayerOpponent} {
		for _, includeColorless := range []bool{false, true} {
			if reflect.DeepEqual(*ability, game.TapManaLandsProduceAbility(relation, includeColorless)) {
				literal, err := renderPlayerRelation(relation)
				if err != nil {
					return "", false, err
				}
				return fmt.Sprintf("game.TapManaLandsProduceAbility(%s, %t)", literal, includeColorless), true, nil
			}
		}
	}
	if linkID, ok := linkedExileColorManaLinkID(ability); ok &&
		reflect.DeepEqual(*ability, game.TapLinkedExileColorManaAbility(linkID)) {
		return fmt.Sprintf("game.TapLinkedExileColorManaAbility(%q)", linkID), true, nil
	}
	if selection, ok := amongControlledColorsSelection(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaAmongControlledColorsAbility(ability.Text, selection)) {
		rendered, err := r.renderSelection(ctx, selection)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.TapManaAmongControlledColorsAbility(%q, %s)", ability.Text, rendered), true, nil
	}
	if selection, ok := eachControlledColorSelection(ability); ok &&
		reflect.DeepEqual(*ability, game.TapManaEachControlledColorAbility(ability.Text, selection)) {
		rendered, err := r.renderSelection(ctx, selection)
		if err != nil {
			return "", false, err
		}
		return fmt.Sprintf("game.TapManaEachControlledColorAbility(%q, %s)", ability.Text, rendered), true, nil
	}

	if game.IsTapSacrificeAnyOneColorManaAbility(ability) {
		_, count, ok := game.ManaAbilityChoiceOutput(ability)
		if ok {
			return fmt.Sprintf("game.TapSacrificeAnyOneColorManaAbility(%q, %d)", ability.Text, count), true, nil
		}
	}

	return "", false, nil
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

// renderMobilizeAmount renders a game.MobilizeAmount composite literal for the
// Mobilize keyword body: the fixed printed count or the typed dynamic count.
func renderMobilizeAmount(amount game.MobilizeAmount) string {
	switch amount.Dynamic {
	case game.MobilizeDynamicCreatureCardsInGraveyard:
		return "game.MobilizeAmount{Dynamic: game.MobilizeDynamicCreatureCardsInGraveyard}"
	default:
		return fmt.Sprintf("game.MobilizeAmount{Fixed: %d}", amount.Fixed)
	}
}

func (r Renderer) renderTriggeredAbility(ctx *renderCtx, ability *game.TriggeredAbility) (string, error) {
	return r.renderGameTriggeredAbility(ctx, *ability)
}

// constructorTriggeredAbility spells a TriggeredAbility as a game constructor call when one
// produces exactly this value. See constructorRenderers in cardgen/cmd/genrender.
func (r Renderer) constructorTriggeredAbility(ctx *renderCtx, v game.TriggeredAbility) (string, bool, error) {
	ability := &v

	if keyword, ok := game.BodyKeywordAbility(ability, game.CumulativeUpkeep); ok {
		if cumulative, ok := keyword.(game.CumulativeUpkeepKeyword); ok &&
			reflect.DeepEqual(*ability, game.CumulativeUpkeepTriggeredAbility(cumulative.Cost)) {
			renderedCost, err := r.renderManaCost(ctx, cumulative.Cost)
			if err != nil {
				return "", false, err
			}
			return fmt.Sprintf("game.CumulativeUpkeepTriggeredAbility(%s)", renderedCost), true, nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Echo); ok {
		if echo, ok := keyword.(game.EchoKeyword); ok &&
			reflect.DeepEqual(*ability, game.EchoTriggeredAbility(echo.Cost)) {
			renderedCost, err := r.renderManaCost(ctx, echo.Cost)
			if err != nil {
				return "", false, err
			}
			return fmt.Sprintf("game.EchoTriggeredAbility(%s)", renderedCost), true, nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Fabricate); ok {
		if fabricate, ok := keyword.(game.FabricateKeyword); ok &&
			reflect.DeepEqual(*ability, game.FabricateTriggeredAbility(fabricate.Count)) {
			return fmt.Sprintf("game.FabricateTriggeredAbility(%d)", fabricate.Count), true, nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Hideaway); ok {
		if hideaway, ok := keyword.(game.HideawayKeyword); ok &&
			reflect.DeepEqual(*ability, game.HideawayTriggeredAbility(hideaway.Amount)) {
			return fmt.Sprintf("game.HideawayTriggeredAbility(%d)", hideaway.Amount), true, nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Soulshift); ok {
		if soulshift, ok := keyword.(game.SoulshiftKeyword); ok &&
			reflect.DeepEqual(*ability, game.SoulshiftTriggeredAbility(soulshift.Count)) {
			return fmt.Sprintf("game.SoulshiftTriggeredAbility(%d)", soulshift.Count), true, nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Rampage); ok {
		if rampage, ok := keyword.(game.RampageKeyword); ok &&
			reflect.DeepEqual(*ability, game.RampageTriggeredAbility(rampage.Count)) {
			return fmt.Sprintf("game.RampageTriggeredAbility(%d)", rampage.Count), true, nil
		}
	}
	if keyword, ok := game.BodyKeywordAbility(ability, game.Mobilize); ok {
		if mobilize, ok := keyword.(game.MobilizeKeyword); ok &&
			reflect.DeepEqual(*ability, game.MobilizeTriggeredBody(mobilize.Amount)) {
			return fmt.Sprintf("game.MobilizeTriggeredBody(%s)", renderMobilizeAmount(mobilize.Amount)), true, nil
		}
	}
	if reflect.DeepEqual(*ability, game.UndyingTriggeredBody) {
		return "game.UndyingTriggeredBody", true, nil
	}
	if reflect.DeepEqual(*ability, game.PersistTriggeredBody) {
		return "game.PersistTriggeredBody", true, nil
	}
	if reflect.DeepEqual(*ability, game.DethroneTriggeredBody) {
		return "game.DethroneTriggeredBody", true, nil
	}
	if reflect.DeepEqual(*ability, game.StartEnginesTriggeredBody) {
		return "game.StartEnginesTriggeredBody", true, nil
	}
	if reflect.DeepEqual(*ability, game.FlankingTriggeredBody) {
		return "game.FlankingTriggeredBody", true, nil
	}
	if reflect.DeepEqual(*ability, game.TrainingTriggeredBody) {
		return "game.TrainingTriggeredBody", true, nil
	}
	if reflect.DeepEqual(*ability, game.MyriadTriggeredBody) {
		return "game.MyriadTriggeredBody", true, nil
	}
	if reflect.DeepEqual(*ability, game.LivingWeaponTriggeredAbility()) {
		return "game.LivingWeaponTriggeredAbility()", true, nil
	}
	if reflect.DeepEqual(*ability, game.RavenousDrawTriggeredAbility()) {
		return "game.RavenousDrawTriggeredAbility()", true, nil
	}
	if reflect.DeepEqual(*ability, game.OffspringEnterTriggeredAbility()) {
		return "game.OffspringEnterTriggeredAbility()", true, nil
	}
	if reflect.DeepEqual(*ability, game.EvokeSacrificeTriggeredAbility()) {
		return "game.EvokeSacrificeTriggeredAbility()", true, nil
	}
	if reflect.DeepEqual(*ability, game.DashTriggeredAbility()) {
		return "game.DashTriggeredAbility()", true, nil
	}
	return "", false, nil
}

func (r Renderer) renderChapterAbility(ctx *renderCtx, ability *game.ChapterAbility) (string, error) {
	return r.renderGameChapterAbility(ctx, *ability)
}

func (r Renderer) renderLoyaltyAbility(ctx *renderCtx, ability *game.LoyaltyAbility) (string, error) {
	return r.renderGameLoyaltyAbility(ctx, *ability)
}

// validateTriggerPatternCardSelection validates CardSelection constraints for a
// TriggerPattern and returns an error if they are unsupported. Spell-cast
// triggers read full card characteristics from the event, while draw, discard,
// and cycle triggers read the moved card's types, supertypes, subtypes, and
// colors from the moved card instance (CR 603.2).
func validateTriggerPatternCardSelection(pattern *game.TriggerPattern) error {
	if pattern.CardSelection.Empty() {
		return nil
	}
	switch pattern.Event {
	case game.EventSpellCast, game.EventCardDiscarded, game.EventCardDrawn, game.EventCycled:
	default:
		return errors.New("render: CardSelection is only supported for EventSpellCast, EventCardDiscarded, EventCardDrawn, and EventCycled trigger patterns")
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
		unsupported.ManaValueLessThanSourcePower = false
		unsupported.ColorChoice = game.ColorChoiceNone
	}
	if pattern.Event == game.EventCardDiscarded || pattern.Event == game.EventCardDrawn || pattern.Event == game.EventCycled {
		unsupported.Supertypes = nil
		unsupported.SubtypesAny = nil
		unsupported.ExcludedSubtype = ""
		unsupported.ColorsAny = nil
		unsupported.Colorless = false
		unsupported.Multicolored = false
	}
	if !unsupported.Empty() {
		return errors.New("render: unsupported CardSelection fields in trigger pattern")
	}
	return nil
}

// matchedRender adapts a hand-written renderer that always matches to the
// (rendered, matched, err) shape the constructor hooks return.
func matchedRender(rendered string, err error) (string, bool, error) {
	return rendered, true, err
}
