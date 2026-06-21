package cardgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/zone"
)

func (r Renderer) renderReplacementAbility(ctx *renderCtx, ability *game.ReplacementAbility) (string, error) {
	if ability.Replacement.EntersAsCopy {
		return r.renderEntersAsCopyReplacement(ctx, ability)
	}
	if ability.Replacement.DrawFromEmptyLibraryWins {
		return fmt.Sprintf("game.DrawFromEmptyLibraryWinReplacement(%q)", ability.Text), nil
	}
	if ability.Replacement.DrawCardMultiplier > 1 {
		return fmt.Sprintf("game.DrawCardMultiplierReplacement(%q, %d, %t)",
			ability.Text, ability.Replacement.DrawCardMultiplier, ability.Replacement.DrawCardExceptFirstInDrawStep), nil
	}
	if ability.Replacement.LifeGainMultiplier > 1 || ability.Replacement.LifeGainAddend != 0 {
		return fmt.Sprintf("game.LifeGainReplacement(%q, %d, %d)",
			ability.Text, ability.Replacement.LifeGainMultiplier, ability.Replacement.LifeGainAddend), nil
	}
	if ability.Replacement.EntersTappedOthers {
		return r.renderGroupEntersTappedReplacement(ctx, ability)
	}
	if len(ability.Replacement.EntersWithCounters) > 0 {
		if ability.UnlessPaid.Exists {
			return "", errors.New("render: ETB counter replacement cannot also require payment")
		}
		if ability.Replacement.EntersTapped && ability.Replacement.Condition.Exists {
			return "", errors.New("render: ETB counter replacement cannot both tap and have a condition")
		}
		placements, err := renderCounterPlacements(ctx, ability.Replacement.EntersWithCounters)
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
	if ability.Replacement.EntryTypeChoice {
		return fmt.Sprintf("game.EntryTypeChoiceReplacement(%q)", ability.Text), nil
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
	return "", fmt.Errorf("render: unsupported replacement ability %q", ability.Text)
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
	if len(replacement.EntersTappedTypes) == 0 {
		return fmt.Sprintf("game.EntersTappedGroupReplacement(%q, %s)", ability.Text, controller), nil
	}
	ctx.need(importTypes)
	typeLiterals := make([]string, 0, len(replacement.EntersTappedTypes))
	for _, cardType := range replacement.EntersTappedTypes {
		literal, err := cardTypeLiteral(cardType)
		if err != nil {
			return "", err
		}
		typeLiterals = append(typeLiterals, literal)
	}
	return fmt.Sprintf("game.EntersTappedGroupReplacement(%q, %s, %s)",
		ability.Text, controller, strings.Join(typeLiterals, ", ")), nil
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
		replacement.ControllerFilter == game.TriggerControllerAny ||
		(replacement.DamageMultiplier <= 1 && replacement.DamageAddend == 0) {
		return "", errors.New("render: unsupported damage replacement shape")
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
	if len(replacement.RedirectTypeFilter) == 0 {
		return fmt.Sprintf("game.GraveyardRedirectReplacement(%q, %s, %t)",
			ability.Text, controller, fromBattlefieldOnly), nil
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
	return fmt.Sprintf("game.GraveyardRedirectReplacement(%q, %s, %t, %s)",
		ability.Text, controller, fromBattlefieldOnly, strings.Join(typeLiterals, ", ")), nil
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

func renderCounterPlacements(ctx *renderCtx, placements []game.CounterPlacement) ([]string, error) {
	rendered := make([]string, 0, len(placements))
	for _, placement := range placements {
		kind, err := renderCounterKind(placement.Kind)
		if err != nil {
			return nil, err
		}
		ctx.need(importCounter)
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
	if cond.ControllerLifeAtLeast < 0 ||
		cond.ControllerHandSizeAtLeast < 0 ||
		cond.AnyPlayerLifeAtMost < 0 ||
		cond.OpponentCountAtLeast < 0 ||
		cond.ControllerGraveyardCardCountAtLeast < 0 ||
		cond.ControllerGraveyardCardTypeCountAtLeast < 0 ||
		cond.ControllerBasicLandTypeCountAtLeast < 0 ||
		cond.ControllerCreaturePowerDiversityAtLeast < 0 {
		return "", fmt.Errorf("render: %s condition has a negative threshold", context)
	}
	// Reject unsupported condition fields.
	if cond.SourceClassLevelAtLeast != 0 ||
		cond.SourceClassLevelLessThan != 0 ||
		cond.SourceNotMonstrous ||
		cond.ControllerHasMaxSpeed ||
		cond.TargetEnteredThisTurn.Exists ||
		cond.CastFromZone.Exists {
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
	if !cond.ControllerControls.Empty() {
		filter := cond.ControllerControls
		if filter.Power.Exists ||
			filter.Toughness.Exists ||
			filter.TotalPower.Exists {
			return "", fmt.Errorf("render: unsupported PermanentFilter shape for %s condition", context)
		}
		filterStr, err := r.renderPermanentFilterForCondition(ctx, filter)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ControllerControls: %s,", filterStr))
		hasPredicate = true
	}
	if cond.ControlsMatching.Exists {
		rendered, err := r.renderSelectionCountForCondition(ctx, cond.ControlsMatching.Val)
		if err != nil {
			return "", err
		}
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ControlsMatching: opt.Val(%s),", rendered))
		hasPredicate = true
	}
	if cond.ControllerLifeAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerLifeAtLeast: %d,", cond.ControllerLifeAtLeast))
		hasPredicate = true
	}
	if cond.ControllerHandSizeAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerHandSizeAtLeast: %d,", cond.ControllerHandSizeAtLeast))
		hasPredicate = true
	}
	if cond.ControllerHandSizeExactly.Exists {
		ctx.need(importOpt)
		fields = append(fields, fmt.Sprintf("ControllerHandSizeExactly: opt.Val(%d),", cond.ControllerHandSizeExactly.Val))
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
	if cond.OpponentCountAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("OpponentCountAtLeast: %d,", cond.OpponentCountAtLeast))
		hasPredicate = true
	}
	if cond.ControllerHandEmpty {
		fields = append(fields, "ControllerHandEmpty: true,")
		hasPredicate = true
	}
	if cond.EventPermanentNameUniqueAmongControlledAndGraveyardCreatures {
		fields = append(fields, "EventPermanentNameUniqueAmongControlledAndGraveyardCreatures: true,")
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
	if cond.ControllerGraveyardCardCountAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerGraveyardCardCountAtLeast: %d,", cond.ControllerGraveyardCardCountAtLeast))
		hasPredicate = true
	}
	if cond.ControllerGraveyardCardTypeCountAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerGraveyardCardTypeCountAtLeast: %d,", cond.ControllerGraveyardCardTypeCountAtLeast))
		hasPredicate = true
	}
	if cond.ControllerBasicLandTypeCountAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerBasicLandTypeCountAtLeast: %d,", cond.ControllerBasicLandTypeCountAtLeast))
		hasPredicate = true
	}
	if cond.ControllerCreaturePowerDiversityAtLeast > 0 {
		fields = append(fields, fmt.Sprintf("ControllerCreaturePowerDiversityAtLeast: %d,", cond.ControllerCreaturePowerDiversityAtLeast))
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
	default:
		return "", fmt.Errorf("render: unsupported control player scope %d", scope)
	}
}

func (Renderer) renderPermanentFilterForCondition(ctx *renderCtx, filter game.PermanentFilter) (string, error) {
	if filter.MinCount < 0 {
		return "", errors.New("render: condition permanent-count threshold cannot be negative")
	}
	var fields []string
	if len(filter.Types) > 0 {
		lits, err := renderTypesCardSlice(ctx, filter.Types)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Types: %s,", lits))
	}
	if len(filter.Supertypes) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(filter.Supertypes))
		for _, st := range filter.Supertypes {
			lit, err := supertypeLiteral(st)
			if err != nil {
				return "", err
			}
			literals = append(literals, lit)
		}
		fields = append(fields, fmt.Sprintf("Supertypes: []types.Super{%s},", strings.Join(literals, ", ")))
	}
	if len(filter.SubtypesAny) > 0 {
		ctx.need(importTypes)
		literals := make([]string, 0, len(filter.SubtypesAny))
		cardTypes := make([]string, 0, len(filter.Types))
		for _, cardType := range filter.Types {
			cardTypes = append(cardTypes, string(cardType))
		}
		for _, subtype := range filter.SubtypesAny {
			literals = append(literals, SubtypeToLiteral(string(subtype), cardTypes))
		}
		fields = append(fields, fmt.Sprintf("SubtypesAny: []types.Sub{%s},", strings.Join(literals, ", ")))
	}
	if len(filter.ColorsAny) > 0 {
		literals, err := renderColorSlice(ctx, filter.ColorsAny)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ColorsAny: %s,", literals))
	}
	if len(filter.ExcludedColors) > 0 {
		literals, err := renderColorSlice(ctx, filter.ExcludedColors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("ExcludedColors: %s,", literals))
	}
	if filter.MinCount != 0 {
		fields = append(fields, fmt.Sprintf("MinCount: %d,", filter.MinCount))
	}
	if filter.ExcludeSource {
		fields = append(fields, "ExcludeSource: true,")
	}
	return structLit("game.PermanentFilter", fields), nil
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
