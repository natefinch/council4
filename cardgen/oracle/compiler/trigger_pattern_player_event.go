package compiler

import (
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game/compare"
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
	event, ok := compilePlayerEventAction(clause.Action.Kind, clause.Card.Kind)
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
	if !modifiers.ok ||
		(occurrenceRequiresWhenever(clause.Occurrence.Kind) && kind != TriggerWhenever && !modifiers.self) {
		return pattern
	}
	turn, ok := compileSpellCastTurnRelation(clause.TurnRelation)
	if !ok {
		return pattern
	}
	pattern.Event = event
	pattern.Player = player
	pattern.OneOrMore = modifiers.oneOrMore
	pattern.ExcludeSelf = modifiers.excludeSelf
	if modifiers.self {
		pattern.Source = TriggerSourceSelf
	}
	pattern.PlayerEventOrdinalThisTurn = modifiers.ordinal
	pattern.ExcludeFirstDrawInDrawStep = modifiers.exceptFirstDrawInDrawStep
	pattern.PlaysExiledWithSource = modifiers.playsExiledWithSource
	pattern.CardSelection = modifiers.cardSelection
	pattern.CastDuringTurn = turn
	return pattern
}

func compilePlayerEventAction(action parser.PlayerEventActionKind, card parser.PlayerEventCardKind) (TriggerEvent, bool) {
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
	case parser.PlayerEventActionSearchLibrary:
		return TriggerEventLibrarySearched, true
	case parser.PlayerEventActionCommitCrime:
		return TriggerEventCrimeCommitted, true
	case parser.PlayerEventActionBecomeMonarch:
		return TriggerEventBecameMonarch, true
	case parser.PlayerEventActionPlay:
		// The play event splits on its card object: "a card exiled with <this>"
		// keys the linked-exile pool, while "a land" is the land-play action.
		switch card {
		case parser.PlayerEventCardExiledWithSource:
			return TriggerEventCardPlayedFromExile, true
		case parser.PlayerEventCardLand:
			return TriggerEventLandPlayed, true
		default:
			return TriggerEventUnknown, false
		}
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
	case parser.TriggerPlayerSelectorMonarch:
		return TriggerPlayerMonarch, true
	default:
		return TriggerPlayerAny, false
	}
}

type compiledPlayerEventModifiers struct {
	oneOrMore                 bool
	excludeSelf               bool
	self                      bool
	playsExiledWithSource     bool
	ordinal                   int
	exceptFirstDrawInDrawStep bool
	cardSelection             TriggerSelection
	ok                        bool
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
	ordinal, exceptFirstDrawInDrawStep, ok := compilePlayerEventOccurrence(action, player, occurrence)
	return compiledPlayerEventModifiers{
		oneOrMore:                 compiledCard.oneOrMore,
		excludeSelf:               compiledCard.excludeSelf,
		self:                      compiledCard.self,
		playsExiledWithSource:     compiledCard.playsExiledWithSource,
		ordinal:                   ordinal,
		exceptFirstDrawInDrawStep: exceptFirstDrawInDrawStep,
		cardSelection:             compiledCard.cardSelection,
		ok:                        ok,
	}
}

type compiledPlayerEventCard struct {
	oneOrMore             bool
	excludeSelf           bool
	self                  bool
	playsExiledWithSource bool
	cardSelection         TriggerSelection
	ok                    bool
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
	case parser.PlayerEventCardThis:
		ok := action == parser.PlayerEventActionCycle ||
			action == parser.PlayerEventActionCycleOrDiscard
		return compiledPlayerEventCard{self: ok, ok: ok}
	case parser.PlayerEventCardExiledWithSource:
		ok := action == parser.PlayerEventActionPlay
		return compiledPlayerEventCard{playsExiledWithSource: ok, ok: ok}
	case parser.PlayerEventCardLand:
		ok := action == parser.PlayerEventActionPlay
		return compiledPlayerEventCard{ok: ok}
	default:
		return compiledPlayerEventCard{}
	}
}

