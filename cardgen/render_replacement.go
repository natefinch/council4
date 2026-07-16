package cardgen

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// classLevelReplacementGate reports whether the condition is exactly a Class
// level gate (SourceClassLevelAtLeast with no other predicate). Such gates are
// rendered generically by wrapping the base replacement in
// game.ClassLevelGatedReplacement so any replacement category can be gated by a
// Class level without a category-specific flag.
func classLevelReplacementGate(condition opt.V[game.Condition]) (int, bool) {
	if !condition.Exists {
		return 0, false
	}
	cond := condition.Val
	level := cond.SourceClassLevelAtLeast
	if level <= 0 {
		return 0, false
	}
	cond.SourceClassLevelAtLeast = 0
	if !cond.Empty() {
		return 0, false
	}
	return level, true
}

func (r Renderer) renderReplacementAbility(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	if reflect.DeepEqual(*ability, game.RavenousEntersWithCountersReplacement()) {
		return "game.RavenousEntersWithCountersReplacement()", nil
	}
	if level, ok := classLevelReplacementGate(ability.Replacement.Condition); ok {
		inner := *ability
		inner.Replacement.Condition = opt.V[game.Condition]{}
		rendered, err := r.renderReplacementAbility(ctx, &inner)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.ClassLevelGatedReplacement(%s, %d)", rendered, level), nil
	}
	if ability.Replacement.EntersAsCopy {
		return r.renderEntersAsCopyReplacement(ctx, ability)
	}
	if ability.Replacement.EntryDevourMultiplier > 0 &&
		(ability.Replacement.EntryDevourType != "" || ability.Replacement.EntryDevourSubtype != "") {
		return renderTypedDevourReplacement(ctx, ability)
	}
	if rendered, handled := renderStringReplacement(ability); handled {
		return rendered, nil
	}
	if ability.Replacement.SpellCopyAddend > 0 {
		return fmt.Sprintf(
			"game.AdditionalSpellCopyReplacement(%q, %d, %t)",
			ability.Text,
			ability.Replacement.SpellCopyAddend,
			ability.Replacement.SpellCopyAdditionalMayChooseNewTargets,
		), nil
	}
	if ability.Replacement.EntersBecomesCharacteristic {
		return r.renderGroupEntersBecomesReplacement(ctx, ability)
	}
	if ability.Replacement.EntersTappedOthers {
		return r.renderGroupEntersTappedReplacement(ctx, ability)
	}
	if ability.Replacement.EntersUntappedOthers {
		return r.renderGroupEntersUntappedReplacement(ctx, ability)
	}
	if ability.Replacement.EntersWithCountersOthers {
		return r.renderGroupEntersWithCountersReplacement(ctx, ability)
	}
	if len(ability.Replacement.EntersWithCounters) > 0 {
		if ability.UnlessPaid.Exists {
			return "", errors.New("render: ETB counter replacement cannot also require payment")
		}
		if ability.Replacement.EntersTapped && ability.Replacement.Condition.Exists {
			return "", errors.New("render: ETB counter replacement cannot both tap and have a condition")
		}
		placements, err := r.renderCounterPlacements(ctx, ability.Replacement.EntersWithCounters)
		if err != nil {
			return "", err
		}
		placementList := strings.Join(placements, ", ")
		if ability.Replacement.EntersTapped {
			return fmt.Sprintf("game.EntersTappedWithCountersReplacement(%q, %s)", ability.Text, placementList), nil
		}
		if ability.Replacement.Condition.Exists {
			condStr, err := r.renderConditionForETBReplacement(ctx, &ability.Replacement.Condition.Val)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.EntersWithCountersIfReplacement(%q, %s, %s)", ability.Text, condStr, placementList), nil
		}
		return fmt.Sprintf("game.EntersWithCountersReplacement(%q, %s)", ability.Text, placementList), nil
	}
	if ability.Replacement.EntersTapped && ability.Replacement.EntryColorChoice {
		if ability.UnlessPaid.Exists || ability.Replacement.Condition.Exists {
			return "", errors.New("render: enters-tapped color-choice replacement cannot also require payment or have a condition")
		}
		if ability.Replacement.EntryColorChoiceExclude != "" {
			ctx.need(importMana)
			colorLiteral, err := renderManaColor(ability.Replacement.EntryColorChoiceExclude)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.EntersTappedColorChoiceExcludingReplacement(%q, %s)", ability.Text, colorLiteral), nil
		}
		return fmt.Sprintf("game.EntersTappedColorChoiceReplacement(%q)", ability.Text), nil
	}
	if ability.Replacement.EntersTapped && ability.UnlessPaid.Exists {
		if ability.Replacement.Condition.Exists {
			return "", errors.New("render: paid ETB replacement cannot also have a condition")
		}
		payment, err := r.renderResolutionPayment(ctx, ability.UnlessPaid.Val)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EntersTappedUnlessPaidReplacement(%q, %s)", ability.Text, payment), nil
	}
	if ability.Replacement.EntersTapped && !ability.UnlessPaid.Exists {
		if !ability.Replacement.Condition.Exists {
			return fmt.Sprintf("game.EntersTappedReplacement(%q)", ability.Text), nil
		}
		condStr, err := r.renderConditionForETBReplacement(ctx, &ability.Replacement.Condition.Val)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.EntersTappedIfReplacement(%q, %s)", ability.Text, condStr), nil
	}
	if ability.Replacement.EntryColorChoice {
		if ability.Replacement.EntryColorChoiceExclude != "" {
			ctx.need(importMana)
			colorLiteral, err := renderManaColor(ability.Replacement.EntryColorChoiceExclude)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("game.EntryColorChoiceExcludingReplacement(%q, %s)", ability.Text, colorLiteral), nil
		}
		return fmt.Sprintf("game.EntryColorChoiceReplacement(%q)", ability.Text), nil
	}
	if ability.Replacement.AttachCardNameChoiceType != "" || ability.Replacement.AttachSubtypeChoiceType != "" {
		if ability.Replacement.AttachCardNameChoiceType == "" || ability.Replacement.AttachSubtypeChoiceType == "" {
			return "", errors.New("render: attachment choices require both card-name and subtype card types")
		}
		nameType, err := cardTypeLiteral(ability.Replacement.AttachCardNameChoiceType)
		if err != nil {
			return "", err
		}
		subtypeType, err := cardTypeLiteral(ability.Replacement.AttachSubtypeChoiceType)
		if err != nil {
			return "", err
		}
		ctx.need(importTypes)
		return fmt.Sprintf("game.AttachmentChoicesReplacement(%q, %s, %s)", ability.Text, nameType, subtypeType), nil
	}
	if ability.Replacement.EntryTypeChoice {
		return fmt.Sprintf("game.EntryTypeChoiceReplacement(%q)", ability.Text), nil
	}
	if ability.Replacement.EntryCardTypeChoice {
		return fmt.Sprintf("game.EntryCardTypeChoiceReplacement(%q)", ability.Text), nil
	}
	if ability.Replacement.ContinuousZoneRedirect {
		return r.renderGraveyardRedirectReplacement(ctx, ability)
	}
	if ability.Replacement.ReplaceToZone != zone.None && ability.UnlessPaid.Exists {
		if ability.Replacement.EntersTapped || ability.Replacement.Condition.Exists {
			return "", errors.New("render: optional entry zone replacement cannot also enter tapped or have a condition")
		}
		payment, err := r.renderResolutionPayment(ctx, ability.UnlessPaid.Val)
		if err != nil {
			return "", err
		}
		replaceToZone, err := renderZone(ability.Replacement.ReplaceToZone)
		if err != nil {
			return "", err
		}
		ctx.need(importZone)
		return fmt.Sprintf("game.EntersUnlessPaidElseZoneReplacement(%q, %s, %s)", ability.Text, payment, replaceToZone), nil
	}
	if ability.Replacement.ReplaceToZone != zone.None {
		replacement, err := renderZoneDestinationReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if ability.Replacement.TokenMultiplier > 0 || ability.Replacement.TokenAddend != 0 {
		replacement, err := renderTokenCreationReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if ability.Replacement.DamagePreventAmount > 0 {
		replacement, err := renderDamagePreventionReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if ability.Replacement.DamageMultiplier > 0 || ability.Replacement.DamageAddend != 0 {
		replacement, err := renderDamageReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if ability.Replacement.CounterMultiplier > 0 || ability.Replacement.CounterAddend != 0 {
		replacement, err := renderCounterPlacementReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if len(ability.Replacement.CreateOneOfEachTokens) > 0 {
		replacement, err := r.renderNamedTokenSetReplacement(ctx, ability)
		if err != nil {
			return "", err
		}
		return replacement, nil
	}
	if ability.Replacement.DamageRecipientSelection != nil {
		return r.renderCombatDamagePreventionToGroupReplacement(ctx, ability)
	}
	if ability.Replacement.DamagePreventAll {
		return r.renderDamagePreventionToCountersReplacement(ctx, ability)
	}
	return "", fmt.Errorf("render: unsupported replacement ability %q", ability.Text)
}

// renderCombatDamagePreventionToGroupReplacement renders the continuous static
// "Prevent all combat damage that would be dealt to <group>." (Goldbug,
// Humanity's Ally) into a game.CombatDamagePreventionToGroupReplacement call
// carrying the recipient group selection.
func (r Renderer) renderCombatDamagePreventionToGroupReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	selection, err := r.renderSelection(ctx, *ability.Replacement.DamageRecipientSelection)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("game.CombatDamagePreventionToGroupReplacement(%q, %s)", ability.Text, selection), nil
}

// renderDamagePreventionToCountersReplacement renders the continuous static "If
// <permanent> would be dealt damage, prevent that damage and put that many
// +1/+1 counters on it." into a game.DamagePreventionToPlusOneCountersReplacement
// call carrying the attached-vs-self recipient and the optional gating condition.
func (r Renderer) renderDamagePreventionToCountersReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	ctx.need(importOpt)
	condition := "opt.V[game.Condition]{}"
	if replacement.Condition.Exists {
		cond := replacement.Condition.Val
		rendered, err := r.renderControllerControlsCondition(ctx, &cond, "prevent-damage-to-counters replacement")
		if err != nil {
			return "", err
		}
		condition = fmt.Sprintf("opt.Val(%s)", rendered)
	}
	return fmt.Sprintf("%s(%q, %t, %s)",
		damagePreventionReplacementFunc(replacement), ability.Text, replacement.DamageRecipientAttached, condition), nil
}

// damagePreventionReplacementFunc names the constructor for a prevent-all
// damage replacement: the Phantom "remove a +1/+1 counter" form or the default
// "put that many +1/+1 counters" form.
func damagePreventionReplacementFunc(replacement game.ReplacementEffect) string {
	if replacement.DamagePreventedRemovesPlusOneCounter {
		return "game.DamagePreventionRemovesCounterReplacement"
	}
	return "game.DamagePreventionToPlusOneCountersReplacement"
}

// renderLifeModifierReplacement renders the life-gain and life-loss
// value-modifying replacements (Boon Reflection, Bloodletter of Aclazotz),
// reporting handled=false when the replacement modifies neither life event.
func renderLifeModifierReplacement(ability *game.ReplacementAbility) (string, bool) {
	if ability.Replacement.LifeGainMultiplier > 1 || ability.Replacement.LifeGainAddend != 0 {
		return fmt.Sprintf("game.LifeGainReplacement(%q, %d, %d)",
			ability.Text, ability.Replacement.LifeGainMultiplier, ability.Replacement.LifeGainAddend), true
	}
	if ability.Replacement.LifeLossMultiplier > 1 || ability.Replacement.LifeLossAddend != 0 {
		return fmt.Sprintf("game.LifeLossReplacement(%q, %d, %d, %t, %t)",
			ability.Text, ability.Replacement.LifeLossMultiplier, ability.Replacement.LifeLossAddend,
			ability.Replacement.LifeLossRecipientOpponent, ability.Replacement.LifeLossDuringControllerTurn), true
	}
	return "", false
}

// renderStringReplacement renders the replacements that depend only on the
// ability text plus a few scalar parameters (the creature-form Devour,
// draw-from-empty-library win, draw multiplier, and the life-modifier family),
// reporting handled=false when the replacement is none of these. Typed Devour
// variants are rendered separately by renderTypedDevourReplacement, which is
// dispatched before this function so the creature form's two-argument call here
// stays unchanged.
func renderStringReplacement(ability *game.ReplacementAbility) (string, bool) {
	if ability.Replacement.EntryDevourMultiplier > 0 {
		return fmt.Sprintf("game.DevourReplacement(%q, %d)", ability.Text, ability.Replacement.EntryDevourMultiplier), true
	}
	if ability.Replacement.EntryTributeCount > 0 {
		return fmt.Sprintf("game.TributeReplacement(%q, %d)", ability.Text, ability.Replacement.EntryTributeCount), true
	}
	if ability.Replacement.DrawFromEmptyLibraryWins {
		return fmt.Sprintf("game.DrawFromEmptyLibraryWinReplacement(%q)", ability.Text), true
	}
	if ability.Replacement.DrawCardMultiplier > 1 {
		if ability.Replacement.Condition.Exists {
			if ability.Replacement.Condition.Val.ControllerHasMaxSpeed {
				return fmt.Sprintf("game.MaxSpeedDrawCardMultiplierReplacement(%q, %d, %t)",
					ability.Text, ability.Replacement.DrawCardMultiplier, ability.Replacement.DrawCardExceptFirstInDrawStep), true
			}
			return "", false
		}
		return fmt.Sprintf("game.DrawCardMultiplierReplacement(%q, %d, %t)",
			ability.Text, ability.Replacement.DrawCardMultiplier, ability.Replacement.DrawCardExceptFirstInDrawStep), true
	}
	if ability.Replacement.DrawCardDigLook > 0 {
		return fmt.Sprintf("game.DrawCardDigReplacement(%q, %d, %d, %s)",
			ability.Text, ability.Replacement.DrawCardDigLook, ability.Replacement.DrawCardDigTake,
			renderDigRemainder(ability.Replacement.DrawCardDigRemainder)), true
	}
	return renderLifeModifierReplacement(ability)
}

// renderTypedDevourReplacement renders the typed Devour variants ("Devour
// artifact N", "Devour land N", "Devour Food N") to their dedicated game
// constructors, emitting the sacrificed permanent's card type or subtype literal
// and requiring the types import.
func renderTypedDevourReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	ctx.need(importTypes)
	if ability.Replacement.EntryDevourSubtype != "" {
		literal := SubtypeToLiteral(string(ability.Replacement.EntryDevourSubtype),
			[]string{"Artifact", "Creature", "Land", "Enchantment", "Planeswalker", "Battle"})
		return fmt.Sprintf("game.DevourSubtypeReplacement(%q, %d, %s)",
			ability.Text, ability.Replacement.EntryDevourMultiplier, literal), nil
	}
	literal, err := cardTypeLiteral(ability.Replacement.EntryDevourType)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("game.DevourTypeReplacement(%q, %d, %s)",
		ability.Text, ability.Replacement.EntryDevourMultiplier, literal), nil
}

// renderDigRemainder renders the DigRemainder destination constant for a
// draw-replacement dig. The graveyard default and the library-bottom variant are
// the only runtime placements.
func renderDigRemainder(remainder game.DigRemainder) string {
	if remainder == game.DigRemainderLibraryBottom {
		return "game.DigRemainderLibraryBottom"
	}
	return "game.DigRemainderGraveyard"
}

// renderGroupEntersTappedReplacement renders a continuous static enters-tapped
// replacement that taps a group of OTHER permanents as they enter (Authority of
// the Consuls family).
func (Renderer) renderGroupEntersTappedReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if !replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists {
		return "", errors.New("render: unsupported group enters-tapped replacement shape")
	}

	controller, err := renderGroupEntersTappedController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	if replacement.EntersTappedSelection == nil {
		return fmt.Sprintf("game.EntersTappedGroupReplacement(%q, %s)", ability.Text, controller), nil
	}
	ctx.need(importTypes)
	recipientTypes := replacement.EntersTappedSelection.RequiredTypesAny
	typeLiterals := make([]string, 0, len(recipientTypes))
	for _, cardType := range recipientTypes {
		literal, err := cardTypeLiteral(cardType)
		if err != nil {
			return "", err
		}
		typeLiterals = append(typeLiterals, literal)
	}
	return fmt.Sprintf("game.EntersTappedGroupReplacement(%q, %s, %s)",
		ability.Text, controller, strings.Join(typeLiterals, ", ")), nil
}

// renderGroupEntersBecomesReplacement renders a continuous static group ETB
// characteristic replacement that gives entering permanents new characteristics
// ("As a historic permanent you control enters, it becomes a 7/7 Dinosaur
// creature in addition to its other types." — Displaced Dinosaurs).
func (Renderer) renderGroupEntersBecomesReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if !replacement.EntersBecomesCharacteristic ||
		replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists {
		return "", errors.New("render: unsupported group enters-becomes replacement shape")
	}

	controller, err := renderGroupEntersTappedController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}

	historic, subjectTypes := decodeEntersBecomesSelection(replacement.EntersBecomesSelection)

	fields := []string{fmt.Sprintf("Controller: %s", controller)}
	if historic {
		fields = append(fields, "Historic: true")
	}
	if len(subjectTypes) > 0 {
		ctx.need(importTypes)
		literals, err := cardTypeLiterals(subjectTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("SubjectTypes: []types.Card{%s}", literals))
	}
	if len(replacement.EntersBecomesAddTypes) > 0 {
		ctx.need(importTypes)
		literals, err := cardTypeLiterals(replacement.EntersBecomesAddTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AddTypes: []types.Card{%s}", literals))
	}
	if len(replacement.EntersBecomesAddSubtypes) > 0 {
		subtypes, err := renderSubtypeSlice(ctx, replacement.EntersBecomesAddSubtypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AddSubtypes: %s", subtypes))
	}
	if len(replacement.EntersBecomesAddColors) > 0 {
		ctx.need(importColor)
		literals, err := colorValueLiterals(replacement.EntersBecomesAddColors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AddColors: []color.Color{%s}", literals))
	}
	if replacement.EntersBecomesBasePower.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("BasePower: opt.Val(%d)", replacement.EntersBecomesBasePower.Val))
	}
	if replacement.EntersBecomesBaseToughness.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("BaseToughness: opt.Val(%d)", replacement.EntersBecomesBaseToughness.Val))
	}

	return fmt.Sprintf("game.EntersBecomesGroupReplacement(%q, game.EntersBecomesGroupParams{%s})",
		ability.Text, strings.Join(fields, ", ")), nil
}

