package cardgen

import (
	"strings"

	"github.com/natefinch/council4/cardgen/oracle"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// lowerTriggerPattern is the only semantic TriggerPattern to runtime
// game.TriggerPattern lowering path.
func lowerTriggerPattern(pattern *oracle.TriggerPattern) (game.TriggerPattern, bool) {
	if pattern.Event == oracle.TriggerEventAbilityActivated && !pattern.ExcludeManaAbility {
		return game.TriggerPattern{}, false
	}
	event, ok := lowerTriggerEvent(pattern.Event)
	if !ok {
		return game.TriggerPattern{}, false
	}
	controller, ok := lowerTriggerController(pattern.Controller)
	if !ok {
		return game.TriggerPattern{}, false
	}
	causeController, ok := lowerTriggerController(pattern.CauseController)
	if !ok {
		return game.TriggerPattern{}, false
	}
	player, ok := lowerTriggerPlayer(pattern.Player)
	if !ok {
		return game.TriggerPattern{}, false
	}
	source, ok := lowerTriggerSource(pattern.Source)
	if !ok {
		return game.TriggerPattern{}, false
	}
	subject, ok := lowerTriggerSubject(pattern.Subject)
	if !ok {
		return game.TriggerPattern{}, false
	}
	subjectSelection, ok := lowerTriggerSelection(pattern.SubjectSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	relatedSelection, ok := lowerTriggerSelection(pattern.RelatedSubjectSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	cardSelection, ok := lowerTriggerSelection(pattern.CardSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	damageSelection, ok := lowerTriggerSelection(pattern.DamageRecipientSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	damageRecipientTypes := []types.Card(nil)
	if pattern.DamageRecipient == oracle.TriggerDamageRecipientPermanent &&
		triggerSelectionIsRequiredTypesOnly(damageSelection) {
		damageRecipientTypes = damageSelection.RequiredTypes
		damageSelection = game.Selection{}
	}
	damageSourceSelection, ok := lowerTriggerSelection(pattern.DamageSourceSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	attackSelection, ok := lowerTriggerSelection(pattern.AttackRecipientSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	stepAttachedSelection, ok := lowerTriggerSelection(pattern.StepPlayerSourceAttachedSelection)
	if !ok {
		return game.TriggerPattern{}, false
	}
	damageRecipient, ok := lowerTriggerDamageRecipient(pattern.DamageRecipient)
	if !ok {
		return game.TriggerPattern{}, false
	}
	attackRecipient, ok := lowerTriggerAttackRecipient(pattern.AttackRecipient)
	if !ok {
		return game.TriggerPattern{}, false
	}
	step, ok := lowerTriggerStep(pattern.Step)
	if !ok {
		return game.TriggerPattern{}, false
	}
	result := game.TriggerPattern{
		Event:                             event,
		Controller:                        controller,
		CauseController:                   causeController,
		Source:                            source,
		ExcludeSelf:                       pattern.ExcludeSelf,
		Player:                            player,
		Subject:                           subject,
		SubjectSelection:                  subjectSelection,
		RelatedSubjectSelection:           relatedSelection,
		CardSelection:                     cardSelection,
		DamageRecipient:                   damageRecipient,
		DamageRecipientIsSource:           pattern.DamageRecipientIsSource,
		DamageRecipientTypes:              damageRecipientTypes,
		DamageRecipientSelection:          damageSelection,
		DamageSourceSelection:             damageSourceSelection,
		AttackRecipient:                   attackRecipient,
		AttackRecipientSelection:          attackSelection,
		Step:                              step,
		StepPlayerSourceAttachedSelection: stepAttachedSelection,
		OneOrMore:                         pattern.OneOrMore,
		OneOrMorePerAttackTarget:          pattern.OneOrMorePerAttackTarget,
		RequireKickerPaid:                 pattern.RequireKickerPaid,
		RequireHistoric:                   pattern.RequireHistoric,
		ExcludeManaAbility:                pattern.ExcludeManaAbility,
		PlayerEventOrdinalThisTurn:        pattern.PlayerEventOrdinalThisTurn,
	}

	switch pattern.CombatQualifier {
	case oracle.TriggerCombatDamage:
		result.RequireCombatDamage = true
	case oracle.TriggerNonCombatDamage:
		result.RequireNonCombatDamage = true
	case oracle.TriggerCombatAny:
	default:
		return game.TriggerPattern{}, false
	}
	if pattern.StackObject == oracle.TriggerStackObjectSpell {
		result.MatchStackObjectKind = true
		result.StackObjectKind = game.StackSpell
	} else if pattern.StackObject != oracle.TriggerStackObjectAny {
		return game.TriggerPattern{}, false
	}
	if pattern.Counter != oracle.TriggerCounterAny {
		result.MatchCounterKind = true
		switch pattern.Counter {
		case oracle.TriggerCounterPlusOnePlusOne:
			result.CounterKind = counter.PlusOnePlusOne
		case oracle.TriggerCounterMinusOneMinusOne:
			result.CounterKind = counter.MinusOneMinusOne
		default:
			return game.TriggerPattern{}, false
		}
	}
	if pattern.MatchFromZone {
		result.FromZone, ok = lowerTriggerZone(pattern.FromZone)
		if !ok {
			return game.TriggerPattern{}, false
		}
		result.MatchFromZone = true
	} else if pattern.FromZone != oracle.TriggerZoneNone {
		return game.TriggerPattern{}, false
	}
	if pattern.MatchToZone && pattern.ExcludeToZone {
		return game.TriggerPattern{}, false
	}
	if pattern.MatchToZone || pattern.ExcludeToZone {
		result.ToZone, ok = lowerTriggerZone(pattern.ToZone)
		if !ok {
			return game.TriggerPattern{}, false
		}
		result.MatchToZone = true
		result.ExcludeToZone = pattern.ExcludeToZone
		if pattern.ExcludeToZone {
			result.MatchToZone = false
		}
	} else if pattern.ToZone != oracle.TriggerZoneNone {
		return game.TriggerPattern{}, false
	}
	if pattern.FaceDown && !pattern.MatchFaceDown {
		return game.TriggerPattern{}, false
	}
	result.MatchFaceDown = pattern.MatchFaceDown
	result.FaceDown = pattern.FaceDown
	return result, true
}

func triggerSelectionIsRequiredTypesOnly(selection game.Selection) bool {
	requiredTypes := selection.RequiredTypes
	selection.RequiredTypes = nil
	return len(requiredTypes) > 0 && selection.Empty()
}

func lowerTriggerAttackRecipient(recipient oracle.TriggerAttackRecipient) (game.AttackRecipientKind, bool) {
	const known = oracle.TriggerAttackRecipientPlayer |
		oracle.TriggerAttackRecipientPlaneswalker |
		oracle.TriggerAttackRecipientBattle
	if recipient&^known != 0 {
		return game.AttackRecipientAny, false
	}
	result := game.AttackRecipientAny
	if recipient&oracle.TriggerAttackRecipientPlayer != 0 {
		result |= game.AttackRecipientPlayer
	}
	if recipient&oracle.TriggerAttackRecipientPlaneswalker != 0 {
		result |= game.AttackRecipientPlaneswalker
	}
	if recipient&oracle.TriggerAttackRecipientBattle != 0 {
		result |= game.AttackRecipientBattle
	}
	return result, true
}

func lowerTriggerKind(kind oracle.TriggerKind) (game.TriggerType, bool) {
	switch kind {
	case oracle.TriggerWhen:
		return game.TriggerWhen, true
	case oracle.TriggerWhenever:
		return game.TriggerWhenever, true
	case oracle.TriggerAt:
		return game.TriggerAt, true
	default:
		return 0, false
	}
}

func lowerTriggerEvent(event oracle.TriggerEvent) (game.EventKind, bool) {
	switch event {
	case oracle.TriggerEventSpellCast:
		return game.EventSpellCast, true
	case oracle.TriggerEventPermanentEnteredBattlefield:
		return game.EventPermanentEnteredBattlefield, true
	case oracle.TriggerEventPermanentDied:
		return game.EventPermanentDied, true
	case oracle.TriggerEventZoneChanged:
		return game.EventZoneChanged, true
	case oracle.TriggerEventCountersAdded:
		return game.EventCountersAdded, true
	case oracle.TriggerEventDamageDealt:
		return game.EventDamageDealt, true
	case oracle.TriggerEventCardDrawn:
		return game.EventCardDrawn, true
	case oracle.TriggerEventAttackerDeclared:
		return game.EventAttackerDeclared, true
	case oracle.TriggerEventBlockerDeclared:
		return game.EventBlockerDeclared, true
	case oracle.TriggerEventCardDiscarded:
		return game.EventCardDiscarded, true
	case oracle.TriggerEventCycled:
		return game.EventCycled, true
	case oracle.TriggerEventBeginningOfStep:
		return game.EventBeginningOfStep, true
	case oracle.TriggerEventLifeGained:
		return game.EventLifeGained, true
	case oracle.TriggerEventLifeLost:
		return game.EventLifeLost, true
	case oracle.TriggerEventPermanentTapped:
		return game.EventPermanentTapped, true
	case oracle.TriggerEventPermanentUntapped:
		return game.EventPermanentUntapped, true
	case oracle.TriggerEventPermanentTurnedFaceUp:
		return game.EventPermanentTurnedFaceUp, true
	case oracle.TriggerEventPermanentSacrificed:
		return game.EventPermanentSacrificed, true
	case oracle.TriggerEventScry:
		return game.EventScry, true
	case oracle.TriggerEventSurveil:
		return game.EventSurveil, true
	case oracle.TriggerEventAbilityActivated:
		return game.EventAbilityActivated, true
	case oracle.TriggerEventObjectBecameTarget:
		return game.EventObjectBecameTarget, true
	case oracle.TriggerEventPermanentMutated:
		return game.EventPermanentMutated, true
	case oracle.TriggerEventAttackerBecameBlocked:
		return game.EventAttackerBecameBlocked, true
	default:
		return game.EventUnknown, false
	}
}

func lowerTriggerController(controller oracle.ControllerKind) (game.TriggerControllerFilter, bool) {
	switch controller {
	case oracle.ControllerAny:
		return game.TriggerControllerAny, true
	case oracle.ControllerYou:
		return game.TriggerControllerYou, true
	case oracle.ControllerOpponent:
		return game.TriggerControllerOpponent, true
	default:
		return game.TriggerControllerAny, false
	}
}

func lowerTriggerPlayer(player oracle.TriggerPlayerRelation) (game.TriggerPlayerFilter, bool) {
	switch player {
	case oracle.TriggerPlayerAny:
		return game.TriggerPlayerAny, true
	case oracle.TriggerPlayerYou:
		return game.TriggerPlayerYou, true
	case oracle.TriggerPlayerOpponent:
		return game.TriggerPlayerOpponent, true
	default:
		return game.TriggerPlayerAny, false
	}
}

func lowerTriggerSource(source oracle.TriggerSourceRelation) (game.TriggerSourceFilter, bool) {
	switch source {
	case oracle.TriggerSourceAny:
		return game.TriggerSourceAny, true
	case oracle.TriggerSourceSelf:
		return game.TriggerSourceSelf, true
	case oracle.TriggerSourceAttachedPermanent:
		return game.TriggerSourceAttachedPermanent, true
	default:
		return game.TriggerSourceAny, false
	}
}

func lowerTriggerSubject(subject oracle.TriggerSubject) (game.TriggerSubjectObject, bool) {
	switch subject {
	case oracle.TriggerSubjectDefault:
		return game.TriggerSubjectDefault, true
	case oracle.TriggerSubjectPermanent:
		return game.TriggerSubjectPermanent, true
	case oracle.TriggerSubjectBlockedAttacker:
		return game.TriggerSubjectBlockedAttacker, true
	case oracle.TriggerSubjectDamageSource:
		return game.TriggerSubjectDamageSource, true
	default:
		return game.TriggerSubjectDefault, false
	}
}

func lowerTriggerDamageRecipient(recipient oracle.TriggerDamageRecipient) (game.DamageRecipientKind, bool) {
	const known = oracle.TriggerDamageRecipientPlayer | oracle.TriggerDamageRecipientPermanent
	if recipient&^known != 0 {
		return game.DamageRecipientNone, false
	}
	result := game.DamageRecipientNone
	if recipient&oracle.TriggerDamageRecipientPlayer != 0 {
		result |= game.DamageRecipientPlayer
	}
	if recipient&oracle.TriggerDamageRecipientPermanent != 0 {
		result |= game.DamageRecipientPermanent
	}
	return result, true
}

func lowerTriggerStep(step oracle.TriggerStep) (game.Step, bool) {
	switch step {
	case oracle.TriggerStepNone:
		return game.StepNone, true
	case oracle.TriggerStepUpkeep:
		return game.StepUpkeep, true
	case oracle.TriggerStepDraw:
		return game.StepDraw, true
	case oracle.TriggerStepBeginningOfCombat:
		return game.StepBeginningOfCombat, true
	case oracle.TriggerStepEndOfCombat:
		return game.StepEndOfCombat, true
	case oracle.TriggerStepEnd:
		return game.StepEnd, true
	case oracle.TriggerStepPrecombatMain:
		return game.StepPrecombatMain, true
	case oracle.TriggerStepPostcombatMain:
		return game.StepPostcombatMain, true
	default:
		return game.StepNone, false
	}
}

func lowerTriggerZone(triggerZone oracle.TriggerZone) (zone.Type, bool) {
	switch triggerZone {
	case oracle.TriggerZoneGraveyard:
		return zone.Graveyard, true
	case oracle.TriggerZoneBattlefield:
		return zone.Battlefield, true
	case oracle.TriggerZoneHand:
		return zone.Hand, true
	case oracle.TriggerZoneExile:
		return zone.Exile, true
	case oracle.TriggerZoneLibrary:
		return zone.Library, true
	case oracle.TriggerZoneStack:
		return zone.Stack, true
	case oracle.TriggerZoneCommand:
		return zone.Command, true
	default:
		return 0, false
	}
}

func lowerTriggerSelection(selection oracle.TriggerSelection) (game.Selection, bool) {
	required, ok := lowerTriggerCardTypes(selection.RequiredTypes)
	if !ok {
		return game.Selection{}, false
	}
	requiredAny, ok := lowerTriggerCardTypes(selection.RequiredTypesAny)
	if !ok {
		return game.Selection{}, false
	}
	excluded, ok := lowerTriggerCardTypes(selection.ExcludedTypes)
	if !ok {
		return game.Selection{}, false
	}
	supertypes, ok := lowerTriggerSupertypes(selection.Supertypes)
	if !ok {
		return game.Selection{}, false
	}
	subtypes, ok := lowerTriggerSubtypes(selection.SubtypesAny)
	if !ok {
		return game.Selection{}, false
	}
	colors, ok := lowerTriggerColors(selection.ColorsAny)
	if !ok {
		return game.Selection{}, false
	}
	excludedColors, ok := lowerTriggerColors(selection.ExcludedColors)
	if !ok {
		return game.Selection{}, false
	}
	tapped, ok := lowerTriggerTriState(selection.Tapped)
	if !ok {
		return game.Selection{}, false
	}
	combatState, ok := lowerTriggerCombatState(selection.CombatState)
	if !ok {
		return game.Selection{}, false
	}
	keyword, ok := lowerTriggerKeyword(selection.Keyword)
	if !ok {
		return game.Selection{}, false
	}
	excludedKeyword, ok := lowerTriggerKeyword(selection.ExcludedKeyword)
	if !ok {
		return game.Selection{}, false
	}
	manaValue, ok := lowerTriggerNumberFilter(selection.ManaValue)
	if !ok {
		return game.Selection{}, false
	}
	power, ok := lowerTriggerNumberFilter(selection.Power)
	if !ok {
		return game.Selection{}, false
	}
	toughness, ok := lowerTriggerNumberFilter(selection.Toughness)
	if !ok {
		return game.Selection{}, false
	}
	result := game.Selection{
		RequiredTypes:    required,
		RequiredTypesAny: requiredAny,
		ExcludedTypes:    excluded,
		Supertypes:       supertypes,
		SubtypesAny:      subtypes,
		ColorsAny:        colors,
		ExcludedColors:   excludedColors,
		Colorless:        selection.Colorless,
		Multicolored:     selection.Multicolored,
		Tapped:           tapped,
		CombatState:      combatState,
		Keyword:          keyword,
		ExcludedKeyword:  excludedKeyword,
		ManaValue:        manaValue,
		Power:            power,
		Toughness:        toughness,
		NonToken:         selection.NonToken,
		TokenOnly:        selection.TokenOnly,
	}
	result.Controller, ok = lowerTriggerSelectionController(selection.Controller)
	if !ok {
		return game.Selection{}, false
	}
	if selection.MatchManaValue {
		if selection.ManaValue.Comparison != oracle.TriggerComparisonUnknown {
			return game.Selection{}, false
		}
		result.ManaValue = opt.Val(compare.Int{
			Op:    compare.GreaterOrEqual,
			Value: selection.ManaValueAtLeast,
		})
	} else if selection.ManaValueAtLeast != 0 {
		return game.Selection{}, false
	}
	return result, true
}

func lowerTriggerSelectionController(controller oracle.ControllerKind) (game.ControllerRelation, bool) {
	switch controller {
	case oracle.ControllerAny:
		return game.ControllerAny, true
	case oracle.ControllerYou:
		return game.ControllerYou, true
	case oracle.ControllerOpponent, oracle.ControllerNotYou:
		return game.ControllerNotYou, true
	default:
		return game.ControllerAny, false
	}
}

func lowerTriggerCombatState(state oracle.TriggerCombatState) (game.CombatStateFilter, bool) {
	switch state {
	case oracle.TriggerCombatStateAny:
		return game.CombatStateAny, true
	case oracle.TriggerCombatStateAttacking:
		return game.CombatStateAttacking, true
	case oracle.TriggerCombatStateBlocking:
		return game.CombatStateBlocking, true
	default:
		return game.CombatStateAny, false
	}
}

func lowerTriggerSupertypes(supertypes []oracle.TriggerSupertype) ([]types.Super, bool) {
	if len(supertypes) == 0 {
		return nil, true
	}
	result := make([]types.Super, 0, len(supertypes))
	for _, supertype := range supertypes {
		switch supertype {
		case oracle.TriggerSupertypeLegendary:
			result = append(result, types.Legendary)
		case oracle.TriggerSupertypeSnow:
			result = append(result, types.Snow)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerTriggerCardTypes(cardTypes []oracle.TriggerCardType) ([]types.Card, bool) {
	if len(cardTypes) == 0 {
		return nil, true
	}
	result := make([]types.Card, 0, len(cardTypes))
	for _, cardType := range cardTypes {
		switch cardType {
		case oracle.TriggerCardTypeArtifact:
			result = append(result, types.Artifact)
		case oracle.TriggerCardTypeBattle:
			result = append(result, types.Battle)
		case oracle.TriggerCardTypeCreature:
			result = append(result, types.Creature)
		case oracle.TriggerCardTypeEnchantment:
			result = append(result, types.Enchantment)
		case oracle.TriggerCardTypeInstant:
			result = append(result, types.Instant)
		case oracle.TriggerCardTypeLand:
			result = append(result, types.Land)
		case oracle.TriggerCardTypePlaneswalker:
			result = append(result, types.Planeswalker)
		case oracle.TriggerCardTypeSorcery:
			result = append(result, types.Sorcery)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerTriggerSubtypes(subtypes []oracle.TriggerSubtype) ([]types.Sub, bool) {
	if len(subtypes) == 0 {
		return nil, true
	}
	result := make([]types.Sub, 0, len(subtypes))
	for _, subtype := range subtypes {
		runtimeSubtype := types.Sub(canonicalTriggerSubtype(string(subtype)))
		if !knownTriggerSubtype(runtimeSubtype) {
			return nil, false
		}
		result = append(result, runtimeSubtype)
	}
	return result, true
}

func canonicalTriggerSubtype(value string) string {
	var result strings.Builder
	result.Grow(len(value))
	uppercaseNext := true
	for _, r := range value {
		if uppercaseNext && r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
		}
		_, _ = result.WriteRune(r)
		uppercaseNext = r == ' ' || r == '-'
	}
	return result.String()
}

func knownTriggerSubtype(subtype types.Sub) bool {
	for _, cardType := range []types.Card{
		types.Artifact,
		types.Battle,
		types.Creature,
		types.Enchantment,
		types.Instant,
		types.Land,
		types.Planeswalker,
		types.Sorcery,
	} {
		if types.KnownSubtypeForType(cardType, subtype) {
			return true
		}
	}
	return false
}

func lowerTriggerTriState(state oracle.TriggerTriState) (game.TriState, bool) {
	switch state {
	case oracle.TriggerTriAny:
		return game.TriAny, true
	case oracle.TriggerTriTrue:
		return game.TriTrue, true
	case oracle.TriggerTriFalse:
		return game.TriFalse, true
	default:
		return game.TriAny, false
	}
}

func lowerTriggerKeyword(keyword oracle.TriggerKeyword) (game.Keyword, bool) {
	switch keyword {
	case oracle.TriggerKeywordUnknown:
		return game.KeywordNone, true
	case oracle.TriggerKeywordDefender:
		return game.Defender, true
	case oracle.TriggerKeywordFlash:
		return game.Flash, true
	case oracle.TriggerKeywordFlying:
		return game.Flying, true
	case oracle.TriggerKeywordHaste:
		return game.Haste, true
	default:
		return game.KeywordNone, false
	}
}

func lowerTriggerNumberFilter(filter oracle.TriggerNumberFilter) (opt.V[compare.Int], bool) {
	var op compare.Op
	switch filter.Comparison {
	case oracle.TriggerComparisonUnknown:
		if filter.Value != 0 {
			return opt.V[compare.Int]{}, false
		}
		return opt.V[compare.Int]{}, true
	case oracle.TriggerComparisonEqual:
		op = compare.Equal
	case oracle.TriggerComparisonAtMost:
		op = compare.LessOrEqual
	case oracle.TriggerComparisonAtLeast:
		op = compare.GreaterOrEqual
	default:
		return opt.V[compare.Int]{}, false
	}
	return opt.Val(compare.Int{Op: op, Value: filter.Value}), true
}

func lowerTriggerColors(colors []oracle.TriggerColor) ([]color.Color, bool) {
	if len(colors) == 0 {
		return nil, true
	}
	result := make([]color.Color, 0, len(colors))
	for _, triggerColor := range colors {
		switch triggerColor {
		case oracle.TriggerColorWhite:
			result = append(result, color.White)
		case oracle.TriggerColorBlue:
			result = append(result, color.Blue)
		case oracle.TriggerColorBlack:
			result = append(result, color.Black)
		case oracle.TriggerColorRed:
			result = append(result, color.Red)
		case oracle.TriggerColorGreen:
			result = append(result, color.Green)
		default:
			return nil, false
		}
	}
	return result, true
}
