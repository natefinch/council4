package cardgen

import (
	"fmt"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// objectCharacteristicAmount builds the dynamic amount for a referenced object's
// own power, toughness, or mana value ("its power"/"its toughness"/"its mana
// value"), used by the life-rider sequence lowerer. It is intentionally separate
// from lowerDynamicAmount so that these object-characteristic forms stay
// fail-closed in every other dynamic-amount path; only the dedicated rider, which
// binds the referent, may lower them.
func objectCharacteristicAmount(kind compiler.DynamicAmountKind, object game.ObjectReference) (game.DynamicAmount, bool) {
	if len(object.Validate()) != 0 {
		return game.DynamicAmount{}, false
	}
	switch kind {
	case compiler.DynamicAmountSourcePower:
		return game.DynamicAmount{Kind: game.DynamicAmountObjectPower, Multiplier: 1, Object: object}, true
	case compiler.DynamicAmountSourceToughness:
		return game.DynamicAmount{Kind: game.DynamicAmountObjectToughness, Multiplier: 1, Object: object}, true
	case compiler.DynamicAmountSourceManaValue:
		return game.DynamicAmount{Kind: game.DynamicAmountObjectManaValue, Multiplier: 1, Object: object}, true
	default:
		return game.DynamicAmount{}, false
	}
}

// dieRollResultKey is the result key under which a RollDie instruction publishes
// its rolled value, consumed by a following "...equal to the result." amount
// (Ancient Copper Dragon and the Ancient Dragon dice cycle).
const dieRollResultKey = game.ResultKey("die-roll-result")

func lowerDynamicAmount(amount compiler.CompiledAmount, object game.ObjectReference) (game.DynamicAmount, bool) {
	dynamic, ok := lowerDynamicAmountKind(amount, object)
	if !ok {
		return game.DynamicAmount{}, false
	}
	dynamic.Addend = amount.Addend
	return dynamic, true
}

func lowerDynamicAmountKind(amount compiler.CompiledAmount, object game.ObjectReference) (game.DynamicAmount, bool) {
	if amount.Multiplier < 1 {
		return game.DynamicAmount{}, false
	}
	dynamic := game.DynamicAmount{Multiplier: amount.Multiplier}
	switch amount.DynamicKind {
	case compiler.DynamicAmountCount:
		if dynamic, ok := dynamicCardZoneAmount(amount.Selector(), amount.Multiplier); ok {
			return dynamic, true
		}
		selection, ok := dynamicAmountSelection(amount.Selector())
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountCountSelector
		dynamic.Group = game.BattlefieldGroup(selection)
	case compiler.DynamicAmountControllerLife:
		dynamic.Kind = game.DynamicAmountControllerLife
	case compiler.DynamicAmountControllerSpeed:
		dynamic.Kind = game.DynamicAmountControllerSpeed
	case compiler.DynamicAmountOpponentCount:
		dynamic.Kind = game.DynamicAmountOpponentCount
	case compiler.DynamicAmountOpponentControllingCount:
		selection, ok := dynamicAmountSelection(amount.Selector())
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountOpponentControllingCount
		dynamic.Group = game.BattlefieldGroup(selection)
	case compiler.DynamicAmountBasicLandTypes:
		dynamic.Kind = game.DynamicAmountControllerBasicLandTypeCount
	case compiler.DynamicAmountSourcePower:
		if len(object.Validate()) != 0 {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountObjectPower
		dynamic.Object = object
	case compiler.DynamicAmountSourceToughness:
		if len(object.Validate()) != 0 {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountObjectToughness
		dynamic.Object = object
	case compiler.DynamicAmountSourceCounterCount:
		if len(object.Validate()) != 0 || !amount.CounterKind.Valid() {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountObjectCounters
		dynamic.Object = object
		dynamic.CounterKind = amount.CounterKind
	case compiler.DynamicAmountGreatestPower, compiler.DynamicAmountGreatestToughness, compiler.DynamicAmountGreatestManaValue:
		selection, ok := dynamicAmountSelection(amount.Selector())
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = greatestInGroupKind(amount.DynamicKind)
		dynamic.Group = game.BattlefieldGroup(selection)
	case compiler.DynamicAmountTotalPower, compiler.DynamicAmountTotalToughness, compiler.DynamicAmountTotalManaValue:
		selection, ok := dynamicAmountSelection(amount.Selector())
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = totalInGroupKind(amount.DynamicKind)
		dynamic.Group = game.BattlefieldGroup(selection)
	case compiler.DynamicAmountColorCount:
		selection, ok := dynamicAmountSelection(amount.Selector())
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountColorCountInGroup
		dynamic.Group = game.BattlefieldGroup(selection)
	case compiler.DynamicAmountSharedCreatureTypeCount:
		selection, ok := dynamicAmountSelection(amount.Selector())
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountSharedCreatureTypeCountInGroup
		dynamic.Group = game.BattlefieldGroup(selection)
	case compiler.DynamicAmountDevotion:
		if len(amount.Colors) == 0 {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountDevotion
		dynamic.Colors = append([]color.Color(nil), amount.Colors...)
	case compiler.DynamicAmountSpellsCastThisTurn:
		dynamic.Kind = game.DynamicAmountSpellsCastThisTurn
	case compiler.DynamicAmountColorsOfManaSpent:
		dynamic.Kind = game.DynamicAmountColorsOfManaSpentToCast
	case compiler.DynamicAmountTimesKicked:
		dynamic.Kind = game.DynamicAmountTimesKicked
	case compiler.DynamicAmountOpponentsAttackedThisCombat:
		dynamic.Kind = game.DynamicAmountOpponentsAttackedThisCombat
	case compiler.DynamicAmountLifeLostThisTurn:
		dynamic.Kind = game.DynamicAmountLifeLostThisTurn
	case compiler.DynamicAmountLifeGainedThisTurn:
		dynamic.Kind = game.DynamicAmountLifeGainedThisTurn
	case compiler.DynamicAmountCardsDrawnThisTurn:
		dynamic.Kind = game.DynamicAmountCardsDrawnThisTurn
	case compiler.DynamicAmountMaxOf:
		operands, ok := lowerDynamicAmountOperands(amount.Operands, object)
		if !ok {
			return game.DynamicAmount{}, false
		}
		dynamic.Kind = game.DynamicAmountMaxOf
		dynamic.Operands = operands
	case compiler.DynamicAmountSacrificedPower:
		dynamic.Kind = game.DynamicAmountObjectPower
		dynamic.Object = game.SacrificedCostReference()
	case compiler.DynamicAmountSacrificedToughness:
		dynamic.Kind = game.DynamicAmountObjectToughness
		dynamic.Object = game.SacrificedCostReference()
	case compiler.DynamicAmountSacrificedManaValue:
		dynamic.Kind = game.DynamicAmountObjectManaValue
		dynamic.Object = game.SacrificedCostReference()
	case compiler.DynamicAmountDieRollResult:
		dynamic.Kind = game.DynamicAmountPreviousEffectResult
		dynamic.ResultKey = dieRollResultKey
	default:
		return game.DynamicAmount{}, false
	}
	return dynamic, true
}

// lowerDynamicAmountOperands lowers the operand list of a "whichever is greater"
// max combinator, requiring at least two operands and lowering each through
// lowerDynamicAmount so every recognized amount form composes. It fails closed
// when any operand is unrecognized.
func lowerDynamicAmountOperands(operands []compiler.CompiledAmount, object game.ObjectReference) ([]game.DynamicAmount, bool) {
	if len(operands) < 2 {
		return nil, false
	}
	lowered := make([]game.DynamicAmount, 0, len(operands))
	for i := range operands {
		dynamic, ok := lowerDynamicAmount(operands[i], object)
		if !ok {
			return nil, false
		}
		lowered = append(lowered, dynamic)
	}
	return lowered, true
}

// greatestInGroupKind maps a compiled greatest-characteristic amount kind to its
// runtime "greatest <characteristic> among group" sibling.
func greatestInGroupKind(kind compiler.DynamicAmountKind) game.DynamicAmountKind {
	switch kind {
	case compiler.DynamicAmountGreatestToughness:
		return game.DynamicAmountGreatestToughnessInGroup
	case compiler.DynamicAmountGreatestManaValue:
		return game.DynamicAmountGreatestManaValueInGroup
	default:
		return game.DynamicAmountGreatestPowerInGroup
	}
}

// totalInGroupKind maps a compiled total-characteristic amount kind to its
// runtime "total <characteristic> across group" sibling.
func totalInGroupKind(kind compiler.DynamicAmountKind) game.DynamicAmountKind {
	switch kind {
	case compiler.DynamicAmountTotalToughness:
		return game.DynamicAmountTotalToughnessInGroup
	case compiler.DynamicAmountTotalManaValue:
		return game.DynamicAmountTotalManaValueInGroup
	default:
		return game.DynamicAmountTotalPowerInGroup
	}
}

// dynamicAmountSelection projects the battlefield-count group selector of a
// dynamic amount ("for each other attacking creature you control") onto a
// Selection through the canonical projector. The guard enforces the count-group
// accept set that no SelectionMask dimension expresses: a countable permanent
// kind (or a multi-type union, or an unknown noun carrying a count
// characteristic), no "all" qualifier, and a you/opponent controller. It then
// delegates the field mapping to the canonical projector, which carries the
// combat, tapped, self-exclusion, and characteristic filters a count group does
// support. dynamicAmountSelectionMask drops the remaining canonical dimensions a
// count group never carries.
func dynamicAmountSelection(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.All {
		return game.Selection{}, false
	}
	_, known := dynamicBattlefieldRequiredType(selector.Kind)
	switch {
	case len(selector.RequiredTypesAny()) >= 2:
	case known:
	case selector.Kind == compiler.SelectorUnknown && selectorHasCountCharacteristic(selector):
	default:
		return game.Selection{}, false
	}
	switch selector.Controller {
	case compiler.ControllerAny, compiler.ControllerYou, compiler.ControllerOpponent:
	default:
		return game.Selection{}, false
	}
	return SelectionForSelectorMasked(selector, dynamicAmountSelectionMask)
}

// dynamicAmountSelectionMask drops the canonical dimensions a battlefield count
// group never carries: the excluded supertype, kind-agnostic counter, "aren't of
// the chosen type" exclusion, conjunctive type set, and per-object token state.
// It fails closed on a source-relative power comparison (a count group has no
// source permanent to compare against, so the predecessor projector rejected
// that filter rather than dropping it) and on a historic disjunction, which a
// battlefield count cannot represent through the count lowering and must not
// silently drop.
var dynamicAmountSelectionMask = SelectionMask{}.Ignoring(
	DimExcludedSupertype,
	DimMatchAnyCounter,
	DimSubtypeChoiceExcluded,
	DimConjunctiveTypes,
	DimNonToken,
	DimTokenOnly,
).Rejecting(
	DimPowerVsSource,
	DimRequiredName,
	DimHistoric,
)

func dynamicBattlefieldRequiredType(kind compiler.SelectorKind) (types.Card, bool) {
	switch kind {
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorLand:
		return types.Land, true
	case compiler.SelectorPermanent:
		return "", true
	default:
		return "", false
	}
}

func dynamicCardZoneAmount(selector compiler.CompiledSelector, multiplier int) (game.DynamicAmount, bool) {
	if selector.Controller != compiler.ControllerYou {
		return game.DynamicAmount{}, false
	}
	if selector.Zone != zone.Graveyard && selector.Zone != zone.Hand {
		return game.DynamicAmount{}, false
	}
	requiredType, known := dynamicZoneRequiredType(selector.Kind)
	if !known {
		return game.DynamicAmount{}, false
	}
	selection, ok := dynamicCountCharacteristics(selector)
	if !ok {
		return game.DynamicAmount{}, false
	}
	if requiredType != "" {
		selection.RequiredTypes = []types.Card{requiredType}
	}
	player := game.ControllerReference()
	return game.DynamicAmount{
		Kind:       game.DynamicAmountCountCardsInZone,
		Multiplier: multiplier,
		Player:     &player,
		CardZone:   selector.Zone,
		Selection:  &selection,
	}, true
}

func dynamicZoneRequiredType(kind compiler.SelectorKind) (types.Card, bool) {
	switch kind {
	case compiler.SelectorCard:
		return "", true
	case compiler.SelectorArtifact:
		return types.Artifact, true
	case compiler.SelectorCreature:
		return types.Creature, true
	case compiler.SelectorEnchantment:
		return types.Enchantment, true
	case compiler.SelectorLand:
		return types.Land, true
	default:
		return "", false
	}
}

// DO-NOT-COPY(filter): wraps selectorCharacteristics for the card-zone count
// contexts (cards in a graveyard or hand) the battlefield-only canonical
// projector fails closed on, and defers the selector Kind and Controller to its
// callers; prefer SelectionForSelectorMasked for new code. (retire: #1393)
//
// dynamicCountCharacteristics maps the characteristic filters of a compiled
// count selector onto a runtime Selection, returning false for any filter the
// executable backend cannot represent exactly so unsupported wordings stay
// rejected. It deliberately ignores the selector Kind and Controller, which
// callers translate per context (battlefield required type versus card zone).
func dynamicCountCharacteristics(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.All || selector.Another || selector.Other ||
		selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped {
		return game.Selection{}, false
	}
	selection, ok := selectorCharacteristics(selector)
	if !ok {
		return game.Selection{}, false
	}
	if selector.MatchManaValue {
		if selector.ManaValueX {
			return game.Selection{}, false
		}
		selection.ManaValue = opt.Val(selector.ManaValue)
	}
	if selector.MatchPower {
		selection.Power = opt.Val(selector.Power)
	}
	if selector.MatchToughness {
		selection.Toughness = opt.Val(selector.Toughness)
	}
	if selector.Historic {
		selection.AnyOf = append(selection.AnyOf, historicSelectionAlternatives()...)
	}
	return selection, true
}

// DO-NOT-COPY(filter): maps only a selector's characteristic filters, deferring
// the Kind, Controller, combat, tapped, and zone dimensions to its callers
// (including card-zone contexts), so it intentionally produces a partial
// selection the full canonical projector cannot; prefer
// SelectionForSelectorMasked for new code. (retire: #1393)
//
// selectorCharacteristics maps the characteristic filters of a compiled selector
// (colors, colorless/multicolored, keyword, excluded types, supertypes,
// subtypes, excluded colors, and a disjunctive required-type union) onto a
// runtime Selection, returning false for any characteristic the executable
// backend cannot represent exactly. It ignores the selector Kind, Controller,
// combat, tapped, and "other" flags, which callers translate per context.
func selectorCharacteristics(selector compiler.CompiledSelector) (game.Selection, bool) {
	if selector.PowerLessThanSource || selector.PowerGreaterThanSource {
		// A source-relative "with lesser/greater power" comparison is meaningful
		// only for a targeted permanent (Mentor), where the target path carries
		// the source. Group, count, and card-zone contexts have no source to
		// compare against, so reject it rather than silently dropping the filter.
		return game.Selection{}, false
	}
	selection := game.Selection{
		Colorless:       selector.Colorless,
		Multicolored:    selector.Multicolored,
		EnteredThisTurn: selector.EnteredThisTurn,
	}
	if selector.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.Keyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.Keyword = keyword
	}
	if selector.ExcludedKeyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(selector.ExcludedKeyword)
		if !ok {
			return game.Selection{}, false
		}
		selection.ExcludedKeyword = keyword
	}
	if selector.MatchCounter {
		selection.MatchCounter = true
		selection.RequiredCounter = selector.RequiredCounter
	}
	if union := selector.RequiredTypesAny(); len(union) > 0 {
		selection.RequiredTypesAny = append([]types.Card(nil), union...)
	}
	if excluded := selector.ExcludedTypes(); len(excluded) > 0 {
		selection.ExcludedTypes = append([]types.Card(nil), excluded...)
	}
	if supertypes := selector.Supertypes(); len(supertypes) > 0 {
		selection.Supertypes = append([]types.Super(nil), supertypes...)
	}
	if subtypes := selector.SubtypesAny(); len(subtypes) > 0 {
		selection.SubtypesAny = append([]types.Sub(nil), subtypes...)
	}
	if excludedSubtypes := selector.ExcludedSubtypes(); len(excludedSubtypes) > 0 {
		if len(excludedSubtypes) > 1 {
			return game.Selection{}, false
		}
		selection.ExcludedSubtype = excludedSubtypes[0]
	}
	if colors := selector.ColorsAny(); len(colors) > 0 {
		selection.ColorsAny = append([]color.Color(nil), colors...)
	}
	if excludedColors := selector.ExcludedColors(); len(excludedColors) > 0 {
		selection.ExcludedColors = append([]color.Color(nil), excludedColors...)
	}
	if selector.SubtypeFromEntryChoice {
		selection.SubtypeChoice = game.SubtypeChoiceSourceEntry
	}
	if selector.SubtypeFromChosenType {
		selection.SubtypeChoice = game.SubtypeChoiceResolution
	}
	return selection, true
}

func selectorHasCountCharacteristic(selector compiler.CompiledSelector) bool {
	return selector.Colorless || selector.Multicolored ||
		selector.Keyword != parser.KeywordUnknown ||
		selector.ExcludedKeyword != parser.KeywordUnknown ||
		selector.MatchCounter ||
		selector.MatchManaValue || selector.MatchPower || selector.MatchToughness ||
		selector.SubtypeFromEntryChoice ||
		selector.SubtypeFromChosenType ||
		selector.EnteredThisTurn ||
		len(selector.SubtypesAny()) > 0 ||
		len(selector.ExcludedSubtypes()) > 0 ||
		len(selector.Supertypes()) > 0 ||
		len(selector.ColorsAny()) > 0 ||
		len(selector.ExcludedColors()) > 0 ||
		len(selector.RequiredTypesAny()) > 0 ||
		len(selector.ExcludedTypes()) > 0
}

func runtimeKeyword(keyword parser.KeywordKind) (game.Keyword, bool) {
	switch keyword {
	case parser.KeywordCycling:
		return game.Cycling, true
	case parser.KeywordEquip:
		return game.Equip, true
	case parser.KeywordFlying:
		return game.Flying, true
	case parser.KeywordReach:
		return game.Reach, true
	case parser.KeywordTrample:
		return game.Trample, true
	case parser.KeywordLifelink:
		return game.Lifelink, true
	case parser.KeywordDeathtouch:
		return game.Deathtouch, true
	case parser.KeywordIndestructible:
		return game.Indestructible, true
	case parser.KeywordHaste:
		return game.Haste, true
	case parser.KeywordMenace:
		return game.Menace, true
	case parser.KeywordVigilance:
		return game.Vigilance, true
	case parser.KeywordDefender:
		return game.Defender, true
	case parser.KeywordFirstStrike:
		return game.FirstStrike, true
	case parser.KeywordDoubleStrike:
		return game.DoubleStrike, true
	case parser.KeywordFlash:
		return game.Flash, true
	case parser.KeywordHexproof:
		return game.Hexproof, true
	case parser.KeywordShroud:
		return game.Shroud, true
	case parser.KeywordDevoid:
		return game.Devoid, true
	case parser.KeywordProwess:
		return game.Prowess, true
	case parser.KeywordExalted:
		return game.Exalted, true
	case parser.KeywordEvolve:
		return game.Evolve, true
	case parser.KeywordWither:
		return game.Wither, true
	case parser.KeywordInfect:
		return game.Infect, true
	case parser.KeywordToxic:
		return game.Toxic, true
	case parser.KeywordUndying:
		return game.Undying, true
	case parser.KeywordPersist:
		return game.Persist, true
	case parser.KeywordRiot:
		return game.Riot, true
	case parser.KeywordUnleash:
		return game.Unleash, true
	case parser.KeywordFear:
		return game.Fear, true
	case parser.KeywordSkulk:
		return game.Skulk, true
	case parser.KeywordIntimidate:
		return game.Intimidate, true
	case parser.KeywordRetrace:
		return game.Retrace, true
	default:
		return game.KeywordNone, false
	}
}

// lowerEventCardCountAmount lowers a "for each card discarded/drawn this way"
// amount into a DynamicAmountEventCardCount. It succeeds only inside a draw or
// discard triggered ability (ctx.triggerCardCountEvent records the triggering
// event kind), keeping the amount closed in spell and non-matching contexts
// where no triggering card count exists.
func lowerEventCardCountAmount(ctx contentCtx, amount compiler.CompiledAmount) (game.DynamicAmount, bool) {
	switch ctx.triggerCardCountEvent {
	case game.EventCardDrawn, game.EventCardDiscarded, game.EventCycled:
	default:
		return game.DynamicAmount{}, false
	}
	multiplier := max(amount.Multiplier, 1)
	return game.DynamicAmount{
		Kind:       game.DynamicAmountEventCardCount,
		Multiplier: multiplier,
	}, true
}

// lowerEventCombatDamageAmount lowers a "that many" token-count amount into a
// DynamicAmountEventDamage. It succeeds only inside a combat-damage triggered
// ability (ctx.triggerEvent records the triggering event kind), keeping the
// amount closed in spell and non-matching contexts where no triggering combat
// damage quantity exists.
func lowerEventCombatDamageAmount(ctx contentCtx, amount compiler.CompiledAmount) (game.DynamicAmount, bool) {
	if ctx.triggerEvent != game.EventDamageDealt {
		return game.DynamicAmount{}, false
	}
	multiplier := max(amount.Multiplier, 1)
	return game.DynamicAmount{
		Kind:       game.DynamicAmountEventDamage,
		Multiplier: multiplier,
	}, true
}

// lowerEventLifeChangeAmount lowers a "that much life" amount into a
// DynamicAmountEventLifeChange. It succeeds only inside a life-gain or life-loss
// triggered ability (ctx.triggerEvent records the triggering event kind),
// keeping the amount closed in spell and non-matching contexts where no
// triggering life quantity exists.
func lowerEventLifeChangeAmount(ctx contentCtx, amount compiler.CompiledAmount) (game.DynamicAmount, bool) {
	switch ctx.triggerEvent {
	case game.EventLifeGained, game.EventLifeLost:
	default:
		return game.DynamicAmount{}, false
	}
	multiplier := max(amount.Multiplier, 1)
	return game.DynamicAmount{
		Kind:       game.DynamicAmountEventLifeChange,
		Multiplier: multiplier,
	}, true
}

// lowerEventCounterCountAmount lowers a "that many" card-count amount into a
// DynamicAmountEventCounterCount. It succeeds only inside a counter-placement
// triggered ability (ctx.triggerEvent records the triggering event kind),
// keeping the amount closed in spell and non-matching contexts where no
// triggering counter quantity exists.
func lowerEventCounterCountAmount(ctx contentCtx, amount compiler.CompiledAmount) (game.DynamicAmount, bool) {
	if ctx.triggerEvent != game.EventCountersAdded {
		return game.DynamicAmount{}, false
	}
	multiplier := max(amount.Multiplier, 1)
	return game.DynamicAmount{
		Kind:       game.DynamicAmountEventCounterCount,
		Multiplier: multiplier,
	}, true
}

// triggeringEventQuantityKind reports whether a compiled dynamic amount kind is
// a "that much"/"that many" anaphor that reads a quantity from the triggering
// event. The parser pins each such phrase to one historically chosen kind
// (EventCardCount, TriggeringLifeChange, TriggeringCombatDamage, or
// TriggeringCounterCount) without knowing which event actually fired, so every
// one of these kinds denotes the same idea: the triggering event's quantity. The
// enclosing trigger event resolves it at lowering time
// (lowerTriggeringEventQuantityAmount), keeping the parser text-blind.
func triggeringEventQuantityKind(kind compiler.DynamicAmountKind) bool {
	switch kind {
	case compiler.DynamicAmountEventCardCount,
		compiler.DynamicAmountTriggeringLifeChange,
		compiler.DynamicAmountTriggeringCombatDamage,
		compiler.DynamicAmountTriggeringCounterCount:
		return true
	default:
		return false
	}
}

// lowerTriggeringEventQuantityAmount resolves a "that much"/"that many"
// triggering-event anaphor onto the runtime DynamicAmount for whichever event
// actually fired, independent of which historical kind the parser pinned. A
// draw, discard, or cycle trigger reads its card count; a damage trigger reads
// the damage dealt; a life-change trigger reads the life gained or lost; a
// counter trigger reads the counters added. Outside one of those triggered
// contexts the anaphor has no source and stays rejected (ok=false).
func lowerTriggeringEventQuantityAmount(ctx contentCtx, amount compiler.CompiledAmount) (game.DynamicAmount, bool) {
	multiplier := max(amount.Multiplier, 1)
	switch ctx.triggerCardCountEvent {
	case game.EventCardDrawn, game.EventCardDiscarded, game.EventCycled:
		return game.DynamicAmount{Kind: game.DynamicAmountEventCardCount, Multiplier: multiplier}, true
	}
	switch ctx.triggerEvent {
	case game.EventDamageDealt:
		return game.DynamicAmount{Kind: game.DynamicAmountEventDamage, Multiplier: multiplier}, true
	case game.EventLifeGained, game.EventLifeLost:
		return game.DynamicAmount{Kind: game.DynamicAmountEventLifeChange, Multiplier: multiplier}, true
	case game.EventCountersAdded:
		return game.DynamicAmount{Kind: game.DynamicAmountEventCounterCount, Multiplier: multiplier}, true
	default:
		return game.DynamicAmount{}, false
	}
}

func exactDamageAmountReferences(amount compiler.CompiledAmount, references []compiler.CompiledReference) bool {
	if amount.DynamicKind != compiler.DynamicAmountSourcePower {
		_, ok := lowerDamageSourceReference(references)
		return ok
	}
	if len(references) != 2 ||
		references[1].Span != amount.ReferenceSpan {
		return false
	}
	// The damage source (references[0], "this creature") and the amount referent
	// (references[1], "its"/"that creature's") may bind different objects: the
	// source deals the damage while the amount reads the entering creature's
	// power. Each must lower independently, but they need not share a binding.
	_, sourceOK := lowerDamageSourceReference(references[:1])
	_, amountOK := lowerDamageSourceReference(references[1:])
	return sourceOK && amountOK
}

// lowerDamageAmountObject resolves the object whose power feeds a dynamic damage
// amount. It binds to the amount's own referent ("its" for the source,
// "that creature's" for the triggering permanent) so "deals damage equal to that
// creature's power" reads the entering creature rather than the damage source.
func lowerDamageAmountObject(amount compiler.CompiledAmount, references []compiler.CompiledReference) (game.ObjectReference, bool) {
	if amount.DynamicKind != compiler.DynamicAmountSourcePower {
		return game.ObjectReference{}, false
	}
	for i := range references {
		if references[i].Span != amount.ReferenceSpan {
			continue
		}
		return lowerObjectReference(references[i], referenceLoweringContext{
			AllowSource: true,
			AllowEvent:  true,
		})
	}
	return game.ObjectReference{}, false
}

func lowerDamageSourceReference(references []compiler.CompiledReference) (game.ObjectReference, bool) {
	if len(references) != 1 {
		return game.ObjectReference{}, false
	}
	return lowerObjectReference(references[0], referenceLoweringContext{
		AllowSource: true,
		AllowEvent:  true,
	})
}

func validModifyPTAmount(effect *compiler.CompiledEffect, referenceCount int) bool {
	if effect.Context != parser.EffectContextTarget && effect.Context != parser.EffectContextPriorSubject {
		return false
	}
	amount := effect.Amount
	if amount.DynamicKind == compiler.DynamicAmountNone {
		return true
	}
	if referenceCount != 0 || amount.DynamicKind == compiler.DynamicAmountSourcePower {
		return false
	}
	return dynamicModifyPTFormValid(effect)
}

// dynamicModifyPTFormValid reports whether a dynamic power/toughness amount uses
// one of the two recognized formula shapes ("… for each …" or "where X is …")
// with deltas the dynamic machinery can render. It is the shared core of
// validModifyPTAmount (target pumps) and the referenced/source self-pump path,
// neither of which differs in how a dynamic amount's deltas must be shaped.
func dynamicModifyPTFormValid(effect *compiler.CompiledEffect) bool {
	amount := effect.Amount
	switch amount.DynamicForm {
	case compiler.DynamicAmountForEach:
		return effect.PowerDelta.Known && effect.ToughnessDelta.Known &&
			dynamicPTMultiplierMatches(amount.Multiplier, effect.PowerDelta, effect.ToughnessDelta)
	case compiler.DynamicAmountWhereX:
		powerOK := effect.PowerDelta.VariableX || effect.PowerDelta.Known
		toughnessOK := effect.ToughnessDelta.VariableX || effect.ToughnessDelta.Known
		return powerOK && toughnessOK &&
			(effect.PowerDelta.VariableX || effect.ToughnessDelta.VariableX)
	default:
		return false
	}
}

// referencedModifyPTQuantities computes the power and toughness deltas for a
// self- or referenced-object power/toughness change ("This creature gets …",
// "it gets …"). A fixed amount yields signed fixed deltas; a dynamic "… for each
// …" or "where X is the number of …" amount yields dynamic deltas counted over
// the amount's own subject. countObject backs an object-relative dynamic amount
// at runtime; it is unused for controller- or zone-counted amounts. It returns
// ok=false for the source-power form ("where X is its power"), whose "its"
// referent the executable backend does not yet bind, and for any dynamic shape
// the runtime cannot model, keeping those fail-closed.
func referencedModifyPTQuantities(
	effect *compiler.CompiledEffect,
	countObject game.ObjectReference,
) (power, toughness game.Quantity, ok bool) {
	if effect.Amount.DynamicKind == compiler.DynamicAmountNone {
		if !effect.PowerDelta.Known || !effect.ToughnessDelta.Known {
			return game.Quantity{}, game.Quantity{}, false
		}
		return game.Fixed(compiledSignedAmountValue(effect.PowerDelta)),
			game.Fixed(compiledSignedAmountValue(effect.ToughnessDelta)), true
	}
	if effect.Amount.DynamicKind == compiler.DynamicAmountSourcePower ||
		!dynamicModifyPTFormValid(effect) {
		return game.Quantity{}, game.Quantity{}, false
	}
	dynamic, ok := lowerDynamicAmount(effect.Amount, countObject)
	if !ok {
		return game.Quantity{}, game.Quantity{}, false
	}
	switch effect.Amount.DynamicForm {
	case compiler.DynamicAmountWhereX:
		return whereXSignedQuantity(&dynamic, effect.PowerDelta),
			whereXSignedQuantity(&dynamic, effect.ToughnessDelta), true
	case compiler.DynamicAmountForEach:
		return dynamicSignedQuantity(&dynamic, effect.PowerDelta),
			dynamicSignedQuantity(&dynamic, effect.ToughnessDelta), true
	default:
		return game.Quantity{}, game.Quantity{}, false
	}
}

// sourcePowerReferences splits the references of a "where X is its power"
// power/toughness pump into the dynamic power referent and the remaining subject
// references. The power referent is the lone reference whose span matches the
// amount's referent span (the "its"/"this creature's"/"<name>'s" that names the
// permanent whose power supplies X). The subject references are whatever pumps
// the effect addresses (the source itself, the triggering permanent, or a prior
// clause's target); a target-context pump carries no subject reference because
// its subject is the target slot. It returns ok=false unless exactly one
// reference is the power referent so a malformed reference set fails closed.
func sourcePowerReferences(effect *compiler.CompiledEffect) (power compiler.CompiledReference, subjects []compiler.CompiledReference, ok bool) {
	found := false
	for _, reference := range effect.References {
		if reference.Span == effect.Amount.ReferenceSpan {
			if found {
				return compiler.CompiledReference{}, nil, false
			}
			power = reference
			found = true
			continue
		}
		subjects = append(subjects, reference)
	}
	return power, subjects, found
}

func dynamicPTMultiplierMatches(
	multiplier int,
	power, toughness compiler.CompiledSignedAmount,
) bool {
	matches := func(amount compiler.CompiledSignedAmount) bool {
		return amount.Value == 0 || amount.Value == multiplier
	}
	return multiplier > 0 && matches(power) && matches(toughness)
}

func dynamicSignedQuantity(
	dynamic *game.DynamicAmount,
	amount compiler.CompiledSignedAmount,
) game.Quantity {
	if amount.Value == 0 {
		return game.Fixed(0)
	}
	value := *dynamic
	if amount.Negative {
		value.Multiplier = -value.Multiplier
	}
	return game.Dynamic(value)
}

// whereXSignedQuantity lowers one power/toughness side of a "where X is …" pump.
// A variable "X" side becomes the dynamic amount (negated for "-X"); a fixed side
// (as in the "+0" of "+X/+0") becomes its signed fixed value.
func whereXSignedQuantity(
	dynamic *game.DynamicAmount,
	side compiler.CompiledSignedAmount,
) game.Quantity {
	if !side.VariableX {
		return game.Fixed(compiledSignedAmountValue(side))
	}
	value := *dynamic
	if side.Negative {
		value.Multiplier = -value.Multiplier
	}
	return game.Dynamic(value)
}

func fixedNumberSyntax(token shared.Token, atoms parser.Atoms, amount int) bool {
	if token.Kind == shared.Integer {
		return token.Text == fmt.Sprint(amount)
	}
	value, ok := atoms.CardinalAt(token.Span)
	return ok && value == amount
}

func singleSelfReference(references []compiler.CompiledReference) bool {
	return len(references) == 1 && references[0].Binding == compiler.ReferenceBindingSource
}

func damageTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !target.Exact || !targetCardinalityIsOne(target) {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: target.Text,
	}
	switch target.Selector.Kind {
	case compiler.SelectorAny:
		// "any other target"/"any another target" as a lone target excludes the
		// ability's source, a meaning the bare "any target" spec cannot express;
		// reject it so single-target damage stays faithful. The two-target damage
		// rider handles its own "other" (distinct-from-prior-target) separately.
		if target.Selector.Other || target.Selector.Another {
			return game.TargetSpec{}, false
		}
		spec.Allow = game.TargetAllowPermanent | game.TargetAllowPlayer
	case compiler.SelectorCreature, compiler.SelectorPlaneswalker, compiler.SelectorBattle:
		permanent, ok := permanentTargetSpec(target)
		if !ok {
			return game.TargetSpec{}, false
		}
		return permanent, true
	case compiler.SelectorPlayer:
		if target.Selector.PlayerOrPlaneswalker {
			spec.Allow = game.TargetAllowPlayer | game.TargetAllowPermanent
			spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Planeswalker}}
			return spec, true
		}
		spec.Allow = game.TargetAllowPlayer
	case compiler.SelectorOpponent:
		spec.Allow = game.TargetAllowPlayer
		if target.Selector.PlayerOrPlaneswalker {
			spec.Allow |= game.TargetAllowPermanent
			spec.Predicate = game.TargetPredicate{
				Player:         game.PlayerOpponent,
				PermanentTypes: []types.Card{types.Planeswalker},
			}
			return spec, true
		}
		spec.Predicate = game.TargetPredicate{Player: game.PlayerOpponent}
	default:
		return game.TargetSpec{}, false
	}
	return spec, true
}

func permanentTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !targetCardinalityIsOne(target) {
		return game.TargetSpec{}, false
	}
	return permanentTargetSpecWithCardinality(target)
}

// permanentTargetSpecWithCardinality builds a permanent TargetSpec that carries
// the target's own MinTargets/MaxTargets range, supporting plural ("two target
// creatures") and optional ("up to N target creatures") cardinalities in
// addition to the single-target form. permanentTargetSpec keeps the
// single-target gate for callers that only lower one target.
func permanentTargetSpecWithCardinality(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !target.Exact || target.Cardinality.Max < 1 || target.Cardinality.Min < 0 ||
		target.Cardinality.Min > target.Cardinality.Max {
		return game.TargetSpec{}, false
	}
	spec := game.TargetSpec{
		MinTargets: target.Cardinality.Min,
		MaxTargets: target.Cardinality.Max,
		Allow:      game.TargetAllowPermanent,
	}
	if len(target.Selector.Alternatives) > 0 {
		return alternativePermanentTargetSpec(&target, &spec)
	}
	switch target.Selector.Kind {
	case compiler.SelectorUnknown:
		// A bare subtype noun ("target Soldier you control") selects any
		// permanent carrying that subtype, with no card-type restriction. The
		// subtype filter below supplies the constraint; without one this is not
		// a recognized permanent target.
		if len(target.Selector.SubtypesAny()) == 0 {
			return game.TargetSpec{}, false
		}

	case compiler.SelectorArtifact:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Artifact}}
	case compiler.SelectorCreature:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Creature}}
	case compiler.SelectorEnchantment:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Enchantment}}
	case compiler.SelectorLand:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Land}}
	case compiler.SelectorPermanent:
	case compiler.SelectorPlaneswalker:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Planeswalker}}
	case compiler.SelectorBattle:
		spec.Predicate = game.TargetPredicate{PermanentTypes: []types.Card{types.Battle}}
	default:
		return game.TargetSpec{}, false
	}
	if (target.Selector.Tapped && target.Selector.Untapped) ||
		((target.Selector.Tapped || target.Selector.Untapped) &&
			(target.Selector.Attacking || target.Selector.Blocking)) ||
		selectorHasUnsupportedPermanentFilters(target.Selector) {
		return game.TargetSpec{}, false
	}
	if union := target.Selector.RequiredTypesAny(); len(union) > 0 {
		// A conjunctive type set ("artifact creature") requires every listed type
		// at once; the flag routes the same type list through the all-of filter
		// instead of the default any-of match.
		spec.Predicate.PermanentTypes = append([]types.Card(nil), union...)
		spec.Predicate.PermanentTypesConjunctive = target.Selector.ConjunctiveTypes
	}
	if excludedTypes := target.Selector.ExcludedTypes(); len(excludedTypes) > 0 {
		spec.Predicate.ExcludedTypes = append([]types.Card(nil), excludedTypes...)
	}
	if supertypes := target.Selector.Supertypes(); len(supertypes) > 0 {
		spec.Predicate.Supertypes = append([]types.Super(nil), supertypes...)
	}
	if excludedSupertypes := target.Selector.ExcludedSupertypes(); len(excludedSupertypes) > 0 {
		spec.Predicate.ExcludedSupertype = excludedSupertypes[0]
	}
	if subtypes := target.Selector.SubtypesAny(); len(subtypes) > 0 {
		spec.Predicate.Subtypes = append([]types.Sub(nil), subtypes...)
	}
	if colors := target.Selector.ColorsAny(); len(colors) > 0 {
		spec.Predicate.Colors = append([]color.Color(nil), colors...)
	}
	if excludedColors := target.Selector.ExcludedColors(); len(excludedColors) > 0 {
		spec.Predicate.ExcludedColors = append([]color.Color(nil), excludedColors...)
	}
	if target.Selector.Keyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(target.Selector.Keyword)
		if !ok {
			return game.TargetSpec{}, false
		}
		spec.Predicate.Keyword = keyword
	}
	if target.Selector.ExcludedKeyword != parser.KeywordUnknown {
		keyword, ok := runtimeKeyword(target.Selector.ExcludedKeyword)
		if !ok {
			return game.TargetSpec{}, false
		}
		spec.Predicate.ExcludedKeyword = keyword
	}
	if target.Selector.MatchManaValue {
		if target.Selector.ManaValueX {
			return game.TargetSpec{}, false
		}
		spec.Predicate.ManaValue = opt.Val(target.Selector.ManaValue)
	}
	if target.Selector.MatchPower {
		spec.Predicate.Power = opt.Val(target.Selector.Power)
	}
	if target.Selector.PowerLessThanSource {
		spec.Predicate.PowerLessThanSource = true
	}
	if target.Selector.PowerGreaterThanSource {
		spec.Predicate.PowerGreaterThanSource = true
	}
	if target.Selector.MatchToughness {
		spec.Predicate.Toughness = opt.Val(target.Selector.Toughness)
	}
	if target.Selector.Another || target.Selector.Other {
		spec.Predicate.Another = true
	}
	if target.Selector.TokenOnly {
		spec.Predicate.TokenOnly = true
	}
	if target.Selector.NonToken {
		spec.Predicate.NonToken = true
	}
	if target.Selector.NameUniqueAmongControlled {
		spec.Predicate.NameUniqueAmongControlled = true
	}

	switch {
	case target.Selector.Attacking && target.Selector.Blocking:
		spec.Predicate.CombatState = game.CombatStateAttackingOrBlocking
	case target.Selector.Attacking:
		spec.Predicate.CombatState = game.CombatStateAttacking
	case target.Selector.Blocking:
		spec.Predicate.CombatState = game.CombatStateBlocking
	case target.Selector.Tapped:
		spec.Predicate.Tapped = game.TriTrue
	case target.Selector.Untapped:
		spec.Predicate.Tapped = game.TriFalse
	default:
	}
	switch target.Selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		spec.Predicate.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		spec.Predicate.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		spec.Predicate.Controller = game.ControllerNotYou
	default:
		return game.TargetSpec{}, false
	}
	spec.Constraint = lowerFirst(target.Text)
	return spec, true
}