// decodeEntersBecomesSelection recovers the Historic flag and SubjectTypes filter
// that EntersBecomesGroupReplacement encoded into the entrant selection so the
// renderer can reconstruct the original params.
func decodeEntersBecomesSelection(selection *game.Selection) (historic bool, subjectTypes []types.Card) {
	if selection == nil {
		return false, nil
	}
	historic = len(selection.AnyOf) > 0
	return historic, selection.RequiredTypes
}

// cardTypeLiterals renders a comma-separated list of card type literals.
func cardTypeLiterals(cardTypes []types.Card) (string, error) {
	literals := make([]string, 0, len(cardTypes))
	for _, cardType := range cardTypes {
		literal, err := cardTypeLiteral(cardType)
		if err != nil {
			return "", err
		}
		literals = append(literals, literal)
	}
	return strings.Join(literals, ", "), nil
}

func (Renderer) renderGroupEntersUntappedReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if !replacement.EntersUntapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists {
		return "", errors.New("render: unsupported group enters-untapped replacement shape")
	}
	controller, err := renderGroupEntersTappedController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	if replacement.EntersTappedSelection == nil {
		return fmt.Sprintf("game.EntersUntappedGroupReplacement(%q, %s)", ability.Text, controller), nil
	}
	ctx.need(importTypes)
	recipientTypes := replacement.EntersTappedSelection.RequiredTypesAny
	typeLiterals := make([]string, 0, len(recipientTypes))
	for _, cardType := range recipientTypes {
		literal, err := cardTypeLiteral(cardType)
		if err != nil {
			return "", err
		}
		typeLiterals = append(typeLiterals, literal)
	}
	return fmt.Sprintf("game.EntersUntappedGroupReplacement(%q, %s, %s)",
		ability.Text, controller, strings.Join(typeLiterals, ", ")), nil
}

