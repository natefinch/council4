package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
)

func compilePlayerEventTriggerPattern(
	clause *parser.PlayerEventTriggerClause,
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
	event, ok := compilePlayerEventAction(clause.Action.Kind)
	if !ok {
		return pattern
	}
	player, ok := compilePlayerEventPlayer(&clause.Player)
	if !ok {
		return pattern
	}
	modifiers := compilePlayerEventModifiers(
		clause.Action.Kind,
		clause.Player.Kind,
		clause.Card,
		clause.Occurrence,
	)
	if !modifiers.ok || occurrenceRequiresWhenever(clause.Occurrence.Kind) && kind != TriggerWhenever {
		return pattern
	}
	pattern.Event = event
	pattern.Player = player
	pattern.OneOrMore = modifiers.oneOrMore
	pattern.ExcludeSelf = modifiers.excludeSelf
	pattern.PlayerEventOrdinalThisTurn = modifiers.ordinal
	pattern.CardSelection = modifiers.cardSelection
	return pattern
}

func compilePlayerEventAction(action parser.PlayerEventActionKind) (TriggerEvent, bool) {
	switch action {
	case parser.PlayerEventActionDraw:
		return TriggerEventCardDrawn, true
	case parser.PlayerEventActionDiscard, parser.PlayerEventActionCycleOrDiscard:
		return TriggerEventCardDiscarded, true
	case parser.PlayerEventActionCycle:
		return TriggerEventCycled, true
	case parser.PlayerEventActionScry:
		return TriggerEventScry, true
	case parser.PlayerEventActionSurveil:
		return TriggerEventSurveil, true
	case parser.PlayerEventActionGainLife:
		return TriggerEventLifeGained, true
	case parser.PlayerEventActionLoseLife:
		return TriggerEventLifeLost, true
	default:
		return TriggerEventUnknown, false
	}
}

func compilePlayerEventPlayer(player *parser.TriggerPlayerSelector) (TriggerPlayerRelation, bool) {
	switch player.Kind {
	case parser.TriggerPlayerSelectorAny:
		return TriggerPlayerAny, true
	case parser.TriggerPlayerSelectorYou:
		return TriggerPlayerYou, true
	case parser.TriggerPlayerSelectorOpponent:
		return TriggerPlayerOpponent, true
	default:
		return TriggerPlayerAny, false
	}
}

type compiledPlayerEventModifiers struct {
	oneOrMore     bool
	excludeSelf   bool
	ordinal       int
	cardSelection TriggerSelection
	ok            bool
}

func compilePlayerEventModifiers(
	action parser.PlayerEventActionKind,
	player parser.TriggerPlayerSelectorKind,
	card parser.PlayerEventCard,
	occurrence parser.PlayerEventOccurrence,
) compiledPlayerEventModifiers {
	compiledCard := compilePlayerEventCard(action, card)
	if !compiledCard.ok {
		return compiledPlayerEventModifiers{}
	}
	ordinal, ok := compilePlayerEventOccurrence(action, player, occurrence)
	return compiledPlayerEventModifiers{
		oneOrMore:     compiledCard.oneOrMore,
		excludeSelf:   compiledCard.excludeSelf,
		ordinal:       ordinal,
		cardSelection: compiledCard.cardSelection,
		ok:            ok,
	}
}

type compiledPlayerEventCard struct {
	oneOrMore     bool
	excludeSelf   bool
	cardSelection TriggerSelection
	ok            bool
}

func compilePlayerEventCard(action parser.PlayerEventActionKind, card parser.PlayerEventCard) compiledPlayerEventCard {
	if !playerEventActionHasCard(action) {
		return compiledPlayerEventCard{ok: card.Kind == parser.PlayerEventCardNone}
	}
	selection, ok := compilePlayerEventCardSelection(card)
	if !ok {
		return compiledPlayerEventCard{}
	}
	switch card.Kind {
	case parser.PlayerEventCardSingle:
		return compiledPlayerEventCard{cardSelection: selection, ok: true}
	case parser.PlayerEventCardOneOrMore:
		ok := action == parser.PlayerEventActionDiscard
		return compiledPlayerEventCard{oneOrMore: ok, cardSelection: selection, ok: ok}
	case parser.PlayerEventCardAnother:
		ok := action == parser.PlayerEventActionDiscard ||
			action == parser.PlayerEventActionCycle ||
			action == parser.PlayerEventActionCycleOrDiscard
		return compiledPlayerEventCard{excludeSelf: ok, ok: ok}
	default:
		return compiledPlayerEventCard{}
	}
}

