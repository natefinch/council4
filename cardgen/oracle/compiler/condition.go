package compiler

import "github.com/natefinch/council4/cardgen/oracle/parser"

// recognizeCondition maps the typed parser syntax that spans this condition onto
// closed semantic data. It consumes typed ConditionClause and
// EventHistoryCondition nodes only; it inspects no Oracle source text or tokens
// to derive meaning. The retained condition text and span remain available for
// diagnostics and exact source consumption accounting.
func recognizeCondition(
	condition *CompiledCondition,
	clauses []parser.ConditionClause,
	eventHistories []parser.EventHistoryCondition,
) {
	condition.Predicate = ConditionPredicateUnsupported
	// The introducer inverts the base predicate: an "unless" condition fires
	// when its predicate is false. Numeric "at most" comparisons invert this
	// again during clause compilation.
	condition.Negated = condition.Kind == ConditionUnless
	if recognizeEventHistoryCondition(condition, eventHistories) {
		return
	}
	if condition.ClauseIndex >= 0 {
		compileConditionClause(condition, &clauses[condition.ClauseIndex])
	}
}

// simplePredicateMap holds the typed parser predicates that map one-to-one onto
// a semantic predicate with no additional clause data. Keeping these out of the
// large clause switch keeps that function's complexity manageable.
var simplePredicateMap = map[parser.ConditionPredicateKind]ConditionPredicate{
	parser.ConditionPredicateControllerHandEmpty:                              ConditionPredicateControllerHandEmpty,
	parser.ConditionPredicateEventSubjectWasKicked:                            ConditionPredicateEventSubjectWasKicked,
	parser.ConditionPredicateEventSubjectWasCast:                              ConditionPredicateEventSubjectWasCast,
	parser.ConditionPredicateEventSubjectWasCastByController:                  ConditionPredicateEventSubjectWasCastByController,
	parser.ConditionPredicateEventSubjectEnteredOrCastFromGraveyard:           ConditionPredicateEventSubjectEnteredOrCastFromGraveyard,
	parser.ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard: ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard,
	parser.ConditionPredicatePriorInstructionNotAccepted:                      ConditionPredicatePriorInstructionNotAccepted,
	parser.ConditionPredicatePriorInstructionAccepted:                         ConditionPredicatePriorInstructionAccepted,
	parser.ConditionPredicateEventPlayerDoesNotPay:                            ConditionPredicateEventPlayerDoesNotPay,
	parser.ConditionPredicateControllerCounterPlacement:                       ConditionPredicateControllerCounterPlacement,
	parser.ConditionPredicateTokenCreationUnderController:                     ConditionPredicateTokenCreationUnderController,
	parser.ConditionPredicateTokenCreationAnyController:                       ConditionPredicateTokenCreationAnyController,
	parser.ConditionPredicateControllerWouldCreateNamedToken:                  ConditionPredicateControllerWouldCreateNamedToken,
	parser.ConditionPredicateWouldDrawFromEmptyLibrary:                        ConditionPredicateWouldDrawFromEmptyLibrary,
	parser.ConditionPredicateCreatedTokenThisTurn:                             ConditionPredicateControllerCreatedTokenThisTurn,
	parser.ConditionPredicateCastDuringControllerMainPhase:                    ConditionPredicateCastDuringControllerMainPhase,
	parser.ConditionPredicateWouldDrawCard:                                    ConditionPredicateWouldDrawCard,
	parser.ConditionPredicateWouldDrawCardExceptFirstInDrawStep:               ConditionPredicateWouldDrawCardExceptFirstInDrawStep,
	parser.ConditionPredicateControllerLifeGain:                               ConditionPredicateControllerLifeGain,
	parser.ConditionPredicateOpponentLifeLossDuringControllerTurn:             ConditionPredicateOpponentLifeLossDuringControllerTurn,
	parser.ConditionPredicateOpponentLifeLoss:                                 ConditionPredicateOpponentLifeLoss,
	parser.ConditionPredicateAnyPlayerLifeLoss:                                ConditionPredicateAnyPlayerLifeLoss,
	parser.ConditionPredicateSourceTributeNotPaid:                             ConditionPredicateSourceTributeNotPaid,
	parser.ConditionPredicateControllerControlsCommander:                      ConditionPredicateControllerControlsCommander,
}

