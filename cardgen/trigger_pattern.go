package cardgen

import (
	"slices"

	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
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
	alongsideSelection, ok := lowerTriggerSelection(pattern.AttacksAlongsideSelection)
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
	castDuringTurn, ok := lowerTriggerCastDuringTurn(pattern.CastDuringTurn)
	if !ok {
		return game.TriggerPattern{}, false
	}
	result := game.TriggerPattern{
		Event:                                   event,
		UnionEvent:                              unionEvent,
		Controller:                              controller,
		CauseController:                         causeController,
		Source:                                  source,
		ExcludeSelf:                             pattern.ExcludeSelf,
		Player:                                  player,
		Subject:                                 subject,
		SubjectSelection:                        subjectSelection,
		SubjectSelectionOrSelf:                  pattern.SubjectSelectionOrSelf,
		DamageSourceSelectionOrSelf:             pattern.DamageSourceSelectionOrSelf,
		RelatedSubjectSelection:                 relatedSelection,
		CardSelection:                           cardSelection,
		DamageRecipient:                         damageRecipient,
		DamageRecipientIsSource:                 pattern.DamageRecipientIsSource,
		DamageRecipientTypes:                    damageRecipientTypes,
		DamageRecipientSelection:                damageSelection,
		DamageSourceSelection:                   damageSourceSelection,
		AttackRecipient:                         attackRecipient,
		AttackRecipientSelection:                attackSelection,
		Step:                                    step,
		StepPlayerSourceAttachedSelection:       stepAttachedSelection,
		StepPlayerIsSourceEnchantedPlayer:       pattern.StepPlayerIsSourceEnchantedPlayer,
		FirstUpkeepStepEachTurn:                 pattern.FirstUpkeepStepEachTurn,
		OneOrMore:                               pattern.OneOrMore,
		OneOrMorePerAttackTarget:                pattern.OneOrMorePerAttackTarget,
		OneOrMorePerDamagedPlayer:               pattern.OneOrMorePerDamagedPlayer,
		AttackAlone:                             pattern.AttackAlone,
		AttackWhileSaddled:                      pattern.AttackWhileSaddled,
		AttacksDifferentPlayerThanAnother:       pattern.AttacksDifferentPlayerThanAnother,
		AttackedPlayerIsSourceEnchantedPlayer:   pattern.AttackedPlayerIsSourceEnchantedPlayer,
		AttackerCountAtLeast:                    pattern.AttackerCountAtLeast,
		AttacksAlongsideSelection:               alongsideSelection,
		AttacksAlongsideCount:                   pattern.AttacksAlongsideCount,
		RequireKickerPaid:                       pattern.RequireKickerPaid,
		RequireHistoric:                         pattern.RequireHistoric,
		ExcludeManaAbility:                      pattern.ExcludeManaAbility,
		PlayerEventOrdinalThisTurn:              pattern.PlayerEventOrdinalThisTurn,
		ExcludeFirstDrawInDrawStep:              pattern.ExcludeFirstDrawInDrawStep,
		MatchSpellCopy:                          pattern.MatchSpellCopy,
		SelfWasCast:                             pattern.SelfWasCast,
		SpellTargetsSource:                      pattern.SpellTargetsSource,
		CastDuringTurn:                          castDuringTurn,
		RequireTappedForMana:                    pattern.TappedForMana,
		RequireProducedManaColor:                pattern.TappedForManaColor,
		RequireProducedManaColorFromEntryChoice: pattern.TappedForManaChosenColor,
		RequireManaProducedByLand:               pattern.ManaProducedByLand,
		ClassBecameLevel:                        pattern.ClassBecameLevel,
		DyingDamagedBySource:                    pattern.DyingDamagedBySource,
	}
	if pattern.DyingDamagedBySource && event != game.EventPermanentDied {
		return game.TriggerPattern{}, false
	}
	if pattern.PlaysExiledWithSource != (event == game.EventCardPlayedFromExile) {
		return game.TriggerPattern{}, false
	}
	if pattern.PlaysExiledWithSource {
		result.PlaysLinkedExileCard = selfExileLinkKey
	}
	if pattern.ClassBecameLevel > 0 && event != game.EventClassLevelGained {
		return game.TriggerPattern{}, false
	}
	if pattern.TappedForMana && event != game.EventPermanentTapped && event != game.EventManaProduced {
		return game.TriggerPattern{}, false
	}
	if pattern.ManaProducedByLand && event != game.EventManaProduced {
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
		spellSelection, ok := spellTargetSelection(targetSelection)
		if !ok {
			return game.TriggerPattern{}, false
		}
		result.SpellTargetAllow = game.TargetAllowPermanent
		result.SpellTargetPattern = opt.Val(spellSelection)
	}
	if castDuringTurn != game.TriggerTurnAny &&
		event != game.EventSpellCast &&
		event != game.EventCardDrawn &&
		event != game.EventLifeGained &&
		event != game.EventLifeLost {
		return game.TriggerPattern{}, false
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
	if pattern.MatchCounter {
		if !pattern.Counter.Valid() {
			return game.TriggerPattern{}, false
		}
		result.MatchCounterKind = true
		result.CounterKind = pattern.Counter
	} else if pattern.CounterThreshold != 0 {
		return game.TriggerPattern{}, false
	}
	result.CounterThreshold = pattern.CounterThreshold
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
	if pattern.MatchFromZone && pattern.ExcludeFromZone {
		return false
	}
	if pattern.MatchFromZone || pattern.ExcludeFromZone {
		fromZone, ok := lowerTriggerZone(pattern.FromZone)
		if !ok {
			return false
		}
		result.FromZone = fromZone
		result.MatchFromZone = pattern.MatchFromZone
		result.ExcludeFromZone = pattern.ExcludeFromZone
	} else if pattern.FromZone != compiler.TriggerZoneNone {
		return false
	}
	if len(pattern.FromZones) > 0 {
		if pattern.MatchFromZone || pattern.ExcludeFromZone || len(pattern.FromZones) < 2 {
			return false
		}
		zones := make([]zone.Type, 0, len(pattern.FromZones))
		for _, from := range pattern.FromZones {
			lowered, ok := lowerTriggerZone(from)
			if !ok {
				return false
			}
			if slices.Contains(zones, lowered) {
				return false
			}
			zones = append(zones, lowered)
		}
		result.FromZones = zones
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
	case compiler.TriggerState:
		return game.TriggerState, true
	default:
		return 0, false
	}
}

func lowerTriggerEvent(event compiler.TriggerEvent) (game.EventKind, bool) {
	switch event {
	case compiler.TriggerEventSpellCast:
		return game.EventSpellCast, true
	case compiler.TriggerEventPermanentEnteredBattlefield, compiler.TriggerEventDoorUnlocked:
		// A Room half's door unlocks as the half enters the battlefield from
		// being cast, so the runtime fires the door-unlock trigger off the same
		// permanent-entered-battlefield event as an ordinary self-enters trigger.
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
	case compiler.TriggerEventAttackerBecameUnblocked:
		return game.EventAttackerBecameUnblocked, true
	case compiler.TriggerEventTokenCreated:
		return game.EventTokenCreated, true
	case compiler.TriggerEventLibrarySearched:
		return game.EventLibrarySearched, true
	case compiler.TriggerEventClassBecameLevel:
		return game.EventClassLevelGained, true
	case compiler.TriggerEventCrimeCommitted:
		return game.EventCrimeCommitted, true
	case compiler.TriggerEventBecameMonarch:
		return game.EventBecameMonarch, true
	case compiler.TriggerEventCardPlayedFromExile:
		return game.EventCardPlayedFromExile, true
	case compiler.TriggerEventLandPlayed:
		return game.EventLandPlayed, true
	case compiler.TriggerEventManaProduced:
		return game.EventManaProduced, true
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
	case compiler.TriggerPlayerMonarch:
		return game.TriggerPlayerMonarch, true
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

func lowerTriggerCastDuringTurn(relation compiler.TriggerCastTurn) (game.TriggerTurnRelation, bool) {
	switch relation {
	case compiler.TriggerCastTurnAny:
		return game.TriggerTurnAny, true
	case compiler.TriggerCastTurnYours:
		return game.TriggerTurnYours, true
	case compiler.TriggerCastTurnNotYours:
		return game.TriggerTurnNotYours, true
	case compiler.TriggerCastTurnEventPlayer:
		return game.TriggerTurnEventPlayer, true
	default:
		return game.TriggerTurnAny, false
	}
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

// lowerTriggerSelection projects a trigger-subject filter onto the canonical
// game.Selection. It is a thin adapter over the shared SelectionForSelector
// projector: triggerSelectionSelector translates the shared-typed TriggerSelection
// fields into a compiler.CompiledSelector, SelectionForSelectorMasked maps that
// onto the runtime Selection, and the two dimensions the selector atoms cannot
// carry (the conjunctive required card-type nouns and the trigger controller
// relation) are applied explicitly afterward. Routing through the canonical
// projector lets triggered abilities inherit every Selection dimension
// automatically instead of maintaining a second hand-written projector.
func lowerTriggerSelection(selection compiler.TriggerSelection) (game.Selection, bool) {
	controller, ok := lowerTriggerSelectionController(selection.Controller)
	if !ok {
		return game.Selection{}, false
	}
	selector, mask, ok := triggerSelectionSelector(selection)
	if !ok {
		return game.Selection{}, false
	}
	result, ok := SelectionForSelectorMasked(selector, mask)
	if !ok {
		return game.Selection{}, false
	}
	// The trigger required-type nouns and controller relation are mapped
	// directly: the canonical selector atoms carry RequiredTypesAny but not the
	// conjunctive RequiredTypes set, and the trigger controller collapses
	// "opponent" onto NotYou, so both are applied after the shared projection.
	result.RequiredTypes = selection.RequiredTypes
	result.RequiredTypesAny = selection.RequiredTypesAny
	result.Controller = controller
	result.MatchModified = selection.Modified
	result.MatchCommander = selection.Commander
	result.MatchGoaded = selection.Goaded
	result.PowerAboveBase = selection.PowerAboveBase
	result.ManaValueLessThanSourcePower = selection.ManaValueLessThanSourcePower
	for i := range selection.AnyOf {
		alternative, ok := lowerTriggerSelection(selection.AnyOf[i])
		if !ok {
			return game.Selection{}, false
		}
		result.AnyOf = append(result.AnyOf, alternative)
	}
	return result, true
}

// triggerSelectionSelector translates a TriggerSelection's shared-typed filter
// dimensions into a compiler.CompiledSelector and the mask that reproduces the
// trigger projector's behavior. It fails closed on any tristate or combat-state
// value the per-enum helpers cannot translate. The required card-type nouns and
// controller relation are mapped by lowerTriggerSelection directly because the
// selector atoms cannot carry the conjunctive required set and the trigger
// controller collapses "opponent" onto NotYou.
func triggerSelectionSelector(selection compiler.TriggerSelection) (compiler.CompiledSelector, SelectionMask, bool) {
	subtypes, ok := lowerTriggerSubtypes(selection.SubtypesAny)
	if !ok {
		return compiler.CompiledSelector{}, SelectionMask{}, false
	}
	tapped, ok := lowerTriggerTriState(selection.Tapped)
	if !ok {
		return compiler.CompiledSelector{}, SelectionMask{}, false
	}
	combatState, ok := lowerTriggerCombatState(selection.CombatState)
	if !ok {
		return compiler.CompiledSelector{}, SelectionMask{}, false
	}

	selector := compiler.CompiledSelector{
		Kind:                   compiler.SelectorPermanent,
		Colorless:              selection.Colorless,
		Multicolored:           selection.Multicolored,
		NonToken:               selection.NonToken,
		TokenOnly:              selection.TokenOnly,
		Keyword:                selection.Keyword,
		ExcludedKeyword:        selection.ExcludedKeyword,
		SubtypeFromEntryChoice: selection.SubtypeFromEntryChoice,
		ColorFromEntryChoice:   selection.ColorFromEntryChoice,
		MatchAnyCounter:        selection.MatchAnyCounter,
		MatchCounter:           selection.MatchCounter,
		RequiredCounter:        selection.RequiredCounter,
	}

	switch tapped {
	case game.TriTrue:
		selector.Tapped = true
	case game.TriFalse:
		selector.Untapped = true
	default:
	}

	switch combatState {
	case game.CombatStateAttacking:
		selector.Attacking = true
	case game.CombatStateBlocking:
		selector.Blocking = true
	default:
	}

	if selection.Power.Op != compare.Any {
		selector.MatchPower = true
		selector.Power = selection.Power
	}
	if selection.Toughness.Op != compare.Any {
		selector.MatchToughness = true
		selector.Toughness = selection.Toughness
	}

	switch {
	case selection.MatchManaValue:
		if selection.ManaValue.Op != compare.Any {
			return compiler.CompiledSelector{}, SelectionMask{}, false
		}
		op := compare.GreaterOrEqual
		value := selection.ManaValueAtLeast
		if selection.ManaValueAtMost != 0 {
			op = compare.LessOrEqual
			value = selection.ManaValueAtMost
		}
		selector.MatchManaValue = true
		selector.ManaValue = compare.Int{Op: op, Value: value}
	case selection.ManaValueAtLeast != 0 || selection.ManaValueAtMost != 0:
		return compiler.CompiledSelector{}, SelectionMask{}, false
	case selection.ManaValue.Op != compare.Any:
		selector.MatchManaValue = true
		selector.ManaValue = selection.ManaValue
	default:
	}

	selector = selector.WithAtoms(compiler.CompiledSelectorAtoms{
		ExcludedTypes:  selection.ExcludedTypes,
		Supertypes:     selection.Supertypes,
		SubtypesAny:    subtypes,
		ColorsAny:      selection.ColorsAny,
		ExcludedColors: selection.ExcludedColors,
	})

	return selector, SelectionMask{}.Rejecting(DimRequiredName), true
}

// spellTargetSelection projects a lowered trigger Selection onto the canonical
// permanent/card Selection used by a spell-cast trigger's SpellTargetPattern. It
// fails closed on Selection features the original TargetPredicate-backed pattern
// could not express, so unsupported target relations remain unsupported rather
// than silently widening trigger coverage. The projection drops the same fields
// the former predicate round-trip discarded, preserving byte-for-byte behavior.
func spellTargetSelection(selection game.Selection) (game.Selection, bool) {
	if selection.ExcludedSubtype != "" ||
		selection.Colorless ||
		selection.Multicolored ||
		selection.NonToken ||
		selection.TokenOnly ||
		selection.MatchCounter ||
		selection.MatchAnyCounter ||
		selection.MatchModified ||
		selection.RequiredCounterCount.Exists ||
		selection.EnteredThisTurn ||
		selection.DealtDamageThisTurn ||
		selection.SubtypeChoice != game.SubtypeChoiceNone ||
		selection.ColorChoice != game.ColorChoiceNone {
		return game.Selection{}, false
	}
	projected := game.Selection{
		RequiredTypes:     selection.RequiredTypes,
		RequiredTypesAny:  selection.RequiredTypesAny,
		ExcludedTypes:     selection.ExcludedTypes,
		Supertypes:        selection.Supertypes,
		ExcludedSupertype: selection.ExcludedSupertype,
		SubtypesAny:       selection.SubtypesAny,
		ColorsAny:         selection.ColorsAny,
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
		ExcludeSource:     selection.ExcludeSource,
	}
	// A type-or-subtype disjunction ("one or more creatures or Vehicles you
	// control", Arcee, Acrobatic Coupe) rides Selection.AnyOf, which the flat
	// type/subtype fields cannot express. Each alternative is projected through
	// the same gate so an alternative with an unsupported feature still fails
	// closed rather than silently widening trigger coverage.
	for i := range selection.AnyOf {
		alternative, ok := spellTargetSelection(selection.AnyOf[i])
		if !ok {
			return game.Selection{}, false
		}
		projected.AnyOf = append(projected.AnyOf, alternative)
	}
	return projected, true
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

func lowerTriggerSubtypes(subtypes []types.Sub) ([]types.Sub, bool) {
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
