package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
)

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
		// A reflexive "When you do," gate shares the prior-instruction-accepted
		// predicate with "If you do,"; carry the parser's reflexive marker onto
		// the compiled condition so lowering can route the gated consequence to a
		// reflexive triggered ability.
		condition.Reflexive = clauses[condition.ClauseIndex].Reflexive
		compileConditionClause(condition, &clauses[condition.ClauseIndex])
	}
}

// simplePredicateMap holds the typed parser predicates that map one-to-one onto
// a semantic predicate with no additional clause data. Keeping these out of the
// large clause switch keeps that function's complexity manageable.
var simplePredicateMap = map[parser.ConditionPredicateKind]ConditionPredicate{
	parser.ConditionPredicateControllerHandEmpty:                              ConditionPredicateControllerHandEmpty,
	parser.ConditionPredicateEventSubjectWasKicked:                            ConditionPredicateEventSubjectWasKicked,
	parser.ConditionPredicateEventSubjectWasBargained:                         ConditionPredicateEventSubjectWasBargained,
	parser.ConditionPredicateEventSubjectWasCast:                              ConditionPredicateEventSubjectWasCast,
	parser.ConditionPredicateEventSubjectWasCastByController:                  ConditionPredicateEventSubjectWasCastByController,
	parser.ConditionPredicateEventSubjectWasCastFromControllerHand:            ConditionPredicateEventSubjectWasCastFromControllerHand,
	parser.ConditionPredicateEventSubjectEnteredOrCastFromGraveyard:           ConditionPredicateEventSubjectEnteredOrCastFromGraveyard,
	parser.ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard: ConditionPredicateEventSubjectEnteredOrCastFromControllerGraveyard,
	parser.ConditionPredicateEventSubjectWasNotPutByThisAbility:               ConditionPredicateEventSubjectWasNotPutByThisAbility,
	parser.ConditionPredicatePriorInstructionNotAccepted:                      ConditionPredicatePriorInstructionNotAccepted,
	parser.ConditionPredicatePriorInstructionAccepted:                         ConditionPredicatePriorInstructionAccepted,
	parser.ConditionPredicateDestroyedThisWay:                                 ConditionPredicateDestroyedThisWay,
	parser.ConditionPredicateDiesThisWay:                                      ConditionPredicateDiesThisWay,
	parser.ConditionPredicateNoLifeLostThisWay:                                ConditionPredicateNoLifeLostThisWay,
	parser.ConditionPredicateEventPlayerDoesNotPay:                            ConditionPredicateEventPlayerDoesNotPay,
	parser.ConditionPredicateDefendingPlayerDoesNotPay:                        ConditionPredicateDefendingPlayerDoesNotPay,
	parser.ConditionPredicateControllerCounterPlacement:                       ConditionPredicateControllerCounterPlacement,
	parser.ConditionPredicateTokenCreationUnderController:                     ConditionPredicateTokenCreationUnderController,
	parser.ConditionPredicateSpellCopyUnderController:                         ConditionPredicateSpellCopyUnderController,
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
	parser.ConditionPredicateSpellWasKicked:                                   ConditionPredicateSpellWasKicked,
	parser.ConditionPredicateSpellWasBargained:                                ConditionPredicateSpellWasBargained,
	parser.ConditionPredicateGiftPromised:                                     ConditionPredicateGiftPromised,
	parser.ConditionPredicateSpellWasCastFromGraveyard:                        ConditionPredicateSpellWasCastFromGraveyard,
	parser.ConditionPredicateSourceSaddled:                                    ConditionPredicateSourceSaddled,
	parser.ConditionPredicateSourceNotSaddled:                                 ConditionPredicateSourceNotSaddled,
	parser.ConditionPredicateFirstCombatPhaseOfTurn:                           ConditionPredicateFirstCombatPhaseOfTurn,
	parser.ConditionPredicateControllerCombatPhase:                            ConditionPredicateControllerCombatPhase,
	parser.ConditionPredicateControlsGreatestPowerCreature:                    ConditionPredicateControlsGreatestPowerCreature,
	parser.ConditionPredicateControlsGreatestToughnessCreature:                ConditionPredicateControlsGreatestToughnessCreature,
	parser.ConditionPredicateEventSubjectPowerGreatestOnBattlefield:           ConditionPredicateEventSubjectPowerGreatestOnBattlefield,
	parser.ConditionPredicateSubjectSharesCreatureTypeWithSource:              ConditionPredicateSubjectSharesCreatureTypeWithSource,
	parser.ConditionPredicateControllerIsMonarch:                              ConditionPredicateControllerIsMonarch,
	parser.ConditionPredicateControllerWasMonarchAtTurnStart:                  ConditionPredicateControllerWasMonarchAtTurnStart,
	parser.ConditionPredicateAnOpponentIsMonarch:                              ConditionPredicateAnOpponentIsMonarch,
	parser.ConditionPredicateNoMonarch:                                        ConditionPredicateNoMonarch,
	parser.ConditionPredicateDefendingPlayerIsMonarch:                         ConditionPredicateDefendingPlayerIsMonarch,
	parser.ConditionPredicateThatPlayerIsMonarch:                              ConditionPredicateThatPlayerIsMonarch,
	parser.ConditionPredicateControllerHasInitiative:                          ConditionPredicateControllerHasInitiative,
	parser.ConditionPredicateControllerHasCityBlessing:                        ConditionPredicateControllerHasCityBlessing,
	parser.ConditionPredicateControllerCompletedADungeon:                      ConditionPredicateControllerCompletedADungeon,
	parser.ConditionPredicateControllerTurn:                                   ConditionPredicateControllerTurn,
	parser.ConditionPredicateLandEnteredThisTurnOrControlsBasic:               ConditionPredicateLandEnteredThisTurnOrControlsBasic,
	parser.ConditionPredicateAllPlayersHandEmpty:                              ConditionPredicateAllPlayersHandEmpty,
}