// renderGroupEntersWithCountersReplacement renders a continuous static
// enters-with-counters replacement that adds counters to a group of OTHER
// permanents as they enter (Tayam, Luminous Enigma family).
func (r Renderer) renderGroupEntersWithCountersReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersWithCountersRecipient == nil ||
		replacement.EntersTapped ||
		len(replacement.EntersWithCounters) == 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists {
		return "", errors.New("render: unsupported group enters-with-counters replacement shape")
	}
	placements, err := r.renderCounterPlacements(ctx, replacement.EntersWithCounters)
	if err != nil {
		return "", err
	}
	recipient, err := r.renderSelection(ctx, *replacement.EntersWithCountersRecipient)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("game.EntersWithCountersGroupReplacement(%q, &%s, %s)",
		ability.Text, recipient, strings.Join(placements, ", ")), nil
}

// renderGroupEntersTappedController renders the trigger-controller filter for a
// group enters-tapped replacement, including the each-player (Any) scope that the
// strict renderTriggerController rejects.
func renderGroupEntersTappedController(controller game.TriggerControllerFilter) (string, error) {
	switch controller {
	case game.TriggerControllerAny:
		return "game.TriggerControllerAny", nil
	case game.TriggerControllerYou:
		return "game.TriggerControllerYou", nil
	case game.TriggerControllerOpponent:
		return "game.TriggerControllerOpponent", nil
	default:
		return "", fmt.Errorf("render: unsupported trigger controller filter %d", controller)
	}
}

func renderDamageReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventDamageDealt ||
		(replacement.DamageMultiplier <= 1 && replacement.DamageAddend == 0) {
		return "", errors.New("render: unsupported damage replacement shape")
	}
	if replacement.ControllerFilter == game.TriggerControllerAny ||
		replacement.DamageRecipientOpponent ||
		replacement.DamageRecipientOpponentPlayerOnly ||
		replacement.DamageNoncombatOnly ||
		len(replacement.DamageSourceTypes) > 0 {
		return renderFilteredDamageReplacement(ctx, ability)
	}
	controller, err := renderTriggerController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	colors := "nil"
	if len(replacement.DamageSourceColors) > 0 {
		colors, err = renderColorSlice(ctx, replacement.DamageSourceColors)
		if err != nil {
			return "", err
		}
	}
	constructor := "game.DamageReplacement"
	if replacement.DamageExcludeSource {
		constructor = "game.DamageReplacementExcludingSource"
	}
	return fmt.Sprintf("%s(%q, %d, %d, %s, %s)",
		constructor,
		ability.Text,
		replacement.DamageMultiplier,
		replacement.DamageAddend,
		colors,
		controller,
	), nil
}

// renderFilteredDamageReplacement renders a damage replacement carrying the
// opponent-recipient, noncombat, card-type source, or any-controller filters
// that the legacy DamageReplacement constructor cannot express.
func renderFilteredDamageReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	controller, err := renderGroupEntersTappedController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	colors := "nil"
	if len(replacement.DamageSourceColors) > 0 {
		colors, err = renderColorSlice(ctx, replacement.DamageSourceColors)
		if err != nil {
			return "", err
		}
	}
	cardTypes := "nil"
	if len(replacement.DamageSourceTypes) > 0 {
		cardTypes, err = renderTypesCardSlice(ctx, replacement.DamageSourceTypes)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("game.DamageReplacementFiltered(%q, &game.DamageReplacementSpec{Multiplier: %d, Addend: %d, SourceColors: %s, SourceTypes: %s, ExcludeSource: %t, RecipientOpponent: %t, RecipientOpponentPlayerOnly: %t, NoncombatOnly: %t, Controller: %s})",
		ability.Text,
		replacement.DamageMultiplier,
		replacement.DamageAddend,
		colors,
		cardTypes,
		replacement.DamageExcludeSource,
		replacement.DamageRecipientOpponent,
		replacement.DamageRecipientOpponentPlayerOnly,
		replacement.DamageNoncombatOnly,
		controller,
	), nil
}