// compileConditionClause mechanically maps one typed ConditionClause onto the
// semantic condition. Any clause whose typed selection, counter, or scope cannot
// be expressed in the closed semantic vocabulary leaves the predicate
// unsupported (fail closed).
func compileConditionClause(condition *CompiledCondition, clause *parser.ConditionClause) {
	if predicate, ok := simplePredicateMap[clause.Predicate]; ok {
		condition.Predicate = predicate
		return
	}
	switch clause.Predicate {
	case parser.ConditionPredicateControllerLifeAtLeast:
		condition.Predicate = ConditionPredicateControllerLifeAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerHandSizeAtLeast:
		condition.Predicate = ConditionPredicateControllerHandSizeAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerHandSizeExactly:
		condition.Predicate = ConditionPredicateControllerHandSizeExactly
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateAnyOpponentPoisonAtLeast:
		condition.Predicate = ConditionPredicateAnyOpponentPoisonAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateAnyPlayerLifeAtMost:
		condition.Predicate = ConditionPredicateAnyPlayerLifeAtMost
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateOpponentCountAtLeast:
		condition.Predicate = ConditionPredicateOpponentCountAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateGraveyardCardCountAtLeast:
		condition.Predicate = ConditionPredicateControllerGraveyardCardCountAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateGraveyardCardTypeCountAtLeast:
		condition.Predicate = ConditionPredicateControllerGraveyardCardTypeCountAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateCreaturePowerDiversityAtLeast:
		condition.Predicate = ConditionPredicateControllerCreaturePowerDiversityAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControls:
		compileControlsCondition(condition, clause)
	case parser.ConditionPredicateControlComparison:
		compileControlComparisonCondition(condition, clause)
	case parser.ConditionPredicateEventSubjectHadCounters:
		condition.Predicate = ConditionPredicateEventSubjectHadCounters
		condition.ObjectBinding = compileConditionObjectBinding(clause.ObjectBinding)
	case parser.ConditionPredicateEventSubjectNameUnique:
		condition.Predicate = ConditionPredicateEventSubjectNameUnique
		condition.ObjectBinding = compileConditionObjectBinding(clause.ObjectBinding)
	case parser.ConditionPredicateTargetColor:
		selection, ok := compileConditionSelection(clause.Selection)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateTargetColor
		condition.Selection = selection
	case parser.ConditionPredicateEventSubjectHadNoCounter:
		counter, ok := compileConditionCounter(clause.Counter)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateEventSubjectHadNoCounter
		condition.Counter = counter
	case parser.ConditionPredicateCounterPlacementOnControlledCreature:
		counter, ok := compileConditionCounter(clause.Counter)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateCounterPlacementOnControlledCreature
		condition.Counter = counter
	case parser.ConditionPredicateCounterPlacementOnAnyCreature:
		counter, ok := compileConditionCounter(clause.Counter)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateCounterPlacementOnAnyCreature
		condition.Counter = counter
	case parser.ConditionPredicateControllerCounterPlacement:
		condition.Predicate = ConditionPredicateControllerCounterPlacement
	case parser.ConditionPredicateCounterPlacementOnControlledPermanent:
		condition.Predicate = ConditionPredicateCounterPlacementOnControlledPermanent
		if counter, ok := compileConditionCounter(clause.Counter); ok {
			condition.Counter = counter
		}
		for _, value := range clause.CounterRecipientTypesAny {
			condition.CounterRecipientTypesAny = append(condition.CounterRecipientTypesAny, compileTriggerCardType(value))
		}
	case parser.ConditionPredicateDamageByControlledSource:
		selection, ok := compileConditionSelection(clause.Selection)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateDamageByControlledSource
		condition.Selection = selection
	case parser.ConditionPredicateSourceWouldDie:
		condition.Predicate = ConditionPredicateSourceWouldDie
		condition.SubjectSpan = clause.SubjectSpan
		condition.SubjectRefID = clause.SubjectRefID
	case parser.ConditionPredicateSourceWouldGoToGraveyard:
		condition.Predicate = ConditionPredicateSourceWouldGoToGraveyard
		condition.SubjectSpan = clause.SubjectSpan
		condition.SubjectRefID = clause.SubjectRefID
	case parser.ConditionPredicateObjectMatches:
		selection, ok := compileConditionSelection(clause.Selection)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateObjectMatches
		condition.ObjectBinding = compileConditionObjectBinding(clause.ObjectBinding)
		condition.Selection = selection
	case parser.ConditionPredicateObjectExists:
		condition.Predicate = ConditionPredicateObjectExists
		condition.ObjectBinding = compileConditionObjectBinding(clause.ObjectBinding)
	case parser.ConditionPredicateCardWouldGoToGraveyard:
		condition.Predicate = ConditionPredicateCardWouldGoToGraveyard
		condition.GraveyardRedirectScope = compileGraveyardRedirectScope(clause.GraveyardRedirectScope)
		condition.GraveyardFromBattlefieldOnly = clause.GraveyardFromBattlefieldOnly
		for _, value := range clause.GraveyardSubjectTypesAny {
			condition.GraveyardSubjectTypesAny = append(condition.GraveyardSubjectTypesAny, compileTriggerCardType(value))
		}
	default:
	}
}

