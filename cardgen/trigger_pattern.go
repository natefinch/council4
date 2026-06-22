package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
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
func lowerTriggerPattern(pattern *compiler.TriggerPattern) (game.TriggerPattern, bool) {
	if pattern.Event == compiler.TriggerEventAbilityActivated && !pattern.ExcludeManaAbility {
		return game.TriggerPattern{}, false
	}
	if pattern.NextOccurrence {
		return game.TriggerPattern{}, false
	}
	if pattern.DamageSourceIsStackObject {
		return game.TriggerPattern{}, false
	}
	event, ok := lowerTriggerEvent(pattern.Event)
	if !ok {
		return game.TriggerPattern{}, false
	}
	unionEvent := game.EventUnknown
	if pattern.UnionEvent != compiler.TriggerEventUnknown {
		unionEvent, ok = lowerTriggerEvent(pattern.UnionEvent)
		if !ok {
			return game.TriggerPattern{}, false
		}
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
	if pattern.DamageRecipient == compiler.TriggerDamageRecipientPermanent &&
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
		UnionEvent:                        unionEvent,
		Controller:                        controller,
		CauseController:                   causeController,
		Source:                            source,
		ExcludeSelf:                       pattern.ExcludeSelf,
		Player:                            player,
		Subject:                           subject,
		SubjectSelection:                  subjectSelection,
		SubjectSelectionOrSelf:            pattern.SubjectSelectionOrSelf,
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
		AttackAlone:                       pattern.AttackAlone,
		AttackWhileSaddled:                pattern.AttackWhileSaddled,
		AttackerCountAtLeast:              pattern.AttackerCountAtLeast,
		RequireKickerPaid:                 pattern.RequireKickerPaid,
		RequireHistoric:                   pattern.RequireHistoric,
		ExcludeManaAbility:                pattern.ExcludeManaAbility,
		PlayerEventOrdinalThisTurn:        pattern.PlayerEventOrdinalThisTurn,
		ExcludeFirstDrawInDrawStep:        pattern.ExcludeFirstDrawInDrawStep,
		MatchSpellCopy:                    pattern.MatchSpellCopy,
		SpellTargetsSource:                pattern.SpellTargetsSource,
		RequireTappedForMana:              pattern.TappedForMana,
		RequireProducedManaColor:          pattern.TappedForManaColor,
	}
	if pattern.TappedForMana && event != game.EventPermanentTapped {
		return game.TriggerPattern{}, false
	}
	if pattern.MatchSpellCopy && event != game.EventSpellCast {
		return game.TriggerPattern{}, false
	}
	if pattern.SpellTargetsSource && event != game.EventSpellCast {
		return game.TriggerPattern{}, false
	}
	if pattern.SpellTargetSelection != nil {
		if event != game.EventSpellCast {
			return game.TriggerPattern{}, false
		}
		targetSelection, ok := lowerTriggerSelection(*pattern.SpellTargetSelection)
		if !ok {
			return game.TriggerPattern{}, false
		}
		predicate, ok := selectionTargetPredicate(targetSelection)
		if !ok {
			return game.TriggerPattern{}, false
		}
		result.SpellTargetAllow = game.TargetAllowPermanent
		result.SpellTargetPattern = opt.Val(predicate)
	}

	switch pattern.CombatQualifier {
	case compiler.TriggerCombatDamage:
		result.RequireCombatDamage = true
	case compiler.TriggerNonCombatDamage:
		result.RequireNonCombatDamage = true
	case compiler.TriggerCombatAny:
	default:
		return game.TriggerPattern{}, false
	}
	if pattern.StackObject == compiler.TriggerStackObjectSpell {
		result.MatchStackObjectKind = true
		result.StackObjectKind = game.StackSpell
	} else if pattern.StackObject != compiler.TriggerStackObjectAny {
		return game.TriggerPattern{}, false
	}
	if pattern.Counter != compiler.TriggerCounterAny {
		result.MatchCounterKind = true
		switch pattern.Counter {
		case compiler.TriggerCounterPlusOnePlusOne:
			result.CounterKind = counter.PlusOnePlusOne
		case compiler.TriggerCounterMinusOneMinusOne:
			result.CounterKind = counter.MinusOneMinusOne
		default:
			return game.TriggerPattern{}, false
		}
	}
	if !lowerTriggerZones(pattern, &result) {
		return game.TriggerPattern{}, false
	}
	if pattern.FaceDown && !pattern.MatchFaceDown {
		return game.TriggerPattern{}, false
	}
	if !attackerCountRelationsLowerable(pattern, event) {
		return game.TriggerPattern{}, false
	}
	result.MatchFaceDown = pattern.MatchFaceDown
	result.FaceDown = pattern.FaceDown
	return result, true
}