// renderDamagePreventionReplacement renders a continuous static damage
// prevention replacement (Sphere of Law, Urza's Armor, Protection of the Hekma)
// into a game.DamagePreventionReplacement call carrying the fixed amount, the
// source color and card-type filters, and the opponent-source restriction.
func renderDamagePreventionReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventDamageDealt ||
		!replacement.DamageRecipientController ||
		replacement.DamagePreventAmount <= 0 {
		return "", errors.New("render: unsupported damage prevention replacement shape")
	}
	colors := "nil"
	if len(replacement.DamageSourceColors) > 0 {
		var err error
		colors, err = renderColorSlice(ctx, replacement.DamageSourceColors)
		if err != nil {
			return "", err
		}
	}
	cardTypes := "nil"
	if len(replacement.DamageSourceTypes) > 0 {
		var err error
		cardTypes, err = renderTypesCardSlice(ctx, replacement.DamageSourceTypes)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("game.DamagePreventionReplacement(%q, &game.DamagePreventionSpec{Amount: %d, SourceColors: %s, SourceTypes: %s, SourceControllerOpponent: %t})",
		ability.Text,
		replacement.DamagePreventAmount,
		colors,
		cardTypes,
		replacement.DamageSourceControllerOpponent,
	), nil
}

func renderCounterPlacementReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventCountersAdded ||
		(replacement.CounterMultiplier <= 1 && replacement.CounterAddend == 0) {
		return "", errors.New("render: unsupported counter-placement replacement shape")
	}
	controller, err := renderGroupEntersTappedController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	if replacement.CounterRecipientSelf {
		if !replacement.MatchCounterKind {
			return "", errors.New("render: self counter-placement replacement requires a counter kind")
		}
		kind, err := renderCounterKind(replacement.CounterKindFilter)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		return fmt.Sprintf("game.SelfCounterPlacementReplacement(%q, %d, %d, %s)",
			ability.Text,
			replacement.CounterMultiplier,
			replacement.CounterAddend,
			kind,
		), nil
	}
	if sel := replacement.CounterRecipientSelection; sel != nil && sel.ExcludeSource {
		selLit, err := (Renderer{}).renderSelection(ctx, *sel)
		if err != nil {
			return "", err
		}
		if replacement.MatchCounterKind {
			kind, err := renderCounterKind(replacement.CounterKindFilter)
			if err != nil {
				return "", err
			}
			ctx.need(importCounter)
			return fmt.Sprintf("game.ControlledPermanentSelectionCounterKindPlacementReplacement(%q, %d, %d, %s, %s, %s)",
				ability.Text,
				replacement.CounterMultiplier,
				replacement.CounterAddend,
				kind,
				selLit,
				controller,
			), nil
		}
		return fmt.Sprintf("game.ControlledPermanentSelectionCounterPlacementReplacement(%q, %d, %d, %s, %s)",
			ability.Text,
			replacement.CounterMultiplier,
			replacement.CounterAddend,
			selLit,
			controller,
		), nil
	}
	if sel := replacement.CounterRecipientSelection; sel != nil && len(sel.RequiredTypesAny) > 0 {
		ctx.need(importTypes)
		typeLiterals := make([]string, 0, len(sel.RequiredTypesAny))
		for _, cardType := range sel.RequiredTypesAny {
			literal, err := cardTypeLiteral(cardType)
			if err != nil {
				return "", err
			}
			typeLiterals = append(typeLiterals, literal)
		}
		typesArg := fmt.Sprintf("[]types.Card{%s}", strings.Join(typeLiterals, ", "))
		if replacement.CounterRecipientControllerPlayer {
			if replacement.MatchCounterKind {
				return "", errors.New("render: controller-player counter-placement recipient does not support a specific counter kind")
			}
			return fmt.Sprintf("game.ControlledPermanentTypesOrControllerCounterPlacementReplacement(%q, %d, %d, %s, %s)",
				ability.Text,
				replacement.CounterMultiplier,
				replacement.CounterAddend,
				typesArg,
				controller,
			), nil
		}
		if replacement.MatchCounterKind {
			kind, err := renderCounterKind(replacement.CounterKindFilter)
			if err != nil {
				return "", err
			}
			ctx.need(importCounter)
			return fmt.Sprintf("game.ControlledPermanentTypesCounterKindPlacementReplacement(%q, %d, %d, %s, %s, %s)",
				ability.Text,
				replacement.CounterMultiplier,
				replacement.CounterAddend,
				kind,
				typesArg,
				controller,
			), nil
		}
		return fmt.Sprintf("game.ControlledPermanentTypesCounterPlacementReplacement(%q, %d, %d, %s, %s)",
			ability.Text,
			replacement.CounterMultiplier,
			replacement.CounterAddend,
			typesArg,
			controller,
		), nil
	}
	if replacement.CounterRecipientAnyPermanent {
		if replacement.MatchCounterKind {
			kind, err := renderCounterKind(replacement.CounterKindFilter)
			if err != nil {
				return "", err
			}
			ctx.need(importCounter)
			return fmt.Sprintf("game.ControlledPermanentCounterKindPlacementReplacement(%q, %d, %d, %s, %s)",
				ability.Text,
				replacement.CounterMultiplier,
				replacement.CounterAddend,
				kind,
				controller,
			), nil
		}
		return fmt.Sprintf("game.ControlledPermanentCounterPlacementReplacement(%q, %d, %d, %s)",
			ability.Text,
			replacement.CounterMultiplier,
			replacement.CounterAddend,
			controller,
		), nil
	}
	if !replacement.MatchCounterKind {
		return fmt.Sprintf("game.AnyCounterPlacementReplacement(%q, %d, %d, %s)",
			ability.Text,
			replacement.CounterMultiplier,
			replacement.CounterAddend,
			controller,
		), nil
	}
	kind, err := renderCounterKind(replacement.CounterKindFilter)
	if err != nil {
		return "", err
	}
	ctx.need(importCounter)
	return fmt.Sprintf("game.CounterPlacementReplacement(%q, %d, %d, %s, %s)",
		ability.Text,
		replacement.CounterMultiplier,
		replacement.CounterAddend,
		kind,
		controller,
	), nil
}

// renderNamedTokenSetReplacement renders Academy Manufactor's one-of-each
// token-type replacement, emitting each replaced token definition as a shared
// package-level var.
func (Renderer) renderNamedTokenSetReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventTokenCreated ||
		replacement.ControllerFilter == game.TriggerControllerAny ||
		len(replacement.CreateOneOfEachTokens) < 2 {
		return "", errors.New("render: unsupported one-of-each token-creation replacement shape")
	}
	controller, err := renderTriggerController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	vars := make([]string, 0, len(replacement.CreateOneOfEachTokens))
	for _, def := range replacement.CreateOneOfEachTokens {
		if def == nil {
			return "", errors.New("render: one-of-each token-creation replacement has a nil token definition")
		}
		vars = append(vars, ctx.tokenDefVar(def))
	}
	return fmt.Sprintf("game.NamedTokenSetReplacement(%q, []*game.CardDef{%s}, %s)",
		ability.Text,
		strings.Join(vars, ", "),
		controller,
	), nil
}

func renderTokenCreationReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventTokenCreated {
		return "", errors.New("render: unsupported token-creation replacement shape")
	}
	if replacement.ControllerFilter != game.TriggerControllerAny &&
		replacement.TokenAddend == 0 &&
		len(replacement.TokenRequiredSubtypes) == 0 &&
		len(replacement.TokenRequiredTypes) == 0 &&
		replacement.TokenMultiplier > 1 {
		controller, err := renderTriggerController(replacement.ControllerFilter)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.TokenCreationReplacement(%q, %d, %s)",
			ability.Text,
			replacement.TokenMultiplier,
			controller,
		), nil
	}
	return renderFilteredTokenCreationReplacement(ctx, ability)
}