// compileControlsCondition maps a typed "controls" predicate. The parser's
// control scope selects the counting players; its numeric comparison becomes a
// minimum threshold, with "at most N" expressed as a negated "at least N+1"
// using the same Selection-count vocabulary the lowering already consumes.
func compileControlsCondition(condition *CompiledCondition, clause *parser.ConditionClause) {
	predicate, ok := controlScopePredicate(clause.Scope)
	if !ok {
		return
	}
	selection, ok := compileConditionSelection(clause.Selection)
	if !ok {
		return
	}
	switch clause.Comparison {
	case parser.ConditionComparisonNone:
		condition.Threshold = 0
	case parser.ConditionComparisonAtLeast:
		condition.Threshold = clause.CompareValue
	case parser.ConditionComparisonAtMost:
		// "At most N" is expressed as a negated "at least N+1". This is only
		// sound for predicates whose runtime count is taken over a fixed set of
		// players. The existential "an opponent controls" predicate means
		// "some opponent controls at least N", so negating it yields "every
		// opponent controls fewer than N+1" rather than the intended "some
		// opponent controls at most N". The closed vocabulary cannot express an
		// existential upper bound, so fail closed.
		if clause.Scope == parser.ConditionControlScopeAnyOpponent {
			return
		}
		condition.Threshold = clause.CompareValue + 1
		condition.Negated = !condition.Negated
	default:
		return
	}
	condition.Predicate = predicate
	condition.Selection = selection
	condition.SourceInGraveyard = clause.SourceInGraveyard
}

func controlScopePredicate(scope parser.ConditionControlScope) (ConditionPredicate, bool) {
	switch scope {
	case parser.ConditionControlScopeController:
		return ConditionPredicateControllerControls, true
	case parser.ConditionControlScopeAnyOpponent:
		return ConditionPredicateAnyOpponentControls, true
	case parser.ConditionControlScopeOpponents:
		return ConditionPredicateOpponentsControl, true
	default:
		return ConditionPredicateUnsupported, false
	}
}