// compilePlayerEventCardSelection lowers a player-event card-type filter into a
// TriggerSelection. A filter is representable only for discard, where
// card-type-filtered discard triggers occur (CR 603.2).
func compilePlayerEventCardSelection(card parser.PlayerEventCard) (TriggerSelection, bool) {
	if len(card.RequiredTypes) == 0 && len(card.ExcludedTypes) == 0 {
		return TriggerSelection{}, true
	}
	var selection TriggerSelection
	for _, value := range card.RequiredTypes {
		compiled := compileTriggerCardType(value)
		if compiled == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		selection.RequiredTypes = append(selection.RequiredTypes, compiled)
	}
	for _, value := range card.ExcludedTypes {
		compiled := compileTriggerCardType(value)
		if compiled == TriggerCardTypeUnknown {
			return TriggerSelection{}, false
		}
		selection.ExcludedTypes = append(selection.ExcludedTypes, compiled)
	}
	return selection, true
}

func compilePlayerEventOccurrence(
	action parser.PlayerEventActionKind,
	player parser.TriggerPlayerSelectorKind,
	occurrence parser.PlayerEventOccurrence,
) (int, bool) {
	switch occurrence.Kind {
	case parser.PlayerEventOccurrenceAny:
		return 0, occurrence.Ordinal == 0
	case parser.PlayerEventOccurrenceFirstEachTurn:
		return 1, occurrence.Ordinal == 1 && playerEventFirstEachTurnAllowed(action, player)
	case parser.PlayerEventOccurrenceOrdinalEachTurn:
		return occurrence.Ordinal, action == parser.PlayerEventActionDraw &&
			occurrence.Ordinal >= 1 &&
			occurrence.Ordinal <= 5
	default:
		return 0, false
	}
}

func occurrenceRequiresWhenever(occurrence parser.PlayerEventOccurrenceKind) bool {
	return occurrence == parser.PlayerEventOccurrenceAny
}

func compilePhaseStepTriggerPattern(
	clause *parser.PhaseStepTriggerClause,
	kind TriggerKind,
	condition *CompiledCondition,
) TriggerPattern {
	pattern := TriggerPattern{
		Span:                 clause.Span,
		Kind:                 kind,
		InterveningCondition: condition,
	}
	if kind != TriggerAt || !knownPhaseStepQuantifier(clause.Quantifier.Kind) {
		return pattern
	}
	step, ok := compilePhaseStepName(clause.Name.Kind)
	if !ok {
		return pattern
	}
	controller, attached, ok := compilePhaseStepPlayer(&clause.Player)
	if !ok {
		return pattern
	}
	pattern.Event = TriggerEventBeginningOfStep
	pattern.Step = step
	pattern.Controller = controller
	pattern.StepPlayerSourceAttachedSelection = attached
	pattern.NextOccurrence = clause.Next
	return pattern
}

func knownPhaseStepQuantifier(kind parser.PhaseStepQuantifierKind) bool {
	switch kind {
	case parser.PhaseStepQuantifierNone,
		parser.PhaseStepQuantifierSingle,
		parser.PhaseStepQuantifierEach,
		parser.PhaseStepQuantifierEachOf:
		return true
	default:
		return false
	}
}