// renderFilteredTokenCreationReplacement emits the spec-based builder for
// token-creation replacements that carry an any-player scope, a subtype filter,
// or an additive amount (Primal Vigor, Xorn).
func renderFilteredTokenCreationReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	controller, err := renderGroupEntersTappedController(replacement.ControllerFilter)
	if err != nil {
		return "", err
	}
	fields := []string{
		fmt.Sprintf("Multiplier: %d,", replacement.TokenMultiplier),
	}
	if replacement.TokenAddend != 0 {
		fields = append(fields, fmt.Sprintf("Addend: %d,", replacement.TokenAddend))
	}
	if len(replacement.TokenRequiredSubtypes) != 0 {
		subtypes, err := renderSubtypeSlice(ctx, replacement.TokenRequiredSubtypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Subtypes: %s,", subtypes))
	}
	if len(replacement.TokenRequiredTypes) != 0 {
		cardTypes, err := renderTypesCardSlice(ctx, replacement.TokenRequiredTypes)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Types: %s,", cardTypes))
	}
	if replacement.TokenAddendDef != nil {
		fields = append(fields, fmt.Sprintf("AddendDef: %s,", ctx.tokenDefVar(replacement.TokenAddendDef)))
	}
	if replacement.TokenReplaceDef != nil {
		fields = append(fields, fmt.Sprintf("ReplaceDef: %s,", ctx.tokenDefVar(replacement.TokenReplaceDef)))
	}
	fields = append(fields, fmt.Sprintf("Filter: %s,", controller))
	return fmt.Sprintf("game.TokenCreationReplacementFiltered(%q, &game.TokenCreationReplacementSpec{%s})",
		ability.Text,
		strings.Join(fields, " "),
	), nil
}

func (Renderer) renderGraveyardRedirectReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.ReplaceToZone != zone.Exile ||
		!replacement.MatchToZone ||
		replacement.ToZone != zone.Graveyard ||
		replacement.MatchEvent != game.EventZoneChanged ||
		replacement.Condition.Exists ||
		ability.UnlessPaid.Exists {
		return "", errors.New("render: unsupported graveyard-redirect replacement shape")
	}
	fromBattlefieldOnly := replacement.MatchFromZone && replacement.FromZone == zone.Battlefield
	if replacement.MatchFromZone && !fromBattlefieldOnly {
		return "", errors.New("render: unsupported graveyard-redirect source zone")
	}
	controller, err := renderGroupEntersTappedController(replacement.RedirectOwnerFilter)
	if err != nil {
		return "", err
	}
	controlFilter, err := renderGroupEntersTappedController(replacement.RedirectControlFilter)
	if err != nil {
		return "", err
	}
	// A named-counter redirect ("instead exile it with a void counter on it." —
	// Dauthi Voidwalker) emits the with-counter constructor variant; counterless
	// redirects (Leyline of the Void) keep the plain constructor byte-for-byte.
	constructor := "GraveyardRedirectReplacement"
	counterArg := ""
	if replacement.RedirectCounter.Exists {
		kind, err := renderCounterKind(replacement.RedirectCounter.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		constructor = "GraveyardRedirectExileWithCounterReplacement"
		counterArg = ", " + kind
	}
	if len(replacement.RedirectTypeFilter) == 0 {
		return fmt.Sprintf("game.%s(%q, %s, %s, %t%s)",
			constructor, ability.Text, controller, controlFilter, fromBattlefieldOnly, counterArg), nil
	}
	ctx.need(importTypes)
	typeLiterals := make([]string, 0, len(replacement.RedirectTypeFilter))
	for _, cardType := range replacement.RedirectTypeFilter {
		literal, err := cardTypeLiteral(cardType)
		if err != nil {
			return "", err
		}
		typeLiterals = append(typeLiterals, literal)
	}
	return fmt.Sprintf("game.%s(%q, %s, %s, %t%s, %s)",
		constructor, ability.Text, controller, controlFilter, fromBattlefieldOnly, counterArg, strings.Join(typeLiterals, ", ")), nil
}

func renderZoneDestinationReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	replacement := ability.Replacement
	if replacement.EntersTapped ||
		len(replacement.EntersWithCounters) != 0 ||
		ability.UnlessPaid.Exists ||
		replacement.Condition.Exists ||
		replacement.MatchEvent != game.EventZoneChanged ||
		!replacement.MatchToZone ||
		replacement.ToZone == zone.None {
		return "", errors.New("render: unsupported zone-destination replacement shape")
	}
	toZone, err := renderZone(replacement.ToZone)
	if err != nil {
		return "", err
	}
	replaceToZone, err := renderZone(replacement.ReplaceToZone)
	if err != nil {
		return "", err
	}
	fields := []string{
		"MatchEvent: game.EventZoneChanged,",
		"MatchToZone: true,",
		fmt.Sprintf("ToZone: %s,", toZone),
		fmt.Sprintf("ReplaceToZone: %s,", replaceToZone),
		"Duration: game.DurationPermanent,",
	}
	if replacement.ShuffleIntoLibrary {
		if replacement.ReplaceToZone != zone.Library {
			return "", errors.New("render: shuffle-into-library replacement must replace to library")
		}
		fields = append(fields, "ShuffleIntoLibrary: true,")
	}
	if replacement.RevealSource {
		fields = append(fields, "RevealSource: true,")
	}
	if replacement.MatchFromZone {
		fromZone, err := renderZone(replacement.FromZone)
		if err != nil {
			return "", err
		}
		fields = append(fields, "MatchFromZone: true,", fmt.Sprintf("FromZone: %s,", fromZone))
	}
	ctx.need(importZone)
	return fmt.Sprintf("game.ReplacementAbility{Text: %q, Replacement: %s}",
		ability.Text,
		structLit("game.ReplacementEffect", fields),
	), nil
}

func (r Renderer) renderCounterPlacements(ctx *renderCtx, placements []game.CounterPlacement) ([]string, error) {
	rendered := make([]string, 0, len(placements))
	for i := range placements {
		placement := placements[i]
		kind, err := renderCounterKind(placement.Kind)
		if err != nil {
			return nil, err
		}
		ctx.need(importCounter)
		if placement.Dynamic.Exists && placement.Dynamic.Val != nil {
			dynamic, err := r.renderDynamicAmount(ctx, placement.Dynamic.Val)
			if err != nil {
				return nil, err
			}
			ctx.need(importOpt)
			rendered = append(rendered, fmt.Sprintf("game.CounterPlacement{Kind: %s, Dynamic: opt.Val(&%s)}", kind, dynamic))
			continue
		}
		if placement.AmountFromX {
			rendered = append(rendered, fmt.Sprintf("game.CounterPlacement{Kind: %s, AmountFromX: true}", kind))
			continue
		}
		if placement.Amount <= 0 {
			return nil, fmt.Errorf("render: invalid ETB counter amount %d", placement.Amount)
		}
		rendered = append(rendered, fmt.Sprintf("game.CounterPlacement{Kind: %s, Amount: %d}", kind, placement.Amount))
	}
	return rendered, nil
}