// compileControlComparisonCondition maps a typed cross-player control-count
// comparison. It fails closed unless exactly one side counts the controller and
// the counted Selection is expressible in the closed vocabulary.
func compileControlComparisonCondition(condition *CompiledCondition, clause *parser.ConditionClause) {
	left, ok := comparisonScopeFromParser(clause.ControlComparison.LeftScope)
	if !ok {
		return
	}
	right, ok := comparisonScopeFromParser(clause.ControlComparison.RightScope)
	if !ok {
		return
	}
	if (left == ConditionComparisonScopeController) == (right == ConditionComparisonScopeController) {
		return
	}
	selection, ok := compileConditionSelection(clause.Selection)
	if !ok {
		return
	}
	condition.Predicate = ConditionPredicateControlComparison
	condition.Selection = selection
	condition.ControlComparisonLeft = left
	condition.ControlComparisonRight = right
	condition.ControlComparisonGreater = clause.ControlComparison.Greater
}

func comparisonScopeFromParser(scope parser.ConditionControlScope) (ConditionComparisonScope, bool) {
	switch scope {
	case parser.ConditionControlScopeController:
		return ConditionComparisonScopeController, true
	case parser.ConditionControlScopeAnyOpponent:
		return ConditionComparisonScopeAnyOpponent, true
	case parser.ConditionControlScopeEachOpponent:
		return ConditionComparisonScopeEachOpponent, true
	default:
		return ConditionComparisonScopeController, false
	}
}

// compileConditionSelection maps a typed parser selection onto the closed
// semantic Selection vocabulary, failing closed if any card type, color,
// supertype, tapped state, or subtype identity is outside that vocabulary.
func compileConditionSelection(syntax parser.ConditionSelection) (ConditionSelection, bool) {
	var selection ConditionSelection
	for _, value := range syntax.RequiredTypes {
		cardType, ok := conditionCardTypeFromTrigger(value)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.RequiredTypes = append(selection.RequiredTypes, cardType)
	}
	for _, value := range syntax.Supertypes {
		supertype, ok := conditionSupertypeFromParser(value)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.Supertypes = append(selection.Supertypes, supertype)
	}
	for _, value := range syntax.SubtypesAny {
		if value == "" {
			return ConditionSelection{}, false
		}
		selection.SubtypesAny = append(selection.SubtypesAny, string(value))
	}
	for _, value := range syntax.ColorsAny {
		color, ok := conditionColorFromTrigger(value)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, color)
	}
	tapped, ok := conditionTriStateFromParser(syntax.Tapped)
	if !ok {
		return ConditionSelection{}, false
	}
	combatState, ok := conditionCombatStateFromParser(syntax.CombatState)
	if !ok {
		return ConditionSelection{}, false
	}
	selection.Colorless = syntax.Colorless
	selection.Multicolored = syntax.Multicolored
	selection.TokenOnly = syntax.TokenOnly
	selection.ExcludeSource = syntax.ExcludeSource
	selection.Tapped = tapped
	selection.CombatState = combatState
	selection.Keyword = syntax.Keyword
	selection.PowerAtLeast = syntax.PowerAtLeast
	selection.MatchPowerAtLeast = syntax.MatchPowerAtLeast
	selection.TotalPowerAtLeast = syntax.TotalPowerAtLeast
	selection.MatchTotalPowerAtLeast = syntax.MatchTotalPowerAtLeast
	selection.DistinctNamesAtLeast = syntax.DistinctNamesAtLeast
	selection.MatchDistinctNamesAtLeast = syntax.MatchDistinctNamesAtLeast
	selection.DamageRecipientOpponent = syntax.DamageRecipientOpponent
	selection.DamageNoncombatOnly = syntax.DamageNoncombatOnly
	selection.DamageSourceAnyController = syntax.DamageSourceAnyController
	selection.AnyCounter = syntax.AnyCounter
	return selection, true
}

