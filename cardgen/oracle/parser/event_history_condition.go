package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func emitEventHistoryConditions(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		tokens := eventHistorySemanticTokens(ability.Tokens, ability.Reminders, ability.Quoted)
		if conditions := parseEventHistoryConditions(tokens, ability.Atoms); len(conditions) > 0 {
			ability.EventHistoryConditions = conditions
		}
		if ability.Modal == nil {
			continue
		}
		for j := range ability.Modal.Options {
			mode := &ability.Modal.Options[j]
			tokens := eventHistorySemanticTokens(mode.Tokens, mode.Reminders, mode.Quoted)
			if conditions := parseEventHistoryConditions(tokens, mode.Atoms); len(conditions) > 0 {
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

func parseEventHistoryConditions(tokens []shared.Token, atoms Atoms) []EventHistoryCondition {
	var conditions []EventHistoryCondition
	for i := 0; i < len(tokens); i++ {
		intro, width := conditionIntroAt(tokens, i)
		if intro != ConditionIntroIf && intro != ConditionIntroOnlyIf {
			continue
		}
		end := eventHistoryConditionEnd(tokens, i)
		if condition, ok := parseEventHistoryCondition(tokens[i:end], width, atoms); ok {
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
func parseEventHistoryCondition(tokens []shared.Token, introWidth int, atoms Atoms) (EventHistoryCondition, bool) {
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
		if clause, minCount, ok := parseEventHistoryEnteredBattlefield(event, atoms); ok {
			condition.TriggerEvent = clause
			condition.MinCount = minCount
		}
	}
	if condition.TriggerEvent == nil {
		if clause, minCount, ok := parseEventHistoryYouCastSpell(event); ok {
			condition.TriggerEvent = clause
			condition.MinCount = minCount
		}
	}
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
			if condition.Negated {
				return condition.Window.Kind == EventHistoryWindowPreviousTurn
			}
			// A positive "you've cast ... this turn" restriction counts the
			// controller's own current-turn spell casts. Only the controller-
			// scoped actor is reduced to a typed event-history clause; the
			// passive "spells were cast" form remains the negated previous-turn
			// shape handled above.
			return condition.TriggerEvent.Actor.Kind == TriggerEventActorYou &&
				condition.Window.Kind == EventHistoryWindowCurrentTurn
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
	if clause := parseEventHistoryDescended(tokens, span); clause != nil {
		return clause
	}
	return parseEventHistoryLeftBattlefield(tokens, span)
}

// eventHistoryPermanentCardTypes lists the card types whose presence makes a
// card a permanent card (CR 110.4a). A descend event-history clause matches a
// card carrying any one of them, expressed as a disjunctive RequiredTypesAny
// union.
func eventHistoryPermanentCardTypes() []TriggerCardType {
	return []TriggerCardType{
		TriggerCardTypeArtifact,
		TriggerCardTypeBattle,
		TriggerCardTypeCreature,
		TriggerCardTypeEnchantment,
		TriggerCardTypeLand,
		TriggerCardTypePlaneswalker,
	}
}

// parseEventHistoryDescended recognizes the descend event-history clause "you
// descended", reduced from "you descended this turn" after the window suffix is
// stripped (CR 701.51). A controller descended this turn if a permanent card was
// put into their graveyard from anywhere during the turn, so the clause compiles
// to a current-turn zone-change matching any nontoken permanent card moving into
// a graveyard owned by the ability's controller (Event.Player carries the moved
// card's owner). Anything other than the exact two-word controller phrase fails
// closed.
func parseEventHistoryDescended(tokens []shared.Token, span shared.Span) *TriggerEventClause {
	if !tokenWordsEqual(tokens, "you", "descended") {
		return nil
	}
	return &TriggerEventClause{
		Kind:   TriggerEventKindZoneChange,
		Span:   span,
		Player: playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span),
		Subject: TriggerEventSubject{
			Kind: TriggerEventSubjectSelection,
			Span: shared.SpanOf(tokens[:1]),
			Selection: TriggerSelection{
				RequiredTypesAny: eventHistoryPermanentCardTypes(),
				NonToken:         true,
			},
		},
		ZoneChange: TriggerEventZoneChange{
			Kind: TriggerEventZoneChangeMoved,
			Span: tokens[1].Span,
		},
		Zone: TriggerEventZoneContext{
			Span:        span,
			MatchToZone: true,
			ToZone:      triggerEventZone(TriggerEventZoneGraveyard, shared.Span{}),
		},
	}
}

// parseEventHistoryYouCastSpell recognizes a controller-scoped past-tense
// spell-cast event-history clause introduced by "you've cast", "you have cast",
// or the bare "you cast". It returns a spell-cast clause whose controller is the
// ability's controller together with the minimum number of matching casts the
// window must contain. Two body shapes are recognized: a counted plural run
// ("two or more spells") that sets the minimum count and matches any spell, and
// a singular filtered run ("a noncreature spell", "an instant or sorcery spell")
// that reuses the live spell-cast trigger's selection grammar and treats a
// single matching cast as sufficient. Anything else fails closed.
func parseEventHistoryYouCastSpell(tokens []shared.Token) (*TriggerEventClause, int, bool) {
	actorWidth, ok := eventHistoryYouCastPrefix(tokens)
	if !ok {
		return nil, 0, false
	}
	actor := TriggerEventActor{
		Kind: TriggerEventActorYou,
		Span: shared.SpanOf(tokens[:actorWidth]),
	}
	body := tokens[actorWidth:]
	if count, ok := eventHistoryCastSpellCount(body); ok {
		return &TriggerEventClause{
			Kind:           TriggerEventKindSpellCast,
			Span:           shared.SpanOf(tokens),
			Actor:          actor,
			SpellSelection: TriggerEventSpellSelection{Span: shared.SpanOf(body)},
		}, count, true
	}
	selection, ok := parseTriggerEventSpellSelection(body)
	if !ok {
		return nil, 0, false
	}
	return &TriggerEventClause{
		Kind:           TriggerEventKindSpellCast,
		Span:           shared.SpanOf(tokens),
		Actor:          actor,
		SpellSelection: selection,
	}, 0, true
}

// eventHistoryYouCastPrefix reports the token width of a controller-scoped
// past-tense cast actor ("you've cast", "you have cast", "you cast") at the
// start of an event-history clause, or false when none is present.
func eventHistoryYouCastPrefix(tokens []shared.Token) (int, bool) {
	if len(tokens) >= 2 && equalWord(tokens[0], "you've") && equalWord(tokens[1], "cast") {
		return 2, true
	}
	if len(tokens) >= 3 && equalWord(tokens[0], "you") && equalWord(tokens[1], "have") && equalWord(tokens[2], "cast") {
		return 3, true
	}
	if len(tokens) >= 2 && equalWord(tokens[0], "you") && equalWord(tokens[1], "cast") {
		return 2, true
	}
	return 0, false
}

// eventHistoryCastSpellCount recognizes a counted plural spell run ("two or more
// spells") and returns the cardinal minimum. The minimum must be at least two so
// the counted form never overlaps the singular "a spell" selection, which
// treats a single cast as sufficient.
func eventHistoryCastSpellCount(tokens []shared.Token) (int, bool) {
	if len(tokens) != 4 ||
		!equalWord(tokens[1], "or") ||
		!equalWord(tokens[2], "more") ||
		!equalWord(tokens[3], "spells") {
		return 0, false
	}
	count, ok := CardinalWordValue(tokens[0].Text)
	if !ok || count < 2 {
		return 0, false
	}
	return count, true
}

// parseEventHistoryLeftBattlefield recognizes the Revolt event-history clause
// "a permanent left the battlefield under your control" (and its creature-scoped
// variant "a creature left the battlefield under your control"). It compiles to
// a current-turn zone-change pattern matching any permanent (or any creature)
// whose controller was the ability's controller leaving the battlefield to any
// zone, exactly like the live trigger "Whenever a permanent you control leaves
// the battlefield". The controller relation is carried on the clause itself
// (TriggerEventClause.Controller), matching the live trigger's representation.
func parseEventHistoryLeftBattlefield(tokens []shared.Token, span shared.Span) *TriggerEventClause {
	rest, ok := cutTokenPrefix(tokens, "a")
	if !ok || len(rest) == 0 {
		return nil
	}
	var selection TriggerSelection
	switch {
	case tokenWordsEqual(rest[:1], "permanent"):
	case tokenWordsEqual(rest[:1], "creature"):
		selection.RequiredTypes = []TriggerCardType{TriggerCardTypeCreature}
	default:
		return nil
	}
	subject := tokens[:2]
	rest = rest[1:]
	if !tokenWordsEqual(rest, "left", "the", "battlefield", "under", "your", "control") {
		return nil
	}
	return &TriggerEventClause{
		Kind:       TriggerEventKindZoneChange,
		Span:       span,
		Controller: ControllerYou,
		Subject: TriggerEventSubject{
			Kind:      TriggerEventSubjectSelection,
			Span:      shared.SpanOf(subject),
			Selection: selection,
		},
		ZoneChange: TriggerEventZoneChange{
			Kind: TriggerEventZoneChangeMoved,
			Span: rest[0].Span,
		},
		Zone: TriggerEventZoneContext{
			Span:          shared.SpanOf(rest),
			MatchFromZone: true,
			FromZone:      triggerEventZone(TriggerEventZoneBattlefield, shared.Span{}),
		},
	}
}

// enteredBattlefieldUnderYourControlWords is the trailing verb phrase of an
// "entered the battlefield under your control" event-history clause. The subject
// before it carries the permanent selection and any count qualifier; the
// controller relation is fixed to the ability's controller by the phrase.
var enteredBattlefieldUnderYourControlWords = []string{
	"entered", "the", "battlefield", "under", "your", "control",
}

// parseEventHistoryEnteredBattlefield recognizes the enters-the-battlefield
// event-history clause "<count> <Selection> entered the battlefield under your
// control", reduced from its "this turn" form after the window suffix is
// stripped. It mirrors the live enters-the-battlefield trigger "Whenever
// <Selection> enters the battlefield under your control": the clause compiles to
// a current-turn zone-change matching any permanent of the selection entering
// the battlefield under the ability's controller. An optional "<cardinal> or
// more" prefix sets the minimum number of matching entries the window must
// contain ("two or more nonland permanents entered ..."); the singular "a"/"an"/
// "another" form treats a single matching entry as sufficient. The full subject
// grammar is shared with the live trigger through parseZoneChangeSubject, so
// selections, the self-excluding "another" qualifier, and the face-down
// restriction all carry over. Anything other than the controller-scoped phrase
// fails closed.
func parseEventHistoryEnteredBattlefield(tokens []shared.Token, atoms Atoms) (*TriggerEventClause, int, bool) {
	subjectTokens, ok := stripTokenSuffix(tokens, enteredBattlefieldUnderYourControlWords...)
	if !ok || len(subjectTokens) == 0 {
		return nil, 0, false
	}
	minCount := 0
	plural := false
	if rest, count, ok := cutEventHistoryCountPrefix(subjectTokens); ok {
		subjectTokens = rest
		minCount = count
		plural = true
	}
	subject := parseZoneChangeSubject(subjectTokens, plural, atoms, "")
	if !subject.ok || subject.controller != ControllerAny || subject.selfOrAnother ||
		subject.dealtDamageBySrc || subject.oneOrMore ||
		subject.player.Kind != TriggerPlayerSelectorUnknown {
		return nil, 0, false
	}
	span := shared.SpanOf(tokens)
	return &TriggerEventClause{
		Kind:        TriggerEventKindZoneChange,
		Span:        span,
		Controller:  ControllerYou,
		Subject:     subject.subject,
		ExcludeSelf: subject.excludeSelf,
		FaceDown:    subject.faceDown,
		ZoneChange: TriggerEventZoneChange{
			Kind: TriggerEventZoneChangeEnteredBattlefield,
			Span: span,
		},
		Zone: TriggerEventZoneContext{
			Span:        span,
			MatchToZone: true,
			ToZone:      triggerEventZone(TriggerEventZoneBattlefield, shared.Span{}),
		},
	}, minCount, true
}

// cutEventHistoryCountPrefix recognizes a leading "<cardinal> or more" count
// qualifier on a plural event-history subject ("two or more nonland permanents",
// "three or more artifacts") and returns the reduced subject tokens together with
// the cardinal minimum. The minimum must be at least two so the counted form
// never overlaps the singular "a <Selection>" subject, which treats a single
// matching event as sufficient.
func cutEventHistoryCountPrefix(tokens []shared.Token) ([]shared.Token, int, bool) {
	if len(tokens) < 4 || !equalWord(tokens[1], "or") || !equalWord(tokens[2], "more") {
		return nil, 0, false
	}
	count, ok := CardinalWordValue(tokens[0].Text)
	if !ok || count < 2 {
		return nil, 0, false
	}
	return tokens[3:], count, true
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
