package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func compileTriggerEventPattern(
	clause *parser.TriggerEventClause,
	kind TriggerKind,
	condition *CompiledCondition,
) TriggerPattern {
	pattern := TriggerPattern{
		Span:                 clause.Span,
		Kind:                 kind,
		InterveningCondition: condition,
	}
	if kind != TriggerWhen && kind != TriggerWhenever {
		return pattern
	}
	compiled, ok := compileTriggerEventClause(clause)
	if !ok {
		return pattern
	}
	compiled.Span = clause.Span
	compiled.Kind = kind
	compiled.InterveningCondition = condition
	return compiled
}

func compileTriggerEventClause(clause *parser.TriggerEventClause) (TriggerPattern, bool) {
	pattern := TriggerPattern{
		ExcludeSelf:               clause.ExcludeSelf,
		OneOrMore:                 clause.OneOrMore,
		OneOrMorePerAttackTarget:  clause.OneOrMorePerAttackTarget,
		ExcludeManaAbility:        clause.ExcludeManaAbility,
		DamageSourceIsStackObject: clause.DamageSourceIsStackObject,
		MatchFaceDown:             clause.FaceDown,
		FaceDown:                  clause.FaceDown,
		TappedForMana:             clause.TappedForMana,
		TappedForManaColor:        clause.TappedForManaColor,
	}
	var ok bool
	pattern.Controller, ok = compileTriggerController(clause.Controller)
	if !ok {
		return TriggerPattern{}, false
	}
	pattern.Player, ok = compileOptionalTriggerPlayer(&clause.Player)
	if !ok {
		return TriggerPattern{}, false
	}
	switch clause.Kind {
	case parser.TriggerEventKindZoneChange:
		ok = compileZoneChangeEvent(clause, &pattern)
	case parser.TriggerEventKindSpellCast:
		ok = compileSpellCastEvent(clause, &pattern)
	case parser.TriggerEventKindAbilityActivated:
		ok = compileAbilityActivatedEvent(clause, &pattern)
	case parser.TriggerEventKindAttack:
		ok = compileAttackEvent(clause, &pattern)
	case parser.TriggerEventKindBlock:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventBlockerDeclared)
	case parser.TriggerEventKindBecameBlocked:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventAttackerBecameBlocked)
	case parser.TriggerEventKindDamageDealt:
		ok = compileDamageEvent(clause, &pattern)
	case parser.TriggerEventKindCounterAdded:
		ok = compileCounterEvent(clause, &pattern)
	case parser.TriggerEventKindBecomesTapped:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventPermanentTapped)
	case parser.TriggerEventKindBecomesUntapped:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventPermanentUntapped)
	case parser.TriggerEventKindTurnedFaceUp:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventPermanentTurnedFaceUp)
	case parser.TriggerEventKindSacrificed:
		ok = compileSacrificeEvent(clause, &pattern)
	case parser.TriggerEventKindMutated:
		ok = compilePermanentSubjectEvent(clause, &pattern, TriggerEventPermanentMutated)
	case parser.TriggerEventKindBecameTarget:
		ok = compileBecameTargetEvent(clause, &pattern)
	case parser.TriggerEventKindTokenCreated:
		ok = compileTokenCreatedEvent(clause, &pattern)
	default:
		return TriggerPattern{}, false
	}
	if !ok {
		return TriggerPattern{}, false
	}
	if clause.UnionKind != parser.TriggerEventKindUnknown {
		unionEvent, unionOK := compileUnionTriggerEvent(clause.UnionKind)
		if !unionOK {
			return TriggerPattern{}, false
		}
		pattern.UnionEvent = unionEvent
	}
	return pattern, true
}

// compileUnionTriggerEvent maps a union secondary event family to its trigger
// event. The union shares the primary clause's subject and player filters, so
// only the bare event identity is needed here.
func compileUnionTriggerEvent(kind parser.TriggerEventKind) (TriggerEvent, bool) {
	switch kind {
	case parser.TriggerEventKindTokenCreated:
		return TriggerEventTokenCreated, true
	case parser.TriggerEventKindSacrificed:
		return TriggerEventPermanentSacrificed, true
	case parser.TriggerEventKindAttack:
		return TriggerEventAttackerDeclared, true
	case parser.TriggerEventKindBlock:
		return TriggerEventBlockerDeclared, true
	case parser.TriggerEventKindBecameBlocked:
		return TriggerEventAttackerBecameBlocked, true
	case parser.TriggerEventKindDied:
		return TriggerEventPermanentDied, true
	default:
		return TriggerEventUnknown, false
	}
}

func compileZoneChangeEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if !compileEventSubject(&clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	if clause.SelfOrAnother {
		// "this permanent or another <Selection> you control" requires a
		// selection subject (not self/attached) and no source restriction.
		if clause.Subject.Kind != parser.TriggerEventSubjectSelection ||
			pattern.Source != TriggerSourceAny ||
			pattern.ExcludeSelf {
			return false
		}
		pattern.SubjectSelectionOrSelf = true
	}
	switch clause.ZoneChange.Kind {
	case parser.TriggerEventZoneChangeEnteredBattlefield:
		pattern.Event = TriggerEventPermanentEnteredBattlefield
	case parser.TriggerEventZoneChangeDied:
		pattern.Event = TriggerEventPermanentDied
	case parser.TriggerEventZoneChangeMoved:
		pattern.Event = TriggerEventZoneChanged
	default:
		return false
	}
	if !compileZoneChangeZones(clause, pattern) {
		return false
	}
	switch clause.Tapped.Kind {
	case parser.TriggerEventTappedStateAny:
	case parser.TriggerEventTappedStateTapped:
		pattern.SubjectSelection.Tapped = TriggerTriTrue
	case parser.TriggerEventTappedStateUntapped:
		pattern.SubjectSelection.Tapped = TriggerTriFalse
	default:
		return false
	}
	return true
}

func compileZoneChangeZones(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if pattern.Event == TriggerEventPermanentDied {
		return true
	}
	if clause.Zone.MatchFromZone {
		pattern.FromZone, _ = compileTriggerEventZone(clause.Zone.FromZone.Kind)
		if pattern.FromZone == TriggerZoneNone {
			return false
		}
		pattern.MatchFromZone = true
	}
	if pattern.Event != TriggerEventZoneChanged {
		return true
	}
	pattern.MatchToZone = clause.Zone.MatchToZone
	pattern.ExcludeToZone = clause.Zone.ExcludeToZone
	if !pattern.MatchToZone && !pattern.ExcludeToZone {
		return true
	}
	pattern.ToZone, _ = compileTriggerEventZone(clause.Zone.ToZone.Kind)
	return pattern.ToZone != TriggerZoneNone &&
		(!pattern.MatchToZone || !pattern.ExcludeToZone)
}

func compileSpellCastEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if clause.Actor.Kind == parser.TriggerEventActorUnknown {
		return false
	}
	controller, ok := compileTriggerActorController(clause.Actor.Kind)
	if !ok {
		return false
	}
	selection, ok := compileTriggerSpellSelection(clause.SpellSelection)
	if !ok {
		return false
	}
	if clause.MatchCopy && controller != ControllerYou {
		return false
	}
	pattern.Event = TriggerEventSpellCast
	pattern.Controller = controller
	pattern.CardSelection = selection
	pattern.MatchSpellCopy = clause.MatchCopy
	pattern.SpellTargetsSource = clause.SpellTargetsSource
	if clause.SpellTargetSelection != nil {
		targetSelection, selectionOK := compileTriggerSelection(*clause.SpellTargetSelection)
		if !selectionOK {
			return false
		}
		pattern.SpellTargetSelection = &targetSelection
	}
	turn, ok := compileSpellCastTurnRelation(clause.SpellCastTurnRelation)
	if !ok {
		return false
	}
	pattern.CastDuringTurn = turn
	pattern.RequireKickerPaid = clause.SpellSelection.Kicker
	pattern.RequireHistoric = clause.SpellSelection.Historic
	if clause.SpellSelection.Ordinal != 0 {
		if clause.SpellSelection.Ordinal < 1 ||
			clause.SpellSelection.Ordinal > 5 ||
			clause.MatchCopy {
			return false
		}
		pattern.PlayerEventOrdinalThisTurn = clause.SpellSelection.Ordinal
	}
	if clause.SpellSelection.FromZone.Kind != parser.TriggerEventZoneNone {
		pattern.FromZone, ok = compileTriggerEventZone(clause.SpellSelection.FromZone.Kind)
		if !ok || controller != ControllerYou {
			return false
		}
		pattern.MatchFromZone = true
	}
	return true
}