func compilePhaseStepName(name parser.PhaseStepNameKind) (TriggerStep, bool) {
	switch name {
	case parser.PhaseStepNameUpkeep:
		return TriggerStepUpkeep, true
	case parser.PhaseStepNameDrawStep:
		return TriggerStepDraw, true
	case parser.PhaseStepNameEndStep:
		return TriggerStepEnd, true
	case parser.PhaseStepNameCombat, parser.PhaseStepNameCombatStep:
		return TriggerStepBeginningOfCombat, true
	case parser.PhaseStepNameEndOfCombat, parser.PhaseStepNameEndOfCombatStep:
		return TriggerStepEndOfCombat, true
	case parser.PhaseStepNamePrecombatMainPhase, parser.PhaseStepNameFirstMainPhase:
		return TriggerStepPrecombatMain, true
	case parser.PhaseStepNamePostcombatMainPhase, parser.PhaseStepNameSecondMainPhase:
		return TriggerStepPostcombatMain, true
	default:
		return TriggerStepNone, false
	}
}

func compilePhaseStepPlayer(player *parser.TriggerPlayerSelector) (ControllerKind, TriggerSelection, bool) {
	switch player.Kind {
	case parser.TriggerPlayerSelectorAny:
		return ControllerAny, TriggerSelection{}, true
	case parser.TriggerPlayerSelectorYou, parser.TriggerPlayerSelectorSourceController:
		return ControllerYou, TriggerSelection{}, true
	case parser.TriggerPlayerSelectorOpponent:
		return ControllerOpponent, TriggerSelection{}, true
	case parser.TriggerPlayerSelectorAttachedController:
		selection, ok := compilePhaseStepAttachedSubject(&player.AttachedSubject)
		return ControllerAny, selection, ok
	default:
		return ControllerAny, TriggerSelection{}, false
	}
}

func compilePhaseStepAttachedSubject(subject *parser.TriggerAttachedSubject) (TriggerSelection, bool) {
	selection, ok := compileTriggerSelection(subject.Selection)
	if !ok || phaseStepAttachedSelectionEmpty(selection) {
		// A wildcard Selection is empty, which the runtime interprets as no
		// attached-controller relation rather than any attached permanent.
		return TriggerSelection{}, false
	}
	return selection, true
}

func playerEventActionHasCard(action parser.PlayerEventActionKind) bool {
	switch action {
	case parser.PlayerEventActionDraw,
		parser.PlayerEventActionDiscard,
		parser.PlayerEventActionCycle,
		parser.PlayerEventActionCycleOrDiscard:
		return true
	default:
		return false
	}
}

func playerEventFirstEachTurnAllowed(action parser.PlayerEventActionKind, player parser.TriggerPlayerSelectorKind) bool {
	switch action {
	case parser.PlayerEventActionDraw,
		parser.PlayerEventActionScry,
		parser.PlayerEventActionSurveil:
		return true
	case parser.PlayerEventActionGainLife, parser.PlayerEventActionLoseLife:
		return player != parser.TriggerPlayerSelectorAny
	default:
		return false
	}
}

func phaseStepAttachedSelectionEmpty(selection TriggerSelection) bool {
	return len(selection.RequiredTypes) == 0 &&
		len(selection.RequiredTypesAny) == 0 &&
		len(selection.ExcludedTypes) == 0 &&
		len(selection.Supertypes) == 0 &&
		len(selection.SubtypesAny) == 0 &&
		len(selection.ColorsAny) == 0 &&
		len(selection.ExcludedColors) == 0 &&
		!selection.Colorless &&
		!selection.Multicolored &&
		selection.Tapped == TriggerTriAny &&
		selection.CombatState == TriggerCombatStateAny &&
		selection.Keyword == TriggerKeywordUnknown &&
		selection.ExcludedKeyword == TriggerKeywordUnknown &&
		!selection.NonToken &&
		!selection.TokenOnly &&
		selection.ManaValueAtLeast == 0 &&
		!selection.MatchManaValue &&
		selection.ManaValue.Comparison == TriggerComparisonUnknown &&
		selection.Power.Comparison == TriggerComparisonUnknown &&
		selection.Toughness.Comparison == TriggerComparisonUnknown &&
		selection.Controller == ControllerAny
}
