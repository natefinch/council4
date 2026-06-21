package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func emitEventHistoryConditions(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		if conditions := parseEventHistoryConditions(tokens); len(conditions) > 0 {
			ability.EventHistoryConditions = conditions
		}
		if ability.Modal == nil {
			continue
		}
		for j := range ability.Modal.Options {
			mode := &ability.Modal.Options[j]
			tokens := eventHistorySemanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
			if conditions := parseEventHistoryConditions(tokens); len(conditions) > 0 {
				mode.EventHistoryConditions = conditions
			}
		}
	}
}

func eventHistorySemanticTokens(
	tokens []shared.Token,
	reminders []Delimited,
	quoted []Delimited,
) []shared.Token {
	result := tokens
	for i := range reminders {
		result = tokensOutsideParserSpan(result, reminders[i].Span)
	}
	for i := range quoted {
		result = tokensOutsideParserSpan(result, quoted[i].Span)
	}
	return result
}

func parseEventHistoryConditions(tokens []shared.Token) []EventHistoryCondition {
	var conditions []EventHistoryCondition
	for i := 0; i < len(tokens); i++ {
		intro, width := conditionIntroAt(tokens, i)
		if intro != ConditionIntroIf && intro != ConditionIntroOnlyIf {
			continue
		}
		end := eventHistoryConditionEnd(tokens, i)
		if condition, ok := parseEventHistoryCondition(tokens[i:end], width); ok {
			conditions = append(conditions, condition)
		}
		i = end - 1
	}
	return conditions
}

func eventHistoryConditionEnd(tokens []shared.Token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Kind == shared.Period || i > start && tokens[i].Kind == shared.Comma {
			return i
		}
	}
	return len(tokens)
}

// parseEventHistoryCondition recognizes a single event-history condition from a
// clause whose first introWidth tokens form its introducer ("if" or the
// two-word "only if" that opens an "Activate only if ..." restriction). The
// retained span covers the whole clause including the introducer so it matches
// the parser's condition segment for that boundary.
func parseEventHistoryCondition(tokens []shared.Token, introWidth int) (EventHistoryCondition, bool) {
	if introWidth >= len(tokens) {
		return EventHistoryCondition{}, false
	}
	event := tokens[introWidth:]
	event, window, ok := parseEventHistoryWindow(event)
	if !ok || len(event) == 0 {
		return EventHistoryCondition{}, false
	}
	condition := EventHistoryCondition{
		Span:   shared.SpanOf(tokens),
		Window: window,
	}
	if rest, ok := cutTokenPrefix(event, "no"); ok {
		condition.Negated = true
		condition.NegationSpan = event[0].Span
		event = rest
	}
	event, condition.MinCount = stripAttackedCreatureCount(event)
	condition.TriggerEvent = parseEventHistoryTriggerEvent(event)
	if condition.TriggerEvent == nil {
		condition.PlayerEvent = parseEventHistoryPlayerEvent(event)
	}
	if condition.TriggerEvent == nil && condition.PlayerEvent == nil {
		return EventHistoryCondition{}, false
	}
	if !eventHistoryCombinationAllowed(&condition) {
		return EventHistoryCondition{}, false
	}
	return condition, true
}

func eventHistoryCombinationAllowed(condition *EventHistoryCondition) bool {
	if condition.TriggerEvent != nil {
		switch condition.TriggerEvent.Kind {
		case TriggerEventKindAttack, TriggerEventKindZoneChange:
			return !condition.Negated && condition.Window.Kind == EventHistoryWindowCurrentTurn
		case TriggerEventKindSpellCast:
			return condition.Negated && condition.Window.Kind == EventHistoryWindowPreviousTurn
		default:
			return false
		}
	}
	if condition.Negated || condition.PlayerEvent == nil {
		return false
	}
	if condition.PlayerEvent.Action.Kind == PlayerEventActionGainLife {
		return condition.Window.Kind == EventHistoryWindowCurrentTurn
	}
	return condition.PlayerEvent.Action.Kind == PlayerEventActionLoseLife
}