// lowerTriggerZones lowers the from-zone and to-zone filters into result,
// reporting false (fail-closed) when a zone is unrepresentable or the from/to
// flags are inconsistent. A to-zone may be required or excluded, never both.
func lowerTriggerZones(pattern *compiler.TriggerPattern, result *game.TriggerPattern) bool {
	if pattern.MatchFromZone {
		fromZone, ok := lowerTriggerZone(pattern.FromZone)
		if !ok {
			return false
		}
		result.FromZone = fromZone
		result.MatchFromZone = true
	} else if pattern.FromZone != compiler.TriggerZoneNone {
		return false
	}
	if pattern.MatchToZone && pattern.ExcludeToZone {
		return false
	}
	if pattern.MatchToZone || pattern.ExcludeToZone {
		toZone, ok := lowerTriggerZone(pattern.ToZone)
		if !ok {
			return false
		}
		result.ToZone = toZone
		result.MatchToZone = !pattern.ExcludeToZone
		result.ExcludeToZone = pattern.ExcludeToZone
	} else if pattern.ToZone != compiler.TriggerZoneNone {
		return false
	}
	return true
}

// attackerCountRelationsLowerable reports whether the attacker-count combat
// relations (AttackAlone, AttackerCountAtLeast) are well-formed for lowering.
// Both relations only apply to attacker-declared events; "N or more" requires
// N >= 2 and must not also be "attacks alone". The count is satisfied either by
// the controller-scoped one-or-more batching ("you attack with N or more
// creatures") or by a self-source pattern ("this creature and at least N other
// creatures attack", Battalion), whose single declared attacker is the source.
func attackerCountRelationsLowerable(pattern *compiler.TriggerPattern, event game.EventKind) bool {
	if (pattern.AttackAlone || pattern.AttackerCountAtLeast != 0) &&
		event != game.EventAttackerDeclared {
		return false
	}
	if pattern.AttackerCountAtLeast != 0 &&
		(pattern.AttackerCountAtLeast < 2 || pattern.AttackAlone ||
			(!pattern.OneOrMore && pattern.Source != compiler.TriggerSourceSelf)) {
		return false
	}
	return true
}

func triggerSelectionIsRequiredTypesOnly(selection game.Selection) bool {
	requiredTypes := selection.RequiredTypes
	selection.RequiredTypes = nil
	return len(requiredTypes) > 0 && selection.Empty()
}

func lowerTriggerAttackRecipient(recipient compiler.TriggerAttackRecipient) (game.AttackRecipientKind, bool) {
	const known = compiler.TriggerAttackRecipientPlayer |
		compiler.TriggerAttackRecipientPlaneswalker |
		compiler.TriggerAttackRecipientBattle
	if recipient&^known != 0 {
		return game.AttackRecipientAny, false
	}
	result := game.AttackRecipientAny
	if recipient&compiler.TriggerAttackRecipientPlayer != 0 {
		result |= game.AttackRecipientPlayer
	}
	if recipient&compiler.TriggerAttackRecipientPlaneswalker != 0 {
		result |= game.AttackRecipientPlaneswalker
	}
	if recipient&compiler.TriggerAttackRecipientBattle != 0 {
		result |= game.AttackRecipientBattle
	}
	return result, true
}

func lowerTriggerKind(kind compiler.TriggerKind) (game.TriggerType, bool) {
	switch kind {
	case compiler.TriggerWhen:
		return game.TriggerWhen, true
	case compiler.TriggerWhenever:
		return game.TriggerWhenever, true
	case compiler.TriggerAt:
		return game.TriggerAt, true
	default:
		return 0, false
	}
}