func conditionCardTypeFromTrigger(value parser.TriggerCardType) (ConditionCardType, bool) {
	switch value {
	case parser.TriggerCardTypeArtifact:
		return ConditionCardTypeArtifact, true
	case parser.TriggerCardTypeBattle:
		return ConditionCardTypeBattle, true
	case parser.TriggerCardTypeCreature:
		return ConditionCardTypeCreature, true
	case parser.TriggerCardTypeEnchantment:
		return ConditionCardTypeEnchantment, true
	case parser.TriggerCardTypeLand:
		return ConditionCardTypeLand, true
	case parser.TriggerCardTypePlaneswalker:
		return ConditionCardTypePlaneswalker, true
	default:
		return ConditionCardTypeUnknown, false
	}
}

func conditionSupertypeFromParser(value parser.ConditionSupertype) (ConditionSupertype, bool) {
	switch value {
	case parser.ConditionSupertypeBasic:
		return ConditionSupertypeBasic, true
	case parser.ConditionSupertypeSnow:
		return ConditionSupertypeSnow, true
	case parser.ConditionSupertypeLegendary:
		return ConditionSupertypeLegendary, true
	default:
		return ConditionSupertypeUnknown, false
	}
}

func conditionColorFromTrigger(value parser.TriggerColor) (ConditionColor, bool) {
	switch value {
	case parser.TriggerColorWhite:
		return ConditionColorWhite, true
	case parser.TriggerColorBlue:
		return ConditionColorBlue, true
	case parser.TriggerColorBlack:
		return ConditionColorBlack, true
	case parser.TriggerColorRed:
		return ConditionColorRed, true
	case parser.TriggerColorGreen:
		return ConditionColorGreen, true
	default:
		return ConditionColorUnknown, false
	}
}

func conditionTriStateFromParser(value parser.ConditionTappedState) (ConditionTriState, bool) {
	switch value {
	case parser.ConditionTappedAny:
		return ConditionTriAny, true
	case parser.ConditionTappedTrue:
		return ConditionTriTrue, true
	case parser.ConditionTappedFalse:
		return ConditionTriFalse, true
	default:
		return ConditionTriAny, false
	}
}

func conditionCombatStateFromParser(value parser.ConditionCombatState) (ConditionCombatState, bool) {
	switch value {
	case parser.ConditionCombatAny:
		return ConditionCombatStateAny, true
	case parser.ConditionCombatAttacking:
		return ConditionCombatStateAttacking, true
	case parser.ConditionCombatBlocking:
		return ConditionCombatStateBlocking, true
	case parser.ConditionCombatAttackingOrBlocking:
		return ConditionCombatStateAttackingOrBlocking, true
	default:
		return ConditionCombatStateAny, false
	}
}

func compileConditionCounter(value parser.ConditionCounterKind) (ConditionCounter, bool) {
	switch value {
	case parser.ConditionCounterPlusOnePlusOne:
		return ConditionCounterPlusOnePlusOne, true
	case parser.ConditionCounterMinusOneMinusOne:
		return ConditionCounterMinusOneMinusOne, true
	default:
		return ConditionCounterUnknown, false
	}
}

func compileConditionObjectBinding(value parser.ConditionObjectBinding) ReferenceBinding {
	switch value {
	case parser.ConditionObjectBindingSource:
		return ReferenceBindingSource
	case parser.ConditionObjectBindingEventPermanent:
		return ReferenceBindingEventPermanent
	default:
		return ReferenceBindingUnsupported
	}
}

func compileGraveyardRedirectScope(value parser.GraveyardRedirectScope) GraveyardRedirectScope {
	switch value {
	case parser.GraveyardRedirectScopeYou:
		return GraveyardRedirectScopeYou
	case parser.GraveyardRedirectScopeOpponent:
		return GraveyardRedirectScopeOpponent
	default:
		return GraveyardRedirectScopeAny
	}
}