func compileSpellCastTurnRelation(relation parser.TriggerCastTurnRelation) (TriggerCastTurn, bool) {
	switch relation {
	case parser.TriggerCastTurnRelationNone:
		return TriggerCastTurnAny, true
	case parser.TriggerCastTurnRelationYourTurn:
		return TriggerCastTurnYours, true
	case parser.TriggerCastTurnRelationNotYourTurn:
		return TriggerCastTurnNotYours, true
	default:
		return TriggerCastTurnAny, false
	}
}

func compileAbilityActivatedEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	player, ok := compileTriggerActorPlayer(clause.Actor.Kind)
	if !ok || !clause.ExcludeManaAbility {
		return false
	}
	selection, ok := compileTriggerSelection(clause.SourceSelection)
	if !ok || selection.Controller != ControllerAny {
		return false
	}
	pattern.Event = TriggerEventAbilityActivated
	pattern.Player = player
	pattern.SubjectSelection = selection
	return true
}

func compileAttackEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if clause.Actor.Kind != parser.TriggerEventActorUnknown {
		controller, ok := compileTriggerActorController(clause.Actor.Kind)
		if !ok {
			return false
		}
		pattern.Controller = controller
	}
	if clause.Subject.Kind != parser.TriggerEventSubjectUnknown &&
		!compileEventSubject(&clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	recipient, ok := compileTriggerAttackRecipient(clause.AttackRecipient.Kind)
	if !ok {
		return false
	}
	selection, ok := compileTriggerSelection(clause.AttackRecipient.Selection)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventAttackerDeclared
	pattern.AttackRecipient = recipient
	pattern.AttackRecipientSelection = selection
	pattern.AttackAlone = clause.AttackAlone
	pattern.AttackWhileSaddled = clause.AttackWhileSaddled
	pattern.AttackerCountAtLeast = clause.AttackerCountAtLeast
	return true
}

func compilePermanentSubjectEvent(
	clause *parser.TriggerEventClause,
	pattern *TriggerPattern,
	event TriggerEvent,
) bool {
	if !compileEventSubject(&clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	related, ok := compileTriggerSelection(clause.RelatedSelection)
	if !ok {
		return false
	}
	pattern.Event = event
	if event != TriggerEventAttackerBecameBlocked {
		pattern.RelatedSubjectSelection = related
	}
	return true
}

func compileDamageEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	qualifier, ok := compileTriggerCombatQualifier(clause.CombatQualifier.Kind)
	if !ok {
		return false
	}
	recipient, ok := compileTriggerDamageRecipient(clause.DamageRecipient.Kind)
	if !ok {
		return false
	}
	recipientSelection, ok := compileTriggerSelection(clause.DamageRecipient.Selection)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventDamageDealt
	pattern.CombatQualifier = qualifier
	pattern.DamageRecipient = recipient
	pattern.DamageRecipientSelection = recipientSelection
	pattern.DamageRecipientIsSource = clause.DamageRecipient.IsSource
	if clause.StackObject.Kind != parser.TriggerEventStackObjectAny {
		pattern.StackObject, ok = compileTriggerStackObject(clause.StackObject.Kind)
		if !ok {
			return false
		}
	}
	if clause.DamageSource.Kind != parser.TriggerEventSubjectUnknown {
		pattern.Subject = TriggerSubjectDamageSource
		return compileDamageSourceSubject(&clause.DamageSource, pattern)
	}
	if clause.DamageSourceIsStackObject {
		if clause.DamageSourceSpellSelection.Kicker ||
			clause.DamageSourceSpellSelection.Historic ||
			clause.DamageSourceSpellSelection.FromZone.Kind != parser.TriggerEventZoneNone {
			return false
		}
		pattern.Subject = TriggerSubjectDamageSource
		selection, selectionOK := compileTriggerSpellSelection(clause.DamageSourceSpellSelection)
		if !selectionOK {
			return false
		}
		pattern.DamageSourceSelection = selection
		return pattern.StackObject == TriggerStackObjectSpell
	}
	if clause.Subject.Kind == parser.TriggerEventSubjectUnknown {
		return recipient == TriggerDamageRecipientPlayer
	}
	if clause.Subject.Kind == parser.TriggerEventSubjectSelf {
		pattern.Subject = TriggerSubjectPermanent
	}
	return compileEventSubject(&clause.Subject, pattern, &pattern.SubjectSelection)
}

func compileCounterEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	counterValue, ok := compileTriggerCounter(clause.Counter.Kind)
	if !ok {
		return false
	}
	causeController, ok := compileTriggerActorController(clause.CauseController)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventCountersAdded
	pattern.Counter = counterValue
	pattern.CauseController = causeController
	switch clause.Subject.Kind {
	case parser.TriggerEventSubjectSelf:
		pattern.Source = TriggerSourceSelf
		return true
	case parser.TriggerEventSubjectSelection:
		return compileEventSubject(&clause.Subject, pattern, &pattern.SubjectSelection)
	default:
		return false
	}
}

func compileSacrificeEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	player, ok := compileTriggerActorPlayer(clause.Actor.Kind)
	if !ok || !compileEventSubject(&clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	pattern.Event = TriggerEventPermanentSacrificed
	pattern.Player = player
	return true
}

func compileTokenCreatedEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	player, ok := compileTriggerActorPlayer(clause.Actor.Kind)
	if !ok || !compileEventSubject(&clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	pattern.Event = TriggerEventTokenCreated
	pattern.Player = player
	return true
}

func compileBecameTargetEvent(clause *parser.TriggerEventClause, pattern *TriggerPattern) bool {
	if !compileEventSubject(&clause.Subject, pattern, &pattern.SubjectSelection) {
		return false
	}
	stackObject, ok := compileTriggerStackObject(clause.StackObject.Kind)
	if !ok {
		return false
	}
	causeController, ok := compileTriggerActorController(clause.CauseController)
	if !ok {
		return false
	}
	pattern.Event = TriggerEventObjectBecameTarget
	pattern.StackObject = stackObject
	pattern.CauseController = causeController
	return true
}

func compileEventSubject(
	subject *parser.TriggerEventSubject,
	pattern *TriggerPattern,
	destination *TriggerSelection,
) bool {
	selection, ok := compileTriggerSelection(subject.Selection)
	if !ok {
		return false
	}
	switch subject.Kind {
	case parser.TriggerEventSubjectSelf:
		pattern.Source = TriggerSourceSelf
		*destination = selection
	case parser.TriggerEventSubjectAttached:
		pattern.Source = TriggerSourceAttachedPermanent
		*destination = selection
	case parser.TriggerEventSubjectSelection:
		*destination = selection
	case parser.TriggerEventSubjectDamageSource:
	default:
		return false
	}
	return true
}

func compileDamageSourceSubject(subject *parser.TriggerEventSubject, pattern *TriggerPattern) bool {
	selection, ok := compileTriggerSelection(subject.Selection)
	if !ok {
		return false
	}
	switch subject.Kind {
	case parser.TriggerEventSubjectSelf:
		pattern.Source = TriggerSourceSelf
	case parser.TriggerEventSubjectAttached:
		pattern.Source = TriggerSourceAttachedPermanent
		pattern.DamageSourceSelection = selection
	case parser.TriggerEventSubjectSelection:
		pattern.DamageSourceSelection = selection
	case parser.TriggerEventSubjectDamageSource:
	default:
		return false
	}
	return true
}

func compileTriggerSpellSelection(syntax parser.TriggerEventSpellSelection) (TriggerSelection, bool) {
	selection := TriggerSelection{
		Colorless:        syntax.Colorless,
		Multicolored:     syntax.Multicolored,
		ManaValueAtLeast: syntax.ManaValueAtLeast,
		MatchManaValue:   syntax.MatchManaValue,
	}
	for _, value := range syntax.Types {
		compiled := compileTriggerCardType(value)
		if compiled == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		selection.RequiredTypes = append(selection.RequiredTypes, compiled)
	}
	for _, value := range syntax.TypesAny {
		compiled := compileTriggerCardType(value)
		if compiled == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		selection.RequiredTypesAny = append(selection.RequiredTypesAny, compiled)
	}
	for _, value := range syntax.ExcludedTypes {
		compiled := compileTriggerCardType(value)
		if compiled == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		selection.ExcludedTypes = append(selection.ExcludedTypes, compiled)
	}
	for _, value := range syntax.ColorsAny {
		compiled := compileTriggerColor(value)
		if compiled == TriggerColorUnknown {
			return TriggerSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, compiled)
	}
	if len(syntax.SubtypesAny) > 0 {
		selection.SubtypesAny = append(selection.SubtypesAny, syntax.SubtypesAny...)
	}
	selection.SubtypeFromEntryChoice = syntax.SubtypeFromEntryChoice
	return selection, true
}