func lowerTriggerEvent(event compiler.TriggerEvent) (game.EventKind, bool) {
	switch event {
	case compiler.TriggerEventSpellCast:
		return game.EventSpellCast, true
	case compiler.TriggerEventPermanentEnteredBattlefield:
		return game.EventPermanentEnteredBattlefield, true
	case compiler.TriggerEventPermanentDied:
		return game.EventPermanentDied, true
	case compiler.TriggerEventZoneChanged:
		return game.EventZoneChanged, true
	case compiler.TriggerEventCountersAdded:
		return game.EventCountersAdded, true
	case compiler.TriggerEventDamageDealt:
		return game.EventDamageDealt, true
	case compiler.TriggerEventCardDrawn:
		return game.EventCardDrawn, true
	case compiler.TriggerEventAttackerDeclared:
		return game.EventAttackerDeclared, true
	case compiler.TriggerEventBlockerDeclared:
		return game.EventBlockerDeclared, true
	case compiler.TriggerEventCardDiscarded:
		return game.EventCardDiscarded, true
	case compiler.TriggerEventCycled:
		return game.EventCycled, true
	case compiler.TriggerEventBeginningOfStep:
		return game.EventBeginningOfStep, true
	case compiler.TriggerEventLifeGained:
		return game.EventLifeGained, true
	case compiler.TriggerEventLifeLost:
		return game.EventLifeLost, true
	case compiler.TriggerEventPermanentTapped:
		return game.EventPermanentTapped, true
	case compiler.TriggerEventPermanentUntapped:
		return game.EventPermanentUntapped, true
	case compiler.TriggerEventPermanentTurnedFaceUp:
		return game.EventPermanentTurnedFaceUp, true
	case compiler.TriggerEventPermanentSacrificed:
		return game.EventPermanentSacrificed, true
	case compiler.TriggerEventScry:
		return game.EventScry, true
	case compiler.TriggerEventSurveil:
		return game.EventSurveil, true
	case compiler.TriggerEventAbilityActivated:
		return game.EventAbilityActivated, true
	case compiler.TriggerEventObjectBecameTarget:
		return game.EventObjectBecameTarget, true
	case compiler.TriggerEventPermanentMutated:
		return game.EventPermanentMutated, true
	case compiler.TriggerEventAttackerBecameBlocked:
		return game.EventAttackerBecameBlocked, true
	case compiler.TriggerEventTokenCreated:
		return game.EventTokenCreated, true
	case compiler.TriggerEventLibrarySearched:
		return game.EventLibrarySearched, true
	default:
		return game.EventUnknown, false
	}
}

func lowerTriggerController(controller compiler.ControllerKind) (game.TriggerControllerFilter, bool) {
	switch controller {
	case compiler.ControllerAny:
		return game.TriggerControllerAny, true
	case compiler.ControllerYou:
		return game.TriggerControllerYou, true
	case compiler.ControllerOpponent:
		return game.TriggerControllerOpponent, true
	default:
		return game.TriggerControllerAny, false
	}
}

func lowerTriggerPlayer(player compiler.TriggerPlayerRelation) (game.TriggerPlayerFilter, bool) {
	switch player {
	case compiler.TriggerPlayerAny:
		return game.TriggerPlayerAny, true
	case compiler.TriggerPlayerYou:
		return game.TriggerPlayerYou, true
	case compiler.TriggerPlayerOpponent:
		return game.TriggerPlayerOpponent, true
	default:
		return game.TriggerPlayerAny, false
	}
}

func lowerTriggerSource(source compiler.TriggerSourceRelation) (game.TriggerSourceFilter, bool) {
	switch source {
	case compiler.TriggerSourceAny:
		return game.TriggerSourceAny, true
	case compiler.TriggerSourceSelf:
		return game.TriggerSourceSelf, true
	case compiler.TriggerSourceAttachedPermanent:
		return game.TriggerSourceAttachedPermanent, true
	default:
		return game.TriggerSourceAny, false
	}
}

func lowerTriggerSubject(subject compiler.TriggerSubject) (game.TriggerSubjectObject, bool) {
	switch subject {
	case compiler.TriggerSubjectDefault:
		return game.TriggerSubjectDefault, true
	case compiler.TriggerSubjectPermanent:
		return game.TriggerSubjectPermanent, true
	case compiler.TriggerSubjectBlockedAttacker:
		return game.TriggerSubjectBlockedAttacker, true
	case compiler.TriggerSubjectDamageSource:
		return game.TriggerSubjectDamageSource, true
	default:
		return game.TriggerSubjectDefault, false
	}
}

func lowerTriggerDamageRecipient(recipient compiler.TriggerDamageRecipient) (game.DamageRecipientKind, bool) {
	const known = compiler.TriggerDamageRecipientPlayer | compiler.TriggerDamageRecipientPermanent
	if recipient&^known != 0 {
		return game.DamageRecipientNone, false
	}
	result := game.DamageRecipientNone
	if recipient&compiler.TriggerDamageRecipientPlayer != 0 {
		result |= game.DamageRecipientPlayer
	}
	if recipient&compiler.TriggerDamageRecipientPermanent != 0 {
		result |= game.DamageRecipientPermanent
	}
	return result, true
}