func parseEventHistoryWindow(tokens []shared.Token) ([]shared.Token, EventHistoryWindow, bool) {
	for _, candidate := range []struct {
		words []string
		kind  EventHistoryWindowKind
	}{
		{words: []string{"this", "turn"}, kind: EventHistoryWindowCurrentTurn},
		{words: []string{"last", "turn"}, kind: EventHistoryWindowPreviousTurn},
	} {
		prefix, ok := stripTokenSuffix(tokens, candidate.words...)
		if !ok {
			continue
		}
		return prefix, EventHistoryWindow{
			Kind: candidate.kind,
			Span: shared.SpanOf(tokens[len(prefix):]),
		}, true
	}
	return nil, EventHistoryWindow{}, false
}

// stripAttackedCreatureCount recognizes a trailing "with <count> creatures"
// qualifier on a past-tense "you attacked" event-history clause and returns the
// reduced "you attacked" tokens together with the minimum attacker count. Each
// attacker-declared event for a creature you control is one matching event, so
// the runtime counts those events against the returned minimum. When no
// qualifier is present the tokens are returned unchanged with a zero count,
// which the runtime treats as a single matching event.
func stripAttackedCreatureCount(tokens []shared.Token) (reduced []shared.Token, minCount int) {
	rest, ok := cutTokenPrefix(tokens, "you", "attacked", "with")
	if !ok {
		return tokens, 0
	}
	attacked := tokens[:2]
	if tokenWordsEqual(rest, "a", "creature") ||
		tokenWordsEqual(rest, "one", "or", "more", "creatures") {
		return attacked, 1
	}
	if count, ok := attackWithCreatureCount(tokens[2:]); ok {
		return attacked, count
	}
	return tokens, 0
}

func parseEventHistoryTriggerEvent(tokens []shared.Token) *TriggerEventClause {
	span := shared.SpanOf(tokens)
	if tokenWordsEqual(tokens, "you", "attacked") {
		return &TriggerEventClause{
			Kind: TriggerEventKindAttack,
			Span: span,
			Actor: TriggerEventActor{
				Kind: TriggerEventActorYou,
				Span: tokens[0].Span,
			},
		}
	}
	if tokenWordsEqual(tokens, "a", "creature", "died") {
		return &TriggerEventClause{
			Kind: TriggerEventKindZoneChange,
			Span: span,
			Subject: TriggerEventSubject{
				Kind: TriggerEventSubjectSelection,
				Span: shared.SpanOf(tokens[:2]),
				Selection: TriggerSelection{
					RequiredTypes: []TriggerCardType{TriggerCardTypeCreature},
				},
			},
			ZoneChange: TriggerEventZoneChange{
				Kind: TriggerEventZoneChangeDied,
				Span: tokens[2].Span,
			},
			Zone: TriggerEventZoneContext{
				Span:          tokens[2].Span,
				MatchFromZone: true,
				FromZone:      triggerEventZone(TriggerEventZoneBattlefield, shared.Span{}),
				MatchToZone:   true,
				ToZone:        triggerEventZone(TriggerEventZoneGraveyard, shared.Span{}),
			},
		}
	}
	if tokenWordsEqual(tokens, "spells", "were", "cast") {
		return &TriggerEventClause{
			Kind: TriggerEventKindSpellCast,
			Span: span,
			Actor: TriggerEventActor{
				Kind: TriggerEventActorPlayer,
				Span: tokens[0].Span,
			},
		}
	}
	return nil
}

func parseEventHistoryPlayerEvent(tokens []shared.Token) *PlayerEventTriggerClause {
	var player TriggerPlayerSelector
	var actionTokens []shared.Token
	switch {
	case len(tokens) >= 1 && equalWord(tokens[0], "you"):
		player = playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
		actionTokens = tokens[1:]
	case len(tokens) >= 2 && syntaxWordsEqual(tokens[:2], "an", "opponent"):
		player = playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens[:2]))
		actionTokens = tokens[2:]
	default:
		return nil
	}
	var action PlayerEventActionKind
	switch {
	case tokenWordsEqual(actionTokens, "gained", "life") && player.Kind == TriggerPlayerSelectorYou:
		action = PlayerEventActionGainLife
	case tokenWordsEqual(actionTokens, "lost", "life"):
		action = PlayerEventActionLoseLife
	default:
		return nil
	}
	return &PlayerEventTriggerClause{
		Span:   shared.SpanOf(tokens),
		Player: player,
		Action: PlayerEventAction{
			Kind: action,
			Span: shared.SpanOf(actionTokens),
		},
		Card:       PlayerEventCard{Kind: PlayerEventCardNone},
		Occurrence: PlayerEventOccurrence{Kind: PlayerEventOccurrenceAny},
	}
}