func (r Renderer) renderResolutionPayment(ctx *renderCtx, payment game.ResolutionPayment) (string, error) {
	var fields []string
	hasCost := payment.ManaCost.Exists || payment.DynamicGenericManaCost.Exists || payment.ManaCostMultiplier.Exists || len(payment.AdditionalCosts) > 0
	if !hasCost {
		return "", errors.New("render: resolution payment has no cost")
	}
	if payment.Prompt != "" {
		fields = append(fields, fmt.Sprintf("Prompt: %q,", payment.Prompt))
	}
	if payment.Payer.Exists {
		payer, err := r.renderPlayerReference(payment.Payer.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Payer: opt.Val(%s),", payer))
	}
	if payment.ManaCost.Exists {
		manaCost, err := renderManaCostMultiline(ctx, payment.ManaCost.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ManaCost: opt.Val(%s),", manaCost))
	}
	if payment.DynamicGenericManaCost.Exists {
		if payment.DynamicGenericManaCost.Val == nil {
			return "", errors.New("render: resolution payment has nil dynamic generic mana cost")
		}
		dynamic, err := r.renderDynamicAmount(ctx, payment.DynamicGenericManaCost.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("DynamicGenericManaCost: opt.Val(&%s),", dynamic))
	}
	if payment.ManaCostMultiplier.Exists {
		if payment.ManaCostMultiplier.Val == nil {
			return "", errors.New("render: resolution payment has nil mana cost multiplier")
		}
		dynamic, err := r.renderDynamicAmount(ctx, payment.ManaCostMultiplier.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ManaCostMultiplier: opt.Val(&%s),", dynamic))
	}
	if len(payment.AdditionalCosts) > 0 {
		additionalCosts, err := r.renderAdditionalCosts(ctx, payment.AdditionalCosts)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AdditionalCosts: %s,", additionalCosts))
	}
	if payment.XValue != 0 {
		fields = append(fields, fmt.Sprintf("XValue: %d,", payment.XValue))
	}
	return structLit("game.ResolutionPayment", fields), nil
}

func (r Renderer) renderPay(ctx *renderCtx, pay game.Pay) (string, error) {
	payment, err := r.renderResolutionPayment(ctx, pay.Payment)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Payment: %s,", payment)}
	if pay.Prompt != "" {
		fields = append(fields, fmt.Sprintf("Prompt: %q,", pay.Prompt))
	}
	return structLit("game.Pay", fields), nil
}

func (r Renderer) renderPayRepeatedly(ctx *renderCtx, pay game.PayRepeatedly) (string, error) {
	payment, err := r.renderResolutionPayment(ctx, pay.Payment)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Payment: %s,", payment)}
	if pay.PublishCount != "" {
		fields = append(fields, fmt.Sprintf("PublishCount: %q,", string(pay.PublishCount)))
	}
	if pay.Prompt != "" {
		fields = append(fields, fmt.Sprintf("Prompt: %q,", pay.Prompt))
	}
	if pay.MaxCount.Exists {
		if pay.MaxCount.Val == nil {
			return "", errors.New("render: PayRepeatedly has nil max count")
		}
		dynamic, err := r.renderDynamicAmount(ctx, pay.MaxCount.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("MaxCount: opt.Val(&%s),", dynamic))
	}
	return structLit("game.PayRepeatedly", fields), nil
}

// renderConditionForETBReplacement renders a game.Condition for use in a
// conditional enters-tapped replacement. Only the exact supported shape is
// accepted; any other combination returns an error.
func (r Renderer) renderConditionForETBReplacement(ctx *renderCtx, cond *game.Condition) (string, error) {
	rendered, err := r.renderControllerControlsCondition(ctx, cond, "ETB replacement")
	if err != nil {
		return "", err
	}
	return "&" + rendered, nil
}

func (r Renderer) renderStaticAbilityCondition(ctx *renderCtx, cond *game.Condition) (string, error) {
	return r.renderControllerControlsCondition(ctx, cond, "static ability")
}

func (r Renderer) renderControllerControlsCondition(ctx *renderCtx, cond *game.Condition, context string) (string, error) {
	for i := range cond.Aggregates {
		if cond.Aggregates[i].Value < 0 {
			return "", fmt.Errorf("render: %s condition has a negative threshold", context)
		}
	}
	if cond.AnyPlayerLifeAtMost < 0 ||
		cond.ControllerTurnOfGameAtMost < 0 ||
		cond.SourceClassLevelAtLeast < 0 ||
		cond.SourceClassLevelLessThan < 0 ||
		cond.SourceLevelCountersAtLeast < 0 ||
		cond.SourceLevelCountersLessThan < 0 ||
		cond.ControllerGraveyardCardOfTypeCountAtLeast < 0 ||
		cond.ControllerGraveyardInstantOrSorceryCountAtLeast < 0 {
		return "", fmt.Errorf("render: %s condition has a negative threshold", context)
	}
	// Reject unsupported condition fields.
	if cond.SourceNotMonstrous ||
		cond.TargetEnteredThisTurn.Exists {
		return "", fmt.Errorf("render: unsupported condition shape for %s", context)
	}
	var fields []string
	if cond.Negate {
		fields = append(fields, "Negate: true,")
	}
	objectFields, hasPredicate, err := r.renderConditionObjectFields(ctx, cond, context)
	if err != nil {
		return "", err
	}
	fields = append(fields, objectFields...)
	if cond.ControlsMatching.Exists {
		rendered, err := r.renderSelectionCountForCondition(ctx, cond.ControlsMatching.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ControlsMatching: opt.Val(%s),", rendered))
		hasPredicate = true
	}
	if len(cond.Aggregates) > 0 {
		rendered, err := r.renderAggregateComparisons(ctx, cond.Aggregates)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Aggregates: %s,", rendered))
		hasPredicate = true
	}
	if cond.SourceClassLevelAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("SourceClassLevelAtLeast: %d,", cond.SourceClassLevelAtLeast))
		hasPredicate = true
	}
	if cond.SourceClassLevelLessThan > 0 {
		fields = append(fields, fmt.Sprintf("SourceClassLevelLessThan: %d,", cond.SourceClassLevelLessThan))
		hasPredicate = true
	}
	if cond.SourceLevelCountersAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("SourceLevelCountersAtLeast: %d,", cond.SourceLevelCountersAtLeast))
		hasPredicate = true
	}
	if cond.SourceLevelCountersLessThan > 0 {
		fields = append(fields, fmt.Sprintf("SourceLevelCountersLessThan: %d,", cond.SourceLevelCountersLessThan))
		hasPredicate = true
	}
	if cond.SourceCountersAtLeast > 0 {
		if !cond.SourceCounterKindKnown {
			return "", errors.New("render condition: source counter kind is unknown")
		}
		kind, err := renderCounterKind(cond.SourceCounterKind)
		if err != nil {
			return "", err
		}
		fields = append(fields,
			fmt.Sprintf("SourceCounterKind: %s,", kind),
			"SourceCounterKindKnown: true,",
			fmt.Sprintf("SourceCountersAtLeast: %d,", cond.SourceCountersAtLeast),
		)
		hasPredicate = true
	}
	if cond.SourceAttachedCombatCounterpartSubtypes != [2]types.Sub{} {
		ctx.need(importTypes)
		literals := make([]string, 2)
		for i, subtype := range cond.SourceAttachedCombatCounterpartSubtypes {
			literals[i] = SubtypeToLiteral(string(subtype), []string{"Creature"})
		}
		fields = append(fields, fmt.Sprintf(
			"SourceAttachedCombatCounterpartSubtypes: [2]types.Sub{%s},",
			strings.Join(literals, ", "),
		))
		hasPredicate = true
	}
	if cond.AnyOpponentPoisonAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("AnyOpponentPoisonAtLeast: %d,", cond.AnyOpponentPoisonAtLeast))
		hasPredicate = true
	}
	if cond.AnyPlayerLifeAtMost > 0 {
		fields = append(fields, fmt.Sprintf("AnyPlayerLifeAtMost: %d,", cond.AnyPlayerLifeAtMost))
		hasPredicate = true
	}
	if cond.ControllerHandEmpty {
		fields = append(fields, "ControllerHandEmpty: true,")
		hasPredicate = true
	}
	if cond.AllPlayersHandEmpty {
		fields = append(fields, "AllPlayersHandEmpty: true,")
		hasPredicate = true
	}
	if cond.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures {
		fields = append(fields, "EventPermanentNameUniqueAmongControlledAndGraveyardCreatures: true,")
		hasPredicate = true
	}
	if cond.SourceTributeNotPaid {
		fields = append(fields, "SourceTributeNotPaid: true,")
		hasPredicate = true
	}
	if cond.ControllerHasMaxSpeed {
		fields = append(fields, "ControllerHasMaxSpeed: true,")
		hasPredicate = true
	}
	if cond.SourceSaddled {
		fields = append(fields, "SourceSaddled: true,")
		hasPredicate = true
	}
	if cond.ControllerControlsCommander {
		fields = append(fields, "ControllerControlsCommander: true,")
		hasPredicate = true
	}
	if cond.LandEnteredThisTurnOrControlsBasicLand {
		fields = append(fields, "LandEnteredThisTurnOrControlsBasicLand: true,")
		hasPredicate = true
	}
	if cond.FirstCombatPhaseOfTurn {
		fields = append(fields, "FirstCombatPhaseOfTurn: true,")
		hasPredicate = true
	}
	if cond.ControllerControlsGreatestPowerCreature {
		fields = append(fields, "ControllerControlsGreatestPowerCreature: true,")
		hasPredicate = true
	}
	if cond.ControllerControlsGreatestToughnessCreature {
		fields = append(fields, "ControllerControlsGreatestToughnessCreature: true,")
		hasPredicate = true
	}
	if cond.EventPermanentPowerGreaterThanEachOtherCreature {
		fields = append(fields, "EventPermanentPowerGreaterThanEachOtherCreature: true,")
		hasPredicate = true
	}
	if cond.ControllerIsMonarch {
		fields = append(fields, "ControllerIsMonarch: true,")
		hasPredicate = true
	}
	if cond.ControllerWasMonarchAtTurnStart {
		fields = append(fields, "ControllerWasMonarchAtTurnStart: true,")
		hasPredicate = true
	}
	if cond.AnOpponentIsMonarch {
		fields = append(fields, "AnOpponentIsMonarch: true,")
		hasPredicate = true
	}
	if cond.NoMonarch {
		fields = append(fields, "NoMonarch: true,")
		hasPredicate = true
	}
	if cond.EventDefendingPlayerIsMonarch {
		fields = append(fields, "EventDefendingPlayerIsMonarch: true,")
		hasPredicate = true
	}
	if cond.ControllerHasInitiative {
		fields = append(fields, "ControllerHasInitiative: true,")
		hasPredicate = true
	}
	if cond.ControllerHasCityBlessing {
		fields = append(fields, "ControllerHasCityBlessing: true,")
		hasPredicate = true
	}
	if cond.ControllerCompletedADungeon {
		fields = append(fields, "ControllerCompletedADungeon: true,")
		hasPredicate = true
	}
	if cond.SourceControllerTurn {
		fields = append(fields, "SourceControllerTurn: true,")
		hasPredicate = true
	}
	if cond.ControllerTurnOfGameAtMost > 0 {
		fields = append(fields, fmt.Sprintf("ControllerTurnOfGameAtMost: %d,", cond.ControllerTurnOfGameAtMost))
		hasPredicate = true
	}
	if cond.SourceAbilityResolutionOrdinalThisTurn > 0 {
		fields = append(fields, fmt.Sprintf("SourceAbilityResolutionOrdinalThisTurn: %d,", cond.SourceAbilityResolutionOrdinalThisTurn))
		hasPredicate = true
	}
	if len(cond.ControllerControlsNamed) > 0 {
		quoted := make([]string, 0, len(cond.ControllerControlsNamed))
		for _, name := range cond.ControllerControlsNamed {
			quoted = append(quoted, fmt.Sprintf("%q", name))
		}
		fields = append(fields, fmt.Sprintf("ControllerControlsNamed: []string{%s},", strings.Join(quoted, ", ")))
		hasPredicate = true
	}
	if cond.ControllerCreatedTokenThisTurn {
		fields = append(fields, "ControllerCreatedTokenThisTurn: true,")
		hasPredicate = true
	}
	if cond.CastDuringControllerMainPhase {
		fields = append(fields, "CastDuringControllerMainPhase: true,")
		hasPredicate = true
	}
	if cond.SpellWasKicked {
		fields = append(fields, "SpellWasKicked: true,")
		hasPredicate = true
	}
	if cond.SpellWasBargained {
		fields = append(fields, "SpellWasBargained: true,")
		hasPredicate = true
	}
	if cond.SpellWasOffspring {
		fields = append(fields, "SpellWasOffspring: true,")
		hasPredicate = true
	}
	if cond.GiftPromised {
		fields = append(fields, "GiftPromised: true,")
		hasPredicate = true
	}
	if cond.EventPermanentWasKicked {
		fields = append(fields, "EventPermanentWasKicked: true,")
		hasPredicate = true
	}
	if cond.EventPermanentWasBargained {
		fields = append(fields, "EventPermanentWasBargained: true,")
		hasPredicate = true
	}
	if cond.EventPermanentWasOffspring {
		fields = append(fields, "EventPermanentWasOffspring: true,")
		hasPredicate = true
	}
	if cond.EventPermanentWasCastFromControllerHand {
		fields = append(fields, "EventPermanentWasCastFromControllerHand: true,")
		hasPredicate = true
	}
	if cond.SpellColorManaSpent.Count > 0 {
		colorLiteral, err := colorValueToLiteral(cond.SpellColorManaSpent.Color)
		if err != nil {
			return "", err
		}
		ctx.need(importColor)
		fields = append(fields, fmt.Sprintf("SpellColorManaSpent: game.ColorManaSpendThreshold{Color: %s, Count: %d},",
			colorLiteral, cond.SpellColorManaSpent.Count))
		hasPredicate = true
	}
	if cond.SpellSameColorManaSpentAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("SpellSameColorManaSpentAtLeast: %d,", cond.SpellSameColorManaSpentAtLeast))
		hasPredicate = true
	}
	if cond.CastFromZone.Exists {
		castZone, err := renderZone(cond.CastFromZone.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importZone)
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("CastFromZone: opt.Val(%s),", castZone))
		hasPredicate = true
	}
	if cond.ControllerGraveyardCardOfTypeCountAtLeast > 0 {
		literal, err := cardTypeLiteral(cond.ControllerGraveyardCountCardType)
		if err != nil {
			return "", err
		}
		ctx.need(importTypes)
		fields = append(fields, fmt.Sprintf("ControllerGraveyardCardOfTypeCountAtLeast: %d,", cond.ControllerGraveyardCardOfTypeCountAtLeast))
		fields = append(fields, fmt.Sprintf("ControllerGraveyardCountCardType: %s,", literal))
		hasPredicate = true
	}
	if cond.ControllerGraveyardInstantOrSorceryCountAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerGraveyardInstantOrSorceryCountAtLeast: %d,", cond.ControllerGraveyardInstantOrSorceryCountAtLeast))
		hasPredicate = true
	}
	if cond.AnyOpponentControls.Exists {
		rendered, err := r.renderSelectionCountForCondition(ctx, cond.AnyOpponentControls.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("AnyOpponentControls: opt.Val(%s),", rendered))
		hasPredicate = true
	}
	if cond.OpponentsControl.Exists {
		rendered, err := r.renderSelectionCountForCondition(ctx, cond.OpponentsControl.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("OpponentsControl: opt.Val(%s),", rendered))
		hasPredicate = true
	}
	if cond.ControlComparison.Exists {
		rendered, err := r.renderControlCountComparison(ctx, cond.ControlComparison.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ControlComparison: opt.Val(%s),", rendered))
		hasPredicate = true
	}
	if cond.EventHistory.Exists {
		rendered, err := r.renderEventHistoryCondition(ctx, &cond.EventHistory.Val, context)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("EventHistory: opt.Val(%s),", rendered))
		hasPredicate = true
	}
	if !hasPredicate {
		return "", fmt.Errorf("render: %s condition has no supported predicate", context)
	}
	return structLit("game.Condition", fields), nil
}