func recognizeEventHistoryCondition(
	condition *CompiledCondition,
	syntax []parser.EventHistoryCondition,
) bool {
	if condition.EventHistoryIndex < 0 {
		return false
	}
	history := &syntax[condition.EventHistoryIndex]
	pattern, ok := compileEventHistoryPattern(history)
	if !ok {
		return false
	}
	window, ok := compileEventHistoryWindow(history.Window.Kind)
	if !ok {
		return false
	}
	condition.Predicate = ConditionPredicateEventHistory
	condition.Negated = history.Negated
	condition.EventHistoryPattern = &pattern
	condition.EventHistoryWindow = window
	condition.EventHistoryMinCount = history.MinCount
	return true
}

func compileEventHistoryPattern(syntax *parser.EventHistoryCondition) (TriggerPattern, bool) {
	if syntax.TriggerEvent != nil && syntax.PlayerEvent != nil ||
		syntax.TriggerEvent == nil && syntax.PlayerEvent == nil {
		return TriggerPattern{}, false
	}
	if syntax.TriggerEvent != nil {
		return compileTriggerEventClause(syntax.TriggerEvent)
	}
	pattern := compilePlayerEventTriggerPattern(syntax.PlayerEvent, TriggerWhenever, nil)
	if pattern.Event == TriggerEventUnknown {
		return TriggerPattern{}, false
	}
	pattern.Kind = TriggerUnknown
	return pattern, true
}

func compileEventHistoryWindow(
	window parser.EventHistoryWindowKind,
) (ConditionEventHistoryWindow, bool) {
	switch window {
	case parser.EventHistoryWindowCurrentTurn:
		return ConditionEventHistoryWindowCurrentTurn, true
	case parser.EventHistoryWindowPreviousTurn:
		return ConditionEventHistoryWindowPreviousTurn, true
	default:
		return ConditionEventHistoryWindowCurrentTurn, false
	}
}

func bindConditionReferences(conditions []CompiledCondition, references []CompiledReference, trigger *CompiledTrigger) {
	for i := range conditions {
		switch conditions[i].Predicate {
		case ConditionPredicateSourceWouldDie, ConditionPredicateSourceWouldGoToGraveyard:
			if !conditionSubjectBindsSource(conditions[i], references) {
				conditions[i].Predicate = ConditionPredicateUnsupported
			}
		case ConditionPredicateObjectMatches,
			ConditionPredicateObjectExists,
			ConditionPredicateEventSubjectHadCounters:
			binding, ok := conditionObjectBinding(conditions[i], references)
			if !ok ||
				binding == ReferenceBindingEventPermanent &&
					(trigger == nil || trigger.Pattern.OneOrMore || !triggerEventBindsPermanent(trigger.Pattern.Event)) ||
				conditions[i].Predicate == ConditionPredicateObjectExists && binding != ReferenceBindingSource ||
				conditions[i].Predicate == ConditionPredicateEventSubjectHadCounters && binding != ReferenceBindingEventPermanent {
				conditions[i].Predicate = ConditionPredicateUnsupported
				continue
			}
			conditions[i].ObjectBinding = binding
		default:
		}
	}
}

func conditionObjectBinding(condition CompiledCondition, references []CompiledReference) (ReferenceBinding, bool) {
	binding := condition.ObjectBinding
	found := binding == ReferenceBindingSource || binding == ReferenceBindingEventPermanent
	for _, reference := range references {
		if !condition.Order.Contains(reference.Order) {
			continue
		}
		if reference.Binding != ReferenceBindingSource &&
			reference.Binding != ReferenceBindingEventPermanent {
			return ReferenceBindingUnsupported, false
		}
		if found && reference.Binding != binding {
			return ReferenceBindingUnsupported, false
		}
		binding = reference.Binding
		found = true
	}
	return binding, found
}

// conditionSubjectBindsSource reports whether a typed source reference exactly
// fills the condition's subject span. The parser owns the subject span; the
// compiler never re-derives it from condition text.
func conditionSubjectBindsSource(
	condition CompiledCondition,
	references []CompiledReference,
) bool {
	for _, reference := range references {
		if reference.NodeID == condition.SubjectRefID && reference.Binding == ReferenceBindingSource {
			return true
		}
	}
	return false
}