func lowerTriggerStep(step compiler.TriggerStep) (game.Step, bool) {
	switch step {
	case compiler.TriggerStepNone:
		return game.StepNone, true
	case compiler.TriggerStepUpkeep:
		return game.StepUpkeep, true
	case compiler.TriggerStepDraw:
		return game.StepDraw, true
	case compiler.TriggerStepBeginningOfCombat:
		return game.StepBeginningOfCombat, true
	case compiler.TriggerStepEndOfCombat:
		return game.StepEndOfCombat, true
	case compiler.TriggerStepEnd:
		return game.StepEnd, true
	case compiler.TriggerStepPrecombatMain:
		return game.StepPrecombatMain, true
	case compiler.TriggerStepPostcombatMain:
		return game.StepPostcombatMain, true
	default:
		return game.StepNone, false
	}
}

func lowerTriggerZone(triggerZone compiler.TriggerZone) (zone.Type, bool) {
	switch triggerZone {
	case compiler.TriggerZoneGraveyard:
		return zone.Graveyard, true
	case compiler.TriggerZoneBattlefield:
		return zone.Battlefield, true
	case compiler.TriggerZoneHand:
		return zone.Hand, true
	case compiler.TriggerZoneExile:
		return zone.Exile, true
	case compiler.TriggerZoneLibrary:
		return zone.Library, true
	case compiler.TriggerZoneStack:
		return zone.Stack, true
	case compiler.TriggerZoneCommand:
		return zone.Command, true
	default:
		return 0, false
	}
}