func (r Renderer) renderConditionObjectFields(
	ctx *renderCtx,
	cond *game.Condition,
	context string,
) (fields []string, hasPredicate bool, err error) {
	if cond.Object.Exists {
		object, err := r.renderObjectReference(cond.Object.Val)
		if err != nil {
			return nil, false, err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("Object: opt.Val(%s),", object))
	}
	if cond.ObjectMatches.Exists {
		if !cond.Object.Exists {
			return nil, false, fmt.Errorf("render: %s ObjectMatches condition has no Object reference", context)
		}
		if len(cond.Types) != 0 {
			return nil, false, fmt.Errorf("render: %s condition sets both legacy Types and ObjectMatches", context)
		}
		selection, err := r.renderSelection(ctx, cond.ObjectMatches.Val)
		if err != nil {
			return nil, false, err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ObjectMatches: opt.Val(%s),", selection))
	}
	if len(cond.Types) > 0 {
		cardTypes, err := renderTypesCardSlice(ctx, cond.Types)
		if err != nil {
			return nil, false, err
		}
		fields = append(fields, fmt.Sprintf("Types: %s,", cardTypes))
	}
	return fields, len(fields) > 0, nil
}

func (r Renderer) renderEventHistoryCondition(
	ctx *renderCtx,
	history *game.EventHistoryCondition,
	context string,
) (string, error) {
	pattern, err := r.renderTriggerPattern(ctx, &history.Pattern)
	if err != nil {
		return "", err
	}
	var window string
	switch history.Window {
	case game.EventHistoryCurrentTurn:
		window = "game.EventHistoryCurrentTurn"
	case game.EventHistoryPreviousTurn:
		window = "game.EventHistoryPreviousTurn"
	default:
		return "", fmt.Errorf("render: unsupported event-history window for %s condition", context)
	}
	if history.MinCount != 0 {
		return fmt.Sprintf("game.EventHistoryCondition{Pattern: %s, Window: %s, MinCount: %d}", pattern, window, history.MinCount), nil
	}
	return fmt.Sprintf("game.EventHistoryCondition{Pattern: %s, Window: %s}", pattern, window), nil
}

func (r Renderer) renderSelectionCountForCondition(ctx *renderCtx, count game.SelectionCount) (string, error) {
	if count.MinCount < 0 {
		return "", errors.New("render: condition permanent-count threshold cannot be negative")
	}
	selection, err := r.renderSelection(ctx, count.Selection)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Selection: %s,", selection)}
	if count.MinCount != 0 {
		fields = append(fields, fmt.Sprintf("MinCount: %d,", count.MinCount))
	}
	if count.TotalPower.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, count.TotalPower.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("TotalPower: opt.Val(%s),", cmp))
	}
	if count.DistinctNames.Exists {
		ctx.need(importOpt)
		cmp, err := renderCompareInt(ctx, count.DistinctNames.Val)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("DistinctNames: opt.Val(%s),", cmp))
	}
	return structLit("game.SelectionCount", fields), nil
}