func alternativePermanentTargetSpec(target *compiler.CompiledTarget, spec *game.TargetSpec) (game.TargetSpec, bool) {
	selector := &target.Selector
	if selector.Kind != compiler.SelectorPermanent ||
		selectorHasUnsupportedPermanentFilters(*selector) ||
		selector.Another || selector.Other ||
		selector.Attacking || selector.Blocking ||
		selector.Tapped || selector.Untapped {
		return game.TargetSpec{}, false
	}
	selection := game.Selection{}
	switch selector.Controller {
	case compiler.ControllerAny:
	case compiler.ControllerYou:
		selection.Controller = game.ControllerYou
	case compiler.ControllerOpponent:
		selection.Controller = game.ControllerOpponent
	case compiler.ControllerNotYou:
		selection.Controller = game.ControllerNotYou
	default:
		return game.TargetSpec{}, false
	}
	for i := range selector.Alternatives {
		alternativeSpec, ok := permanentTargetSpecWithCardinality(compiler.CompiledTarget{
			Cardinality: compiler.TargetCardinality{Min: 1, Max: 1},
			Selector:    selector.Alternatives[i],
			Exact:       true,
		})
		if !ok || alternativeSpec.Selection.Exists {
			return game.TargetSpec{}, false
		}
		selection.AnyOf = append(selection.AnyOf, alternativeSpec.Predicate.Selection())
	}
	spec.Selection = opt.Val(selection)
	return *spec, true
}

