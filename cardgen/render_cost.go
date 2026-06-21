package cardgen

import (
	"errors"
	"fmt"
	"strings"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// renderEternalizeFamilyAbility renders an Eternalize or Embalm activated ability
// as the canonical builder call with its mana cost and the card's printed
// creature subtypes.
func (r Renderer) renderEternalizeFamilyAbility(ctx *renderCtx, builder string, manaCost cost.Mana, subtypes []types.Sub) (string, error) {
	renderedCost, err := r.renderManaCost(ctx, manaCost)
	if err != nil {
		return "", err
	}
	args := make([]string, 0, len(subtypes)+1)
	args = append(args, renderedCost)
	for _, subtype := range subtypes {
		ctx.need(importTypes)
		args = append(args, SubtypeToLiteral(string(subtype), nil))
	}
	return fmt.Sprintf("%s(%s)", builder, strings.Join(args, ", ")), nil
}

func (r Renderer) renderDamageRecipient(ctx *renderCtx, recipient game.DamageRecipient) (string, error) {
	if object, ok := recipient.AnyTargetObjectReference(); ok {
		return fmt.Sprintf("game.AnyTargetDamageRecipient(%d)", object.TargetIndex()), nil
	}
	if object, ok := recipient.ObjectReference(); ok {
		rendered, err := r.renderObjectReference(object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.ObjectDamageRecipient(%s)", rendered), nil
	}
	if player, ok := recipient.PlayerReference(); ok {
		rendered, err := r.renderPlayerReference(player)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.PlayerDamageRecipient(%s)", rendered), nil
	}
	if group, ok := recipient.GroupReference(); ok {
		rendered, err := r.renderGroupReference(ctx, group)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.GroupDamageRecipient(%s)", rendered), nil
	}
	if group, ok := recipient.PlayerGroupReference(); ok {
		switch group.Kind {
		case game.PlayerGroupReferenceOpponents:
			return "game.PlayerGroupDamageRecipient(game.OpponentsReference())", nil
		case game.PlayerGroupReferenceAllPlayers:
			return "game.PlayerGroupDamageRecipient(game.AllPlayersReference())", nil
		}
	}
	return "", errors.New("render: unsupported damage recipient")
}

func (Renderer) renderObjectReference(reference game.ObjectReference) (string, error) {
	switch reference.Kind() {
	case game.ObjectReferenceTargetPermanent:
		return fmt.Sprintf("game.TargetPermanentReference(%d)", reference.TargetIndex()), nil
	case game.ObjectReferenceTargetStackObject:
		return fmt.Sprintf("game.TargetStackObjectReference(%d)", reference.TargetIndex()), nil
	case game.ObjectReferenceTargetObject:
		return fmt.Sprintf("game.TargetObjectReference(%d)", reference.TargetIndex()), nil
	case game.ObjectReferenceCapturedTargetStackObject:
		return fmt.Sprintf("game.CapturedTargetStackObjectReference(%d)", reference.TargetIndex()), nil
	case game.ObjectReferenceSourcePermanent:
		return "game.SourcePermanentReference()", nil
	case game.ObjectReferenceSourceAttachedPermanent:
		return "game.SourceAttachedPermanentReference()", nil
	case game.ObjectReferenceTargetAttachedPermanent:
		return fmt.Sprintf("game.TargetAttachedPermanentReference(%d)", reference.TargetIndex()), nil
	case game.ObjectReferenceLinkedObject:
		return fmt.Sprintf("game.LinkedObjectReference(%q)", reference.LinkID()), nil
	case game.ObjectReferenceEventPermanent:
		return "game.EventPermanentReference()", nil
	case game.ObjectReferenceEventRelatedPermanent:
		return "game.EventRelatedPermanentReference()", nil
	case game.ObjectReferenceSourceCard:
		return "game.SourceCardPermanentReference()", nil
	case game.ObjectReferenceSacrificedCost:
		return "game.SacrificedCostReference()", nil
	default:
		return "", fmt.Errorf("render: unsupported object reference kind %d", reference.Kind())
	}
}

func (r Renderer) renderPlayerReference(reference game.PlayerReference) (string, error) {
	switch reference.Kind() {
	case game.PlayerReferenceController:
		return "game.ControllerReference()", nil
	case game.PlayerReferenceTargetPlayer:
		return fmt.Sprintf("game.TargetPlayerReference(%d)", reference.TargetIndex()), nil
	case game.PlayerReferenceObjectController:
		object, _ := reference.Object()
		rendered, err := r.renderObjectReference(object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.ObjectControllerReference(%s)", rendered), nil
	case game.PlayerReferenceObjectOwner:
		object, _ := reference.Object()
		rendered, err := r.renderObjectReference(object)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.ObjectOwnerReference(%s)", rendered), nil
	case game.PlayerReferenceEventPlayer:
		return "game.EventPlayerReference()", nil
	case game.PlayerReferenceCapturedTargetController:
		return fmt.Sprintf("game.CapturedTargetControllerReference(%d)", reference.TargetIndex()), nil
	case game.PlayerReferenceDefendingPlayer:
		return "game.DefendingPlayerReference()", nil
	default:
		return "", fmt.Errorf("render: unsupported player reference kind %d", reference.Kind())
	}
}

func (r Renderer) renderKeywordAbility(ctx *renderCtx, keyword game.KeywordAbility) (string, error) {
	if simple, ok := keyword.(game.SimpleKeyword); ok {
		kw, err := renderKeyword(simple.Kind)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.SimpleKeyword{Kind: %s}", kw), nil
	}
	if ward, ok := keyword.(game.WardKeyword); ok {
		wardCost, err := r.renderManaCost(ctx, ward.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.WardKeyword{Cost: %s}", wardCost), nil
	}
	if cumulative, ok := keyword.(game.CumulativeUpkeepKeyword); ok {
		cumulativeCost, err := r.renderManaCost(ctx, cumulative.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.CumulativeUpkeepKeyword{Cost: %s}", cumulativeCost), nil
	}
	if cycling, ok := keyword.(game.CyclingKeyword); ok {
		cyclingCost, err := r.renderManaCost(ctx, cycling.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.CyclingKeyword{Cost: %s}", cyclingCost), nil
	}
	if ninjutsu, ok := keyword.(game.NinjutsuKeyword); ok {
		ninjutsuCost, err := r.renderManaCost(ctx, ninjutsu.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.NinjutsuKeyword{Cost: %s}", ninjutsuCost), nil
	}
	if mutate, ok := keyword.(game.MutateKeyword); ok {
		mutateCost, err := r.renderManaCost(ctx, mutate.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.MutateKeyword{Cost: %s}", mutateCost), nil
	}
	if kicker, ok := keyword.(game.KickerKeyword); ok {
		kickerCost, err := r.renderManaCost(ctx, kicker.Cost)
		if err != nil {
			return "", err
		}
		if len(kicker.BonusContent.Modes) != 0 {
			return "", errors.New("render: Kicker bonus content must be rendered by its owning ability")
		}
		return fmt.Sprintf("game.KickerKeyword{Cost: %s}", kickerCost), nil
	}
	if madness, ok := keyword.(game.MadnessKeyword); ok {
		madnessCost, err := r.renderManaCost(ctx, madness.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.MadnessKeyword{Cost: %s}", madnessCost), nil
	}
	if flashback, ok := keyword.(game.FlashbackKeyword); ok {
		flashbackCost, err := r.renderManaCost(ctx, flashback.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.FlashbackKeyword{Cost: %s}", flashbackCost), nil
	}
	if morph, ok := keyword.(game.MorphKeyword); ok {
		morphCost, err := r.renderManaCost(ctx, morph.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.MorphKeyword{Cost: %s}", morphCost), nil
	}
	if disguise, ok := keyword.(game.DisguiseKeyword); ok {
		disguiseCost, err := r.renderManaCost(ctx, disguise.Cost)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("game.DisguiseKeyword{Cost: %s}", disguiseCost), nil
	}
	if toxic, ok := keyword.(game.ToxicKeyword); ok {
		return fmt.Sprintf("game.ToxicKeyword{Amount: %d}", toxic.Amount), nil
	}
	return "", fmt.Errorf("render: unsupported keyword ability %T", keyword)
}

func (Renderer) renderManaCost(ctx *renderCtx, manaCost cost.Mana) (string, error) {
	ctx.need(importCost)
	if len(manaCost) == 0 {
		return "cost.Mana{}", nil
	}
	symbols := make([]string, 0, len(manaCost))
	for _, symbol := range manaCost {
		sym, err := renderManaSymbol(ctx, symbol)
		if err != nil {
			return "", err
		}
		symbols = append(symbols, sym)
	}
	return "cost.Mana{" + strings.Join(symbols, ", ") + "}", nil
}

// renderManaCostMultiline renders a printed face ManaCost as a multi-line
// cost.Mana literal so gofmt preserves the canonical generated-card layout.
func renderManaCostMultiline(ctx *renderCtx, manaCost cost.Mana) (string, error) {
	ctx.need(importCost)
	if len(manaCost) == 0 {
		return "cost.Mana{}", nil
	}
	symbols := make([]string, 0, len(manaCost))
	for _, symbol := range manaCost {
		sym, err := renderManaSymbol(ctx, symbol)
		if err != nil {
			return "", err
		}
		symbols = append(symbols, sym)
	}
	return "cost.Mana{\n\t\t\t" + strings.Join(symbols, ",\n\t\t\t") + ",\n\t\t}", nil
}

func (Renderer) renderAdditionalCosts(ctx *renderCtx, costs []cost.Additional) (string, error) {
	ctx.need(importCost)
	if len(costs) == 1 &&
		costs[0].Kind == cost.AdditionalTap &&
		costs[0].Text == "" &&
		costs[0].Amount == 0 &&
		costs[0].Source == zone.None {
		return "cost.Tap", nil
	}
	elements := make([]string, 0, len(costs))
	for _, additional := range costs {
		rendered, err := renderAdditional(ctx, additional)
		if err != nil {
			return "", err
		}
		elements = append(elements, rendered+",")
	}
	return sliceLit("cost.Additional", elements), nil
}

func (r Renderer) renderAlternativeCosts(ctx *renderCtx, alternatives []cost.Alternative) (string, error) {
	ctx.need(importCost)
	elements := make([]string, 0, len(alternatives))
	for _, alternative := range alternatives {
		fields := []string{}
		if alternative.Label != "" {
			fields = append(fields, fmt.Sprintf("Label: %q,", alternative.Label))
		}
		if alternative.ManaCost.Exists {
			rendered, err := r.renderManaCost(ctx, alternative.ManaCost.Val)
			if err != nil {
				return "", err
			}
			ctx.need(importOpt)
			fields = append(fields, fmt.Sprintf("ManaCost: opt.Val(%s),", rendered))
		}
		if len(alternative.AdditionalCosts) > 0 {
			rendered, err := r.renderAdditionalCosts(ctx, alternative.AdditionalCosts)
			if err != nil {
				return "", err
			}
			fields = append(fields, fmt.Sprintf("AdditionalCosts: %s,", rendered))
		}
		switch alternative.Condition {
		case cost.AlternativeConditionNone:
		case cost.AlternativeConditionControlsCommander:
			fields = append(fields, "Condition: cost.AlternativeConditionControlsCommander,")
		case cost.AlternativeConditionNotYourTurn:
			fields = append(fields, "Condition: cost.AlternativeConditionNotYourTurn,")
		default:
			return "", fmt.Errorf("render: unsupported alternative-cost condition %d", alternative.Condition)
		}
		elements = append(elements, structLit("cost.Alternative", fields)+",")
	}
	return sliceLit("cost.Alternative", elements), nil
}

func renderAdditional(ctx *renderCtx, additional cost.Additional) (string, error) {
	ctx.need(importCost)
	kind, err := renderAdditionalKind(additional.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if additional.Text != "" {
		fields = append(fields, fmt.Sprintf("Text: %q,", additional.Text))
	}
	if additional.Amount != 0 {
		fields = append(fields, fmt.Sprintf("Amount: %d,", additional.Amount))
	}
	if additional.AmountFromX {
		fields = append(fields, "AmountFromX: true,")
	}
	if additional.AmountDynamic != cost.AdditionalDynamicAmountNone {
		dynamic, err := renderAdditionalDynamicAmount(additional.AmountDynamic)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("AmountDynamic: %s,", dynamic))
	}
	if additional.Source != zone.None {
		ctx.need(importZone)
		zoneLiteral, err := renderZone(additional.Source)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Source: %s,", zoneLiteral))
	}
	if additional.MatchPermanentType {
		cardType, err := cardTypeLiteral(additional.PermanentType)
		if err != nil {
			return "", err
		}
		ctx.need(importTypes)
		fields = append(fields,
			"MatchPermanentType: true,",
			fmt.Sprintf("PermanentType: %s,", cardType),
		)
		if additional.PermanentTypeAlt != "" {
			altType, err := cardTypeLiteral(additional.PermanentTypeAlt)
			if err != nil {
				return "", err
			}
			fields = append(fields, fmt.Sprintf("PermanentTypeAlt: %s,", altType))
		}
	}
	if additional.MatchCardType {
		cardType, err := cardTypeLiteral(additional.CardType)
		if err != nil {
			return "", err
		}
		ctx.need(importTypes)
		fields = append(fields,
			"MatchCardType: true,",
			fmt.Sprintf("CardType: %s,", cardType),
		)
	}
	if additional.MatchCardColor {
		colorLiteral, err := colorValueToLiteral(additional.CardColor)
		if err != nil {
			return "", err
		}
		ctx.need(importColor)
		fields = append(fields,
			"MatchCardColor: true,",
			fmt.Sprintf("CardColor: %s,", colorLiteral),
		)
	}
	if additional.RequireTapped {
		fields = append(fields, "RequireTapped: true,")
	}
	if additional.ExcludeSource {
		fields = append(fields, "ExcludeSource: true,")
	}
	if additional.RequireSupertype != "" {
		supertype, err := supertypeLiteral(additional.RequireSupertype)
		if err != nil {
			return "", err
		}
		ctx.need(importTypes)
		fields = append(fields, fmt.Sprintf("RequireSupertype: %s,", supertype))
	}
	if additional.SubtypesAny != (cost.SubtypeSet{}) {
		ctx.need(importTypes)
		literals := make([]string, 0, len(additional.SubtypesAny))
		for _, subtype := range additional.SubtypesAny {
			if subtype == "" {
				continue
			}
			literals = append(literals, SubtypeToLiteral(string(subtype), []string{"Land", "Creature"}))
		}
		fields = append(fields, fmt.Sprintf("SubtypesAny: cost.SubtypeSet{%s},", strings.Join(literals, ", ")))
	}
	if additional.Kind == cost.AdditionalRemoveCounter || additional.Kind == cost.AdditionalPutCounter {
		counterKind, err := renderCounterKind(additional.CounterKind)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		fields = append(fields, fmt.Sprintf("CounterKind: %s,", counterKind))
	}
	if additional.ChoiceGroup != 0 {
		fields = append(fields, fmt.Sprintf("ChoiceGroup: %d,", additional.ChoiceGroup))
	}
	return structLit("", fields), nil
}

func (r Renderer) renderQuantity(ctx *renderCtx, quantity game.Quantity) (string, error) {
	dynamic := quantity.DynamicAmount()
	if !dynamic.Exists {
		return fmt.Sprintf("game.Fixed(%d)", quantity.Value()), nil
	}
	rendered, err := r.renderDynamicAmount(ctx, &dynamic.Val)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("game.Dynamic(%s)", rendered), nil
}

func (r Renderer) renderDynamicAmount(ctx *renderCtx, dynamic *game.DynamicAmount) (string, error) {
	kind, err := renderDynamicAmountKind(dynamic.Kind)
	if err != nil {
		return "", err
	}
	fields := []string{fmt.Sprintf("Kind: %s,", kind)}
	if dynamic.Constant != 0 {
		fields = append(fields, fmt.Sprintf("Constant: %d,", dynamic.Constant))
	}
	if dynamic.Multiplier != 0 {
		fields = append(fields, fmt.Sprintf("Multiplier: %d,", dynamic.Multiplier))
	}
	if dynamic.Kind == game.DynamicAmountTargetCounters ||
		dynamic.Kind == game.DynamicAmountObjectCounters ||
		dynamic.CounterKind != 0 {
		counterKind, err := renderCounterKind(dynamic.CounterKind)
		if err != nil {
			return "", err
		}
		ctx.need(importCounter)
		fields = append(fields, fmt.Sprintf("CounterKind: %s,", counterKind))
	}
	if !dynamic.Group.Empty() {
		group, err := r.renderGroupReference(ctx, dynamic.Group)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Group: %s,", group))
	}
	if dynamic.Object.Kind() != game.ObjectReferenceNone {
		object, err := r.renderObjectReference(dynamic.Object)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Object: %s,", object))
	}
	if dynamic.Player != nil && dynamic.Player.Kind() != game.PlayerReferenceNone {
		player, err := r.renderPlayerReference(*dynamic.Player)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Player: func() *game.PlayerReference { ref := %s; return &ref }(),", player))
	}
	if dynamic.CardZone != zone.None {
		cardZone, err := renderZone(dynamic.CardZone)
		if err != nil {
			return "", err
		}
		ctx.need(importZone)
		fields = append(fields, fmt.Sprintf("CardZone: %s,", cardZone))
	}
	if dynamic.Selection != nil &&
		(!dynamic.Selection.Empty() || dynamic.Kind == game.DynamicAmountCountCardsInZone) {
		selection, err := r.renderSelection(ctx, *dynamic.Selection)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Selection: &%s,", selection))
	}
	if dynamic.ResultKey != "" {
		fields = append(fields, fmt.Sprintf("ResultKey: game.ResultKey(%q),", string(dynamic.ResultKey)))
	}
	if len(dynamic.Colors) > 0 {
		ctx.need(importColor)
		colorLits, err := colorValueLiterals(dynamic.Colors)
		if err != nil {
			return "", err
		}
		fields = append(fields, fmt.Sprintf("Colors: []color.Color{%s},", colorLits))
	}
	if dynamic.ColorFrom != "" {
		fields = append(fields, fmt.Sprintf("ColorFrom: game.ChoiceKey(%q),", string(dynamic.ColorFrom)))
	}
	return structLit("game.DynamicAmount", fields), nil
}

func renderDynamicAmountKind(kind game.DynamicAmountKind) (string, error) {
	switch kind {
	case game.DynamicAmountConstant:
		return "game.DynamicAmountConstant", nil
	case game.DynamicAmountX:
		return "game.DynamicAmountX", nil
	case game.DynamicAmountTargetPower:
		return "game.DynamicAmountTargetPower", nil
	case game.DynamicAmountTargetToughness:
		return "game.DynamicAmountTargetToughness", nil
	case game.DynamicAmountTargetManaValue:
		return "game.DynamicAmountTargetManaValue", nil
	case game.DynamicAmountTargetCounters:
		return "game.DynamicAmountTargetCounters", nil
	case game.DynamicAmountControllerLife:
		return "game.DynamicAmountControllerLife", nil
	case game.DynamicAmountControllerHandSize:
		return "game.DynamicAmountControllerHandSize", nil
	case game.DynamicAmountControllerGraveyardSize:
		return "game.DynamicAmountControllerGraveyardSize", nil
	case game.DynamicAmountControllerBasicLandTypeCount:
		return "game.DynamicAmountControllerBasicLandTypeCount", nil
	case game.DynamicAmountCountSelector:
		return "game.DynamicAmountCountSelector", nil
	case game.DynamicAmountCountCardsInZone:
		return "game.DynamicAmountCountCardsInZone", nil
	case game.DynamicAmountPreviousEffectResult:
		return "game.DynamicAmountPreviousEffectResult", nil
	case game.DynamicAmountOpponentCount:
		return "game.DynamicAmountOpponentCount", nil
	case game.DynamicAmountEventDamage:
		return "game.DynamicAmountEventDamage", nil
	case game.DynamicAmountEventCardCount:
		return "game.DynamicAmountEventCardCount", nil
	case game.DynamicAmountPreviousEffectExcessDamage:
		return "game.DynamicAmountPreviousEffectExcessDamage", nil
	case game.DynamicAmountObjectPower:
		return "game.DynamicAmountObjectPower", nil
	case game.DynamicAmountObjectToughness:
		return "game.DynamicAmountObjectToughness", nil
	case game.DynamicAmountObjectManaValue:
		return "game.DynamicAmountObjectManaValue", nil
	case game.DynamicAmountChosenNumber:
		return "game.DynamicAmountChosenNumber", nil
	case game.DynamicAmountObjectCounters:
		return "game.DynamicAmountObjectCounters", nil
	case game.DynamicAmountCapturedTargetManaValue:
		return "game.DynamicAmountCapturedTargetManaValue", nil
	case game.DynamicAmountGreatestPowerInGroup:
		return "game.DynamicAmountGreatestPowerInGroup", nil
	case game.DynamicAmountGreatestToughnessInGroup:
		return "game.DynamicAmountGreatestToughnessInGroup", nil
	case game.DynamicAmountGreatestManaValueInGroup:
		return "game.DynamicAmountGreatestManaValueInGroup", nil
	case game.DynamicAmountTotalPowerInGroup:
		return "game.DynamicAmountTotalPowerInGroup", nil
	case game.DynamicAmountTotalToughnessInGroup:
		return "game.DynamicAmountTotalToughnessInGroup", nil
	case game.DynamicAmountColorCountInGroup:
		return "game.DynamicAmountColorCountInGroup", nil
	case game.DynamicAmountSharedCreatureTypeCountInGroup:
		return "game.DynamicAmountSharedCreatureTypeCountInGroup", nil
	case game.DynamicAmountDevotion:
		return "game.DynamicAmountDevotion", nil
	case game.DynamicAmountSpellsCastThisTurn:
		return "game.DynamicAmountSpellsCastThisTurn", nil
	case game.DynamicAmountEventLifeChange:
		return "game.DynamicAmountEventLifeChange", nil
	default:
		return "", fmt.Errorf("render: unsupported dynamic amount kind %d", kind)
	}
}