// compileConditionClause mechanically maps one typed ConditionClause onto the
// semantic condition. Any clause whose typed selection, counter, or scope cannot
// be expressed in the closed semantic vocabulary leaves the predicate
// unsupported (fail closed).
func compileConditionClause(condition *CompiledCondition, clause *parser.ConditionClause) {
	if predicate, ok := simplePredicateMap[clause.Predicate]; ok {
		condition.Predicate = predicate
		if clause.Negated {
			condition.Negated = !condition.Negated
		}
		return
	}
	switch clause.Predicate {
	case parser.ConditionPredicateControllerLifeAtLeast:
		condition.Predicate = ConditionPredicateControllerLifeAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerLifeAtMost:
		condition.Predicate = ConditionPredicateControllerLifeAtMost
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerLifeAtLeastAboveStarting:
		condition.Predicate = ConditionPredicateControllerLifeAtLeastAboveStarting
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerHandSizeAtLeast:
		condition.Predicate = ConditionPredicateControllerHandSizeAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerHandSizeExactly:
		condition.Predicate = ConditionPredicateControllerHandSizeExactly
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerLibrarySizeAtLeast:
		condition.Predicate = ConditionPredicateControllerLibrarySizeAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerLifeExactly:
		condition.Predicate = ConditionPredicateControllerLifeExactly
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateSpellXAtLeast:
		condition.Predicate = ConditionPredicateSpellXAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateSourceAbilityResolutionOrdinalThisTurn:
		condition.Predicate = ConditionPredicateSourceAbilityResolutionOrdinalThisTurn
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControllerTurnOfGameAtMost:
		condition.Predicate = ConditionPredicateControllerTurnOfGameAtMost
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateColoredManaSpentToCastAtLeast:
		manaColor := compileTriggerColor(clause.ManaSpentColor)
		if manaColor == "" {
			return
		}
		condition.Predicate = ConditionPredicateColoredManaSpentToCastAtLeast
		condition.Threshold = clause.Threshold
		condition.ManaSpentColor = manaColor
	case parser.ConditionPredicateSameColorManaSpentToCastAtLeast:
		condition.Predicate = ConditionPredicateSameColorManaSpentToCastAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateEventSpellManaSpentToCastAtLeast:
		condition.Predicate = ConditionPredicateEventSpellManaSpentToCastAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateEventSpellCreatureManaSpentToCastAtLeast:
		condition.Predicate = ConditionPredicateEventSpellCreatureManaSpentToCastAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateEventSpellNoManaSpentToCast:
		condition.Predicate = ConditionPredicateEventSpellManaSpentToCastAtMost
		condition.Threshold = 0
	case parser.ConditionPredicateTriggeringPlayerHandSizeAtMost:
		condition.Predicate = ConditionPredicateTriggeringPlayerHandSizeAtMost
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateTriggeringPlayerHandSizeAtLeast:
		condition.Predicate = ConditionPredicateTriggeringPlayerHandSizeAtLeast
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
	case parser.ConditionPredicateGraveyardCardOfTypeCountAtLeast:
		condition.Predicate = ConditionPredicateControllerGraveyardCardOfTypeCountAtLeast
		condition.Threshold = clause.Threshold
		condition.GraveyardCountCardType = compileTriggerCardType(clause.GraveyardCountCardType)
	case parser.ConditionPredicateGraveyardInstantOrSorceryCountAtLeast:
		condition.Predicate = ConditionPredicateControllerGraveyardInstantOrSorceryCountAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateGraveyardPermanentCardCountAtLeast:
		condition.Predicate = ConditionPredicateControllerGraveyardPermanentCardCountAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateGraveyardManaValueCountAtLeast:
		condition.Predicate = ConditionPredicateControllerGraveyardManaValueCountAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateAnyOpponentGraveyardCardCountAtLeast:
		condition.Predicate = ConditionPredicateAnyOpponentGraveyardCardCountAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateCreaturePowerDiversityAtLeast:
		condition.Predicate = ConditionPredicateControllerCreaturePowerDiversityAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateAttackersAttackingControllerAtLeast:
		condition.Predicate = ConditionPredicateAttackersAttackingControllerAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateNoAttackerAttackedController:
		condition.Predicate = ConditionPredicateNoAttackerAttackedController
	case parser.ConditionPredicateObjectAttackedThisTurn:
		condition.Predicate = ConditionPredicateObjectAttackedThisTurn
		condition.ObjectBinding = compileConditionObjectBinding(clause.ObjectBinding)
	case parser.ConditionPredicateControllerGainedLifeThisTurnAtLeast:
		condition.Predicate = ConditionPredicateControllerGainedLifeThisTurnAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast:
		condition.Predicate = ConditionPredicateAnyOpponentDealtDamageThisTurnAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateAnyOpponentLostLifeThisTurnAtLeast:
		condition.Predicate = ConditionPredicateAnyOpponentLostLifeThisTurnAtLeast
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateAnyLibrarySizeAtMost:
		condition.Predicate = ConditionPredicateAnyLibrarySizeAtMost
		condition.Threshold = clause.Threshold
	case parser.ConditionPredicateControls:
		compileControlsCondition(condition, clause)
	case parser.ConditionPredicateControllerControlsNamed:
		if len(clause.ControlledNames) == 0 {
			return
		}
		condition.Predicate = ConditionPredicateControllerControlsNamed
		condition.ControlledNames = append(condition.ControlledNames, clause.ControlledNames...)
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
	case parser.ConditionPredicateControlsGreatestManaValueInGroup:
		selection, ok := compileConditionSelection(clause.Selection)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateControlsGreatestManaValueInGroup
		condition.Selection = selection
	case parser.ConditionPredicateEventSubjectHadNoCounter:
		counter, ok := compileConditionCounter(clause.Counter)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateEventSubjectHadNoCounter
		condition.Counter = counter
	case parser.ConditionPredicateEventSubjectHadCounter:
		counter, ok := compileConditionCounter(clause.Counter)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateEventSubjectHadCounter
		condition.Counter = counter
	case parser.ConditionPredicateCounterPlacementOnControlledCreature:
		counter, ok := compileConditionCounter(clause.Counter)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateCounterPlacementOnControlledCreature
		condition.Counter = counter
	case parser.ConditionPredicateCounterPlacementOnSelf:
		counter, ok := compileConditionCounter(clause.Counter)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateCounterPlacementOnSelf
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
		condition.CounterRecipientExcludesSource = clause.CounterRecipientExcludesSource
		condition.CounterRecipientIncludesController = clause.CounterRecipientIncludesController
	case parser.ConditionPredicateDamageByControlledSource:
		selection, ok := compileConditionSelection(clause.Selection)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateDamageByControlledSource
		condition.Selection = selection
	case parser.ConditionPredicateDamageWouldBeDealtToPermanent:
		selection, ok := compileConditionSelection(clause.Selection)
		if !ok {
			return
		}
		condition.Predicate = ConditionPredicateDamageWouldBeDealtToPermanent
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
		if clause.Negated {
			condition.Negated = !condition.Negated
		}
	case parser.ConditionPredicateObjectExists:
		condition.Predicate = ConditionPredicateObjectExists
		condition.ObjectBinding = compileConditionObjectBinding(clause.ObjectBinding)
	case parser.ConditionPredicateCardWouldGoToGraveyard:
		condition.Predicate = ConditionPredicateCardWouldGoToGraveyard
		condition.GraveyardRedirectScope = compileGraveyardRedirectScope(clause.GraveyardRedirectScope)
		condition.GraveyardRedirectControlScope = compileGraveyardRedirectControlScope(clause.GraveyardRedirectControlScope)
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
	case parser.ConditionControlScopeDefendingPlayer:
		return ConditionPredicateDefendingPlayerControls, true
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
	case parser.ConditionControlScopeTriggeringPlayer:
		return ConditionComparisonScopeTriggeringPlayer, true
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
		col, ok := conditionColorFromTrigger(value)
		if !ok {
			return ConditionSelection{}, false
		}
		selection.ColorsAny = append(selection.ColorsAny, col)
	}
	tapped, ok := conditionTriStateFromParser(syntax.Tapped)
	if !ok {
		return ConditionSelection{}, false
	}
	combatState, ok := conditionCombatStateFromParser(syntax.CombatState)
	if !ok {
		return ConditionSelection{}, false
	}
	attachment, ok := conditionAttachmentFromParser(syntax.Attachment)
	if !ok {
		return ConditionSelection{}, false
	}
	selection.Colorless = syntax.Colorless
	selection.Multicolored = syntax.Multicolored
	selection.TokenOnly = syntax.TokenOnly
	selection.NonToken = syntax.NonToken
	selection.ExcludeSource = syntax.ExcludeSource
	selection.Tapped = tapped
	selection.CombatState = combatState
	selection.Attachment = attachment
	selection.Keyword = syntax.Keyword
	selection.PowerAtLeast = syntax.PowerAtLeast
	selection.MatchPowerAtLeast = syntax.MatchPowerAtLeast
	selection.TotalPowerAtLeast = syntax.TotalPowerAtLeast
	selection.MatchTotalPowerAtLeast = syntax.MatchTotalPowerAtLeast
	selection.TotalPowerAtMost = syntax.TotalPowerAtMost
	selection.MatchTotalPowerAtMost = syntax.MatchTotalPowerAtMost
	selection.DistinctNamesAtLeast = syntax.DistinctNamesAtLeast
	selection.MatchDistinctNamesAtLeast = syntax.MatchDistinctNamesAtLeast
	selection.DamageRecipientOpponent = syntax.DamageRecipientOpponent
	selection.DamageRecipientOpponentPlayerOnly = syntax.DamageRecipientOpponentPlayerOnly
	selection.DamageNoncombatOnly = syntax.DamageNoncombatOnly
	selection.DamageSourceAnyController = syntax.DamageSourceAnyController
	selection.DamageRecipientController = syntax.DamageRecipientController
	selection.DamageSourceControllerOpponent = syntax.DamageSourceControllerOpponent
	selection.DamageRecipientSelf = syntax.DamageRecipientSelf
	selection.DamageRecipientAttached = syntax.DamageRecipientAttached
	selection.DamageRecipientMonarchGate = syntax.DamageRecipientMonarchGate
	selection.AnyCounter = syntax.AnyCounter
	selection.CounterKind = syntax.CounterKind
	selection.CounterKindKnown = syntax.CounterKindKnown
	selection.CounterCountAtLeast = syntax.CounterCountAtLeast
	selection.CounterCountLessThan = syntax.CounterCountLessThan
	return selection, true
}