func (r Renderer) renderControlCountComparison(ctx *renderCtx, cmp game.ControlCountComparison) (string, error) {
	selection, err := r.renderSelection(ctx, cmp.Selection)
	if err != nil {
		return "", err
	}
	left, err := renderControlPlayerScope(cmp.Left)
	if err != nil {
		return "", err
	}
	right, err := renderControlPlayerScope(cmp.Right)
	if err != nil {
		return "", err
	}
	op, err := renderCompareOp(cmp.Op)
	if err != nil {
		return "", err
	}
	ctx.need(importCompare)
	fields := []string{
		fmt.Sprintf("Selection: %s,", selection),
		fmt.Sprintf("Left: %s,", left),
		fmt.Sprintf("Right: %s,", right),
		fmt.Sprintf("Op: %s,", op),
	}
	return structLit("game.ControlCountComparison", fields), nil
}

func renderControlPlayerScope(scope game.ControlPlayerScope) (string, error) {
	switch scope {
	case game.ControlPlayerController:
		return "game.ControlPlayerController", nil
	case game.ControlPlayerAnyOpponent:
		return "game.ControlPlayerAnyOpponent", nil
	case game.ControlPlayerEachOpponent:
		return "game.ControlPlayerEachOpponent", nil
	case game.ControlPlayerTriggeringPlayer:
		return "game.ControlPlayerTriggeringPlayer", nil
	default:
		return "", fmt.Errorf("render: unsupported control player scope %d", scope)
	}
}

// renderEntersAsCopyReplacement renders the self enters-as-copy replacement
// (Clone family) into a game.EntersAsCopyReplacement constructor call carrying
// the copied-permanent selection, the optional "you may" flag, and the
// recognized copiable riders.
func (r Renderer) renderEntersAsCopyReplacement(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	if ability.Replacement.EntersAsCopySelection == nil {
		return "", errors.New("render: enters-as-copy replacement requires a selection")
	}
	selection, err := r.renderSelection(ctx, *ability.Replacement.EntersAsCopySelection)
	if err != nil {
		return "", err
	}
	args := []string{
		fmt.Sprintf("%q", ability.Text),
		"&" + selection,
		fmt.Sprintf("%t", ability.Replacement.EntersAsCopyOptional),
		fmt.Sprintf("%t", ability.Replacement.EntersAsCopyNotLegendary),
	}
	conditionalCounters := "nil"
	if len(ability.Replacement.EntersAsCopyConditionalCounters) > 0 {
		placements, err := r.renderEntersAsCopyConditionalCounters(ctx, ability.Replacement.EntersAsCopyConditionalCounters)
		if err != nil {
			return "", err
		}
		ctx.need(importGame)
		conditionalCounters = fmt.Sprintf("[]game.ConditionalCounterPlacement{%s}", strings.Join(placements, ", "))
	}
	args = append(args, conditionalCounters, fmt.Sprintf("%t", ability.Replacement.EntersAsCopyUntilEndOfTurn))
	addKeywords := "nil"
	if len(ability.Replacement.EntersAsCopyAddKeywords) > 0 {
		rendered := make([]string, 0, len(ability.Replacement.EntersAsCopyAddKeywords))
		for _, keyword := range ability.Replacement.EntersAsCopyAddKeywords {
			literal, err := renderKeyword(keyword)
			if err != nil {
				return "", err
			}
			rendered = append(rendered, literal)
		}
		ctx.need(importGame)
		addKeywords = fmt.Sprintf("[]game.Keyword{%s}", strings.Join(rendered, ", "))
	}
	args = append(args, addKeywords)
	addSubtypes, err := renderSubtypeSlice(ctx, ability.Replacement.EntersAsCopyAddSubtypes)
	if err != nil {
		return "", err
	}
	args = append(args, addSubtypes)
	if len(ability.Replacement.EntersAsCopyAddTypes) > 0 {
		ctx.need(importTypes)
		for _, cardType := range ability.Replacement.EntersAsCopyAddTypes {
			literal, err := cardTypeLiteral(cardType)
			if err != nil {
				return "", err
			}
			args = append(args, literal)
		}
	}
	rendered := fmt.Sprintf("game.EntersAsCopyReplacement(%s)", strings.Join(args, ", "))
	if ability.Replacement.EntersAsCopyBasePower.Exists || ability.Replacement.EntersAsCopyBaseToughness.Exists {
		rendered = fmt.Sprintf("game.EntersAsCopyWithBasePowerToughness(%s, %d, %d)", rendered, ability.Replacement.EntersAsCopyBasePower.Val, ability.Replacement.EntersAsCopyBaseToughness.Val)
	}
	if ability.Replacement.EntersAsCopyMaxManaValueFromManaSpent {
		rendered = fmt.Sprintf("game.EntersAsCopyWithManaSpentBound(%s)", rendered)
	}
	if ability.Replacement.EntersAsCopyTapped {
		rendered = fmt.Sprintf("game.EntersTappedAsCopy(%s)", rendered)
	}
	if ability.Replacement.EntersAsCopyRetainName {
		rendered = fmt.Sprintf("game.EntersAsCopyWithRetainedName(%s)", rendered)
	}
	if ability.Replacement.EntersAsCopyAddOtherAbilities {
		rendered = fmt.Sprintf("game.EntersAsCopyWithOtherAbilities(%s)", rendered)
	}
	if len(ability.Replacement.EntersAsCopyAddAbilities) > 0 {
		added := make([]string, 0, len(ability.Replacement.EntersAsCopyAddAbilities))
		for _, addAbility := range ability.Replacement.EntersAsCopyAddAbilities {
			abilitySource, err := r.renderEmblemAbility(ctx, addAbility)
			if err != nil {
				return "", err
			}
			added = append(added, "new("+abilitySource+")")
		}
		rendered = fmt.Sprintf("game.EntersAsCopyWithAddedAbilities(%s, %s)", rendered, strings.Join(added, ", "))
	}
	return rendered, nil
}

// renderEntersAsCopyConditionalCounters renders the conditional copiable counter
// riders (Spark Double) into game.ConditionalCounterPlacement literals.
func (Renderer) renderEntersAsCopyConditionalCounters(ctx *renderCtx, placements []game.ConditionalCounterPlacement) ([]string, error) {
	rendered := make([]string, 0, len(placements))
	for _, placement := range placements {
		kind, err := renderCounterKind(placement.Kind)
		if err != nil {
			return nil, err
		}
		cardType, err := cardTypeLiteral(placement.IfType)
		if err != nil {
			return nil, err
		}
		ctx.need(importCounter)
		ctx.need(importTypes)
		rendered = append(rendered, fmt.Sprintf("game.ConditionalCounterPlacement{Kind: %s, Amount: %d, IfType: %s}", kind, placement.Amount, cardType))
	}
	return rendered, nil
}