// compilePlayerEventCardSelection lowers a player-event card-type filter into a
// TriggerSelection. A filter is representable only for discard, where
// card-type-filtered discard triggers occur (CR 603.2).
func compilePlayerEventCardSelection(card parser.PlayerEventCard) (TriggerSelection, bool) {
	if len(card.RequiredTypes) == 0 && len(card.ExcludedTypes) == 0 &&
		len(card.RequiredTypesAny) == 0 && len(card.RequiredSubtypesAny) == 0 {
		return TriggerSelection{}, true
	}
	var selection TriggerSelection
	for _, value := range card.RequiredTypes {
		compiled := compileTriggerCardType(value)
		if compiled == "" {
			return TriggerSelection{}, false
		}
		selection.RequiredTypes = append(selection.RequiredTypes, compiled)
	}
	for _, value := range card.ExcludedTypes {
		compiled := compileTriggerCardType(value)
		if compiled == "" {
			return TriggerSelection{}, false
		}
		selection.ExcludedTypes = append(selection.ExcludedTypes, compiled)
	}
	for _, value := range card.RequiredTypesAny {
		compiled := compileTriggerCardType(value)
		if compiled == "" {
			return TriggerSelection{}, false
		}
		selection.RequiredTypesAny = append(selection.RequiredTypesAny, compiled)
	}
	selection.SubtypesAny = append(selection.SubtypesAny, card.RequiredSubtypesAny...)
	return selection, true
}

func compilePlayerEventOccurrence(
	action parser.PlayerEventActionKind,
	player parser.TriggerPlayerSelectorKind,
	occurrence parser.PlayerEventOccurrence,
) (ordinal int, exceptFirstDrawInDrawStep, ok bool) {
	switch occurrence.Kind {
	case parser.PlayerEventOccurrenceAny:
		return 0, false, occurrence.Ordinal == 0
	case parser.PlayerEventOccurrenceFirstEachTurn:
		return 1, false, occurrence.Ordinal == 1 && playerEventFirstEachTurnAllowed(action, player)
	case parser.PlayerEventOccurrenceOrdinalEachTurn:
		return occurrence.Ordinal, false, action == parser.PlayerEventActionDraw &&
			occurrence.Ordinal >= 1 &&
			occurrence.Ordinal <= 5
	case parser.PlayerEventOccurrenceExceptFirstInDrawStep:
		return 0, true, action == parser.PlayerEventActionDraw
	default:
		return 0, false, false
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
	if clause.Player.Kind == parser.TriggerPlayerSelectorMonarch {
		// The monarch step-player scope rides Player (matched against the step's
		// active player), not the source-controller-relative Controller.
		pattern.Player = TriggerPlayerMonarch
	} else {
		pattern.Controller = controller
	}
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
	case parser.TriggerPlayerSelectorAny, parser.TriggerPlayerSelectorMonarch:
		// Monarch scope is routed to pattern.Player by the caller; the returned
		// controller is unused (Any) in that path, matching the Any selector.
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
		parser.PlayerEventActionCycleOrDiscard,
		parser.PlayerEventActionPlay:
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
	case parser.PlayerEventActionDiscard, parser.PlayerEventActionCycle:
		return player == parser.TriggerPlayerSelectorYou
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
		selection.Keyword == parser.KeywordUnknown &&
		selection.ExcludedKeyword == parser.KeywordUnknown &&
		!selection.NonToken &&
		!selection.TokenOnly &&
		selection.ManaValueAtLeast == 0 &&
		selection.ManaValueAtMost == 0 &&
		!selection.MatchManaValue &&
		selection.ManaValue.Op == compare.Any &&
		selection.Power.Op == compare.Any &&
		selection.Toughness.Op == compare.Any &&
		selection.Controller == ControllerAny
}