func compileOptionalTriggerPlayer(player *parser.TriggerPlayerSelector) (TriggerPlayerRelation, bool) {
	if player.Kind == parser.TriggerPlayerSelectorUnknown {
		return TriggerPlayerAny, true
	}
	return compilePlayerEventPlayer(player)
}

func compileTriggerActorPlayer(actor parser.TriggerEventActorKind) (TriggerPlayerRelation, bool) {
	switch actor {
	case parser.TriggerEventActorYou:
		return TriggerPlayerYou, true
	case parser.TriggerEventActorOpponent:
		return TriggerPlayerOpponent, true
	case parser.TriggerEventActorPlayer:
		return TriggerPlayerAny, true
	default:
		return TriggerPlayerAny, false
	}
}

func compileTriggerActorController(actor parser.TriggerEventActorKind) (ControllerKind, bool) {
	switch actor {
	case parser.TriggerEventActorUnknown, parser.TriggerEventActorPlayer:
		return ControllerAny, true
	case parser.TriggerEventActorYou:
		return ControllerYou, true
	case parser.TriggerEventActorOpponent:
		return ControllerOpponent, true
	default:
		return ControllerAny, false
	}
}

func compileTriggerEventZone(value parser.TriggerEventZoneKind) (TriggerZone, bool) {
	switch value {
	case parser.TriggerEventZoneNone:
		return TriggerZoneNone, true
	case parser.TriggerEventZoneGraveyard:
		return TriggerZoneGraveyard, true
	case parser.TriggerEventZoneBattlefield:
		return TriggerZoneBattlefield, true
	case parser.TriggerEventZoneHand:
		return TriggerZoneHand, true
	case parser.TriggerEventZoneExile:
		return TriggerZoneExile, true
	case parser.TriggerEventZoneLibrary:
		return TriggerZoneLibrary, true
	case parser.TriggerEventZoneStack:
		return TriggerZoneStack, true
	case parser.TriggerEventZoneCommand:
		return TriggerZoneCommand, true
	default:
		return TriggerZoneNone, false
	}
}

func compileTriggerCombatQualifier(value parser.TriggerEventCombatQualifierKind) (TriggerCombatQualifier, bool) {
	switch value {
	case parser.TriggerEventCombatQualifierAny:
		return TriggerCombatAny, true
	case parser.TriggerEventCombatQualifierCombat:
		return TriggerCombatDamage, true
	case parser.TriggerEventCombatQualifierNoncombat:
		return TriggerNonCombatDamage, true
	default:
		return TriggerCombatAny, false
	}
}

func compileTriggerDamageRecipient(value parser.TriggerEventDamageRecipientKind) (TriggerDamageRecipient, bool) {
	const known = parser.TriggerEventDamageRecipientPlayer | parser.TriggerEventDamageRecipientPermanent
	if value&^known != 0 {
		return TriggerDamageRecipientAny, false
	}
	return TriggerDamageRecipient(value), true
}

func compileTriggerAttackRecipient(value parser.TriggerEventAttackRecipientKind) (TriggerAttackRecipient, bool) {
	const known = parser.TriggerEventAttackRecipientPlayer |
		parser.TriggerEventAttackRecipientPlaneswalker |
		parser.TriggerEventAttackRecipientBattle
	if value&^known != 0 {
		return TriggerAttackRecipientAny, false
	}
	return TriggerAttackRecipient(value), true
}

func compileTriggerStackObject(value parser.TriggerEventStackObjectKind) (TriggerStackObject, bool) {
	switch value {
	case parser.TriggerEventStackObjectAny:
		return TriggerStackObjectAny, true
	case parser.TriggerEventStackObjectSpell:
		return TriggerStackObjectSpell, true
	default:
		return TriggerStackObjectAny, false
	}
}

func compileTriggerCounter(value parser.TriggerEventCounterKind) (TriggerCounter, bool) {
	switch value {
	case parser.TriggerEventCounterPlusOnePlusOne:
		return TriggerCounterPlusOnePlusOne, true
	case parser.TriggerEventCounterMinusOneMinusOne:
		return TriggerCounterMinusOneMinusOne, true
	default:
		return TriggerCounterAny, false
	}
}