// selectorHasUnsupportedPermanentFilters reports whether a permanent target
// selector carries a characteristic the runtime TargetPredicate cannot represent
// exactly. Subtypes, supertypes, colors, excluded colors, a recognized keyword,
// mana value, power, and toughness comparisons all map onto the predicate, so
// only zone restrictions and the colorless/multicolored color shapes (which the
// predicate cannot express) stay rejected, keeping unsupported wordings closed.
func selectorHasUnsupportedPermanentFilters(selector compiler.CompiledSelector) bool {
	return selector.Zone != zone.None ||
		selector.Colorless ||
		selector.Multicolored ||
		selector.Historic
}

func stackSpellTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !targetCardinalityIsOne(target) ||
		target.Selector.Another || target.Selector.Other ||
		target.Selector.Controller != compiler.ControllerAny ||
		target.Selector.Attacking || target.Selector.Blocking ||
		target.Selector.Tapped || target.Selector.Untapped ||
		len(target.Selector.Supertypes()) != 0 ||
		len(target.Selector.SubtypesAny()) != 0 ||
		target.Selector.Keyword != parser.KeywordUnknown ||
		target.Selector.Zone != zone.None ||
		target.Selector.MatchPower ||
		target.Selector.MatchToughness {
		return game.TargetSpec{}, false
	}
	if target.Selector.Kind != compiler.SelectorSpell {
		return game.TargetSpec{}, false
	}
	required := target.Selector.RequiredTypesAny()
	excluded := target.Selector.ExcludedTypes()
	colors := target.Selector.ColorsAny()
	excludedColors := target.Selector.ExcludedColors()
	if len(excluded) > 1 || len(required) > 0 && len(excluded) > 0 {
		return game.TargetSpec{}, false
	}
	predicate := game.TargetPredicate{
		StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
		ExcludedSpellCardTypes: append([]types.Card(nil), excluded...),
	}
	if target.Selector.MatchManaValue {
		if target.Selector.ManaValueX {
			return game.TargetSpec{}, false
		}
		predicate.ManaValue = opt.Val(target.Selector.ManaValue)
	}
	if len(required) == 1 {
		predicate.SpellCardTypes = append([]types.Card(nil), required...)
	} else if len(required) > 1 {
		predicate.SpellCardTypesAny = append([]types.Card(nil), required...)
	}
	// Color qualifiers stand alone: the supported wordings ("blue", "nonblue",
	// "colorless", "multicolored" spell) carry no card-type filter, so reject any
	// mix of a color shape with a type filter or with another color shape.
	hasTypeFilter := len(required) > 0 || len(excluded) > 0
	colorShapes := len(colors) + len(excludedColors)
	if target.Selector.Colorless {
		colorShapes++
	}
	if target.Selector.Multicolored {
		colorShapes++
	}
	if colorShapes > 0 {
		if hasTypeFilter || colorShapes > 1 || len(colors) > 1 || len(excludedColors) > 1 {
			return game.TargetSpec{}, false
		}
		switch {
		case len(colors) == 1:
			predicate.SpellColors = append([]color.Color(nil), colors...)
		case len(excludedColors) == 1:
			predicate.SpellExcludedColors = append([]color.Color(nil), excludedColors...)
		case target.Selector.Colorless:
			predicate.SpellColorless = true
		case target.Selector.Multicolored:
			predicate.SpellMulticolored = true
		default:
		}
	}
	spec := game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Allow:      game.TargetAllowStackObject,
		Predicate:  predicate,
		Constraint: lowerFirst(target.Text),
	}
	return spec, true
}

func counterAbilityTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if !targetCardinalityIsOne(target) ||
		target.Selector.Another || target.Selector.Other {
		return game.TargetSpec{}, false
	}
	var kinds []game.StackObjectKind
	allowsSpell := false
	switch target.Selector.Kind {
	case compiler.SelectorActivatedAbility:
		kinds = []game.StackObjectKind{game.StackActivatedAbility}
	case compiler.SelectorTriggeredAbility:
		kinds = []game.StackObjectKind{game.StackTriggeredAbility}
	case compiler.SelectorActivatedOrTriggeredAbility:
		kinds = []game.StackObjectKind{game.StackActivatedAbility, game.StackTriggeredAbility}
	case compiler.SelectorSpellActivatedOrTriggeredAbility:
		kinds = []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility}
		allowsSpell = true
	case compiler.SelectorTriggeredAbilityOrSpell:
		kinds = []game.StackObjectKind{game.StackTriggeredAbility, game.StackSpell}
		allowsSpell = true
	default:
		return game.TargetSpec{}, false
	}
	controller, ok := counterAbilityController(target.Selector.Controller)
	if !ok {
		return game.TargetSpec{}, false
	}
	predicate := game.TargetPredicate{
		StackObjectKinds:       kinds,
		Controller:             controller,
		StackObjectSourceTypes: append([]types.Card(nil), target.Selector.SourceTypes()...),
	}
	// Spell-shape qualifiers restrict only the spell choice in a mixed target;
	// they require that a spell is among the allowed kinds.
	supertypes := target.Selector.Supertypes()
	if (len(supertypes) > 0 || target.Selector.Colorless) && !allowsSpell {
		return game.TargetSpec{}, false
	}
	predicate.SpellSupertypes = append([]types.Super(nil), supertypes...)
	predicate.SpellColorless = target.Selector.Colorless
	return game.TargetSpec{
		MinTargets: 1,
		MaxTargets: 1,
		Constraint: lowerFirst(target.Text),
		Allow:      game.TargetAllowStackObject,
		Predicate:  predicate,
	}, true
}

func counterAbilityController(controller compiler.ControllerKind) (game.ControllerRelation, bool) {
	switch controller {
	case compiler.ControllerAny:
		return game.ControllerAny, true
	case compiler.ControllerYou:
		return game.ControllerYou, true
	case compiler.ControllerOpponent:
		return game.ControllerOpponent, true
	case compiler.ControllerNotYou:
		return game.ControllerNotYou, true
	default:
		return game.ControllerAny, false
	}
}

func counterTargetSpec(target compiler.CompiledTarget) (game.TargetSpec, bool) {
	if spec, ok := stackSpellTargetSpec(target); ok {
		return spec, true
	}
	return counterAbilityTargetSpec(target)
}