func conditionCardTypeFromTrigger(value parser.TriggerCardType) (types.Card, bool) {
	switch value {
	case parser.TriggerCardTypeArtifact:
		return types.Artifact, true
	case parser.TriggerCardTypeBattle:
		return types.Battle, true
	case parser.TriggerCardTypeCreature:
		return types.Creature, true
	case parser.TriggerCardTypeEnchantment:
		return types.Enchantment, true
	case parser.TriggerCardTypeLand:
		return types.Land, true
	case parser.TriggerCardTypePlaneswalker:
		return types.Planeswalker, true
	default:
		return "", false
	}
}

func conditionSupertypeFromParser(value parser.ConditionSupertype) (types.Super, bool) {
	switch value {
	case parser.ConditionSupertypeBasic:
		return types.Basic, true
	case parser.ConditionSupertypeSnow:
		return types.Snow, true
	case parser.ConditionSupertypeLegendary:
		return types.Legendary, true
	default:
		return "", false
	}
}

func conditionColorFromTrigger(value parser.TriggerColor) (color.Color, bool) {
	switch value {
	case parser.TriggerColorWhite:
		return color.White, true
	case parser.TriggerColorBlue:
		return color.Blue, true
	case parser.TriggerColorBlack:
		return color.Black, true
	case parser.TriggerColorRed:
		return color.Red, true
	case parser.TriggerColorGreen:
		return color.Green, true
	default:
		return "", false
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

func conditionAttachmentFromParser(value parser.ConditionAttachmentState) (ConditionAttachment, bool) {
	switch value {
	case parser.ConditionAttachmentAny:
		return ConditionAttachmentNone, true
	case parser.ConditionAttachmentEnchanted:
		return ConditionAttachmentEnchanted, true
	case parser.ConditionAttachmentEquipped:
		return ConditionAttachmentEquipped, true
	default:
		return ConditionAttachmentNone, false
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
	case parser.ConditionObjectBindingSourceAttached:
		return ReferenceBindingSourceAttached
	case parser.ConditionObjectBindingCreatedToken:
		return ReferenceBindingCreatedToken
	case parser.ConditionObjectBindingPriorInstructionResult:
		return ReferenceBindingPriorInstructionResult
	case parser.ConditionObjectBindingTarget:
		return ReferenceBindingTarget
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

func compileGraveyardRedirectControlScope(value parser.GraveyardRedirectControlScope) GraveyardRedirectControlScope {
	switch value {
	case parser.GraveyardRedirectControlScopeYou:
		return GraveyardRedirectControlScopeYou
	case parser.GraveyardRedirectControlScopeOpponent:
		return GraveyardRedirectControlScopeOpponent
	default:
		return GraveyardRedirectControlScopeAny
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
			ConditionPredicateEventSubjectHadCounters,
			ConditionPredicateObjectAttackedThisTurn:
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
	found := binding == ReferenceBindingSource ||
		binding == ReferenceBindingEventPermanent ||
		binding == ReferenceBindingSourceAttached ||
		binding == ReferenceBindingCreatedToken ||
		binding == ReferenceBindingPriorInstructionResult ||
		binding == ReferenceBindingTarget
	for _, reference := range references {
		if !condition.Order.Contains(reference.Order) {
			continue
		}
		if reference.Binding != ReferenceBindingSource &&
			reference.Binding != ReferenceBindingEventPermanent &&
			reference.Binding != ReferenceBindingPriorInstructionResult &&
			reference.Binding != ReferenceBindingTarget {
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