func lowerTriggerSelection(selection compiler.TriggerSelection) (game.Selection, bool) {
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
	if selection.SubtypeFromEntryChoice {
		result.SubtypeChoice = game.SubtypeChoiceSourceEntry
	}
	result.Controller, ok = lowerTriggerSelectionController(selection.Controller)
	if !ok {
		return game.Selection{}, false
	}
	if selection.MatchManaValue {
		if selection.ManaValue.Comparison != compiler.TriggerComparisonUnknown {
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

// selectionTargetPredicate converts a lowered Selection into the equivalent
// TargetPredicate used by a spell-cast trigger's SpellTargetPattern. It is the
// inverse of TargetPredicate.Selection() and fails closed on Selection features
// that a TargetPredicate cannot express, so unsupported target relations remain
// unsupported rather than silently dropping a filter.
func selectionTargetPredicate(selection game.Selection) (game.TargetPredicate, bool) {
	if len(selection.AnyOf) > 0 ||
		selection.ExcludedSubtype != "" ||
		selection.Colorless ||
		selection.Multicolored ||
		selection.NonToken ||
		selection.TokenOnly ||
		selection.MatchCounter ||
		selection.MatchAnyCounter ||
		selection.MatchModified ||
		selection.RequiredCounterCount.Exists ||
		selection.EnteredThisTurn ||
		selection.SubtypeChoice != game.SubtypeChoiceNone ||
		selection.ColorChoice != game.ColorChoiceNone {
		return game.TargetPredicate{}, false
	}
	predicate := game.TargetPredicate{
		ExcludedTypes:     selection.ExcludedTypes,
		Supertypes:        selection.Supertypes,
		ExcludedSupertype: selection.ExcludedSupertype,
		Subtypes:          selection.SubtypesAny,
		Colors:            selection.ColorsAny,
		ExcludedColors:    selection.ExcludedColors,
		Controller:        selection.Controller,
		Player:            selection.Player,
		Tapped:            selection.Tapped,
		CombatState:       selection.CombatState,
		Keyword:           selection.Keyword,
		ExcludedKeyword:   selection.ExcludedKeyword,
		ManaValue:         selection.ManaValue,
		Power:             selection.Power,
		Toughness:         selection.Toughness,
		Another:           selection.ExcludeSource,
	}
	if len(selection.RequiredTypes) > 0 {
		predicate.PermanentTypes = selection.RequiredTypes
		predicate.PermanentTypesConjunctive = true
	} else {
		predicate.PermanentTypes = selection.RequiredTypesAny
	}
	return predicate, true
}

func lowerTriggerSelectionController(controller compiler.ControllerKind) (game.ControllerRelation, bool) {
	switch controller {
	case compiler.ControllerAny:
		return game.ControllerAny, true
	case compiler.ControllerYou:
		return game.ControllerYou, true
	case compiler.ControllerOpponent, compiler.ControllerNotYou:
		return game.ControllerNotYou, true
	default:
		return game.ControllerAny, false
	}
}

func lowerTriggerCombatState(state compiler.TriggerCombatState) (game.CombatStateFilter, bool) {
	switch state {
	case compiler.TriggerCombatStateAny:
		return game.CombatStateAny, true
	case compiler.TriggerCombatStateAttacking:
		return game.CombatStateAttacking, true
	case compiler.TriggerCombatStateBlocking:
		return game.CombatStateBlocking, true
	default:
		return game.CombatStateAny, false
	}
}

func lowerTriggerSupertypes(supertypes []compiler.TriggerSupertype) ([]types.Super, bool) {
	if len(supertypes) == 0 {
		return nil, true
	}
	result := make([]types.Super, 0, len(supertypes))
	for _, supertype := range supertypes {
		switch supertype {
		case compiler.TriggerSupertypeLegendary:
			result = append(result, types.Legendary)
		case compiler.TriggerSupertypeSnow:
			result = append(result, types.Snow)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerTriggerCardTypes(cardTypes []compiler.TriggerCardType) ([]types.Card, bool) {
	if len(cardTypes) == 0 {
		return nil, true
	}
	result := make([]types.Card, 0, len(cardTypes))
	for _, cardType := range cardTypes {
		switch cardType {
		case compiler.TriggerCardTypeArtifact:
			result = append(result, types.Artifact)
		case compiler.TriggerCardTypeBattle:
			result = append(result, types.Battle)
		case compiler.TriggerCardTypeCreature:
			result = append(result, types.Creature)
		case compiler.TriggerCardTypeEnchantment:
			result = append(result, types.Enchantment)
		case compiler.TriggerCardTypeInstant:
			result = append(result, types.Instant)
		case compiler.TriggerCardTypeLand:
			result = append(result, types.Land)
		case compiler.TriggerCardTypePlaneswalker:
			result = append(result, types.Planeswalker)
		case compiler.TriggerCardTypeSorcery:
			result = append(result, types.Sorcery)
		default:
			return nil, false
		}
	}
	return result, true
}

func lowerTriggerSubtypes(subtypes []compiler.TriggerSubtype) ([]types.Sub, bool) {
	if len(subtypes) == 0 {
		return nil, true
	}
	result := make([]types.Sub, 0, len(subtypes))
	result = append(result, subtypes...)
	return result, true
}

func lowerTriggerTriState(state compiler.TriggerTriState) (game.TriState, bool) {
	switch state {
	case compiler.TriggerTriAny:
		return game.TriAny, true
	case compiler.TriggerTriTrue:
		return game.TriTrue, true
	case compiler.TriggerTriFalse:
		return game.TriFalse, true
	default:
		return game.TriAny, false
	}
}

func lowerTriggerKeyword(keyword compiler.TriggerKeyword) (game.Keyword, bool) {
	switch keyword {
	case compiler.TriggerKeywordUnknown:
		return game.KeywordNone, true
	case compiler.TriggerKeywordDefender:
		return game.Defender, true
	case compiler.TriggerKeywordFlash:
		return game.Flash, true
	case compiler.TriggerKeywordFlying:
		return game.Flying, true
	case compiler.TriggerKeywordHaste:
		return game.Haste, true
	default:
		return game.KeywordNone, false
	}
}

func lowerTriggerNumberFilter(filter compiler.TriggerNumberFilter) (opt.V[compare.Int], bool) {
	var op compare.Op
	switch filter.Comparison {
	case compiler.TriggerComparisonUnknown:
		if filter.Value != 0 {
			return opt.V[compare.Int]{}, false
		}
		return opt.V[compare.Int]{}, true
	case compiler.TriggerComparisonEqual:
		op = compare.Equal
	case compiler.TriggerComparisonAtMost:
		op = compare.LessOrEqual
	case compiler.TriggerComparisonAtLeast:
		op = compare.GreaterOrEqual
	default:
		return opt.V[compare.Int]{}, false
	}
	return opt.Val(compare.Int{Op: op, Value: filter.Value}), true
}

func lowerTriggerColors(colors []compiler.TriggerColor) ([]color.Color, bool) {
	if len(colors) == 0 {
		return nil, true
	}
	result := make([]color.Color, 0, len(colors))
	for _, triggerColor := range colors {
		switch triggerColor {
		case compiler.TriggerColorWhite:
			result = append(result, color.White)
		case compiler.TriggerColorBlue:
			result = append(result, color.Blue)
		case compiler.TriggerColorBlack:
			result = append(result, color.Black)
		case compiler.TriggerColorRed:
			result = append(result, color.Red)
		case compiler.TriggerColorGreen:
			result = append(result, color.Green)
		default:
			return nil, false
		}
	}
	return result, true
}
