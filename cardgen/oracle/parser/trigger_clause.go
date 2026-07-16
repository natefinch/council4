package parser

import (
	"slices"
	"strings"

	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func parseTriggerClause(source string, tokens []shared.Token, cardName string) *TriggerClause {
	end := triggerBodyComma(tokens, cardName)
	if end < 0 {
		end = len(tokens)
	}
	clauseTokens := tokens[:end]
	if len(clauseTokens) == 0 {
		return nil
	}
	clause := &TriggerClause{
		Span:   shared.SpanOf(clauseTokens),
		Text:   shared.SliceSpan(source, shared.SpanOf(clauseTokens)),
		Tokens: cloneTokens(clauseTokens),
		Introduction: TriggerIntroduction{
			Kind: triggerIntroductionKind(clauseTokens[0]),
			Span: clauseTokens[0].Span,
		},
	}
	if len(clauseTokens) == 1 {
		return clause
	}
	phrase := phraseFromTokens(source, clauseTokens[1:])
	clause.Event = phrase.Text
	clause.EventSpan = phrase.Span
	clause.eventTokens = phrase.Tokens
	switch clause.Introduction.Kind {
	case TriggerIntroductionAt:
		clause.PhaseStep = parsePhaseStepTriggerClause(clauseTokens[1:])
	case TriggerIntroductionWhen, TriggerIntroductionWhenever:
		clause.PlayerEvent = parsePlayerEventTriggerClause(clauseTokens[1:], clause.Introduction.Kind, cardName)
	default:
	}
	return clause
}

func triggerIntroductionKind(token shared.Token) TriggerIntroductionKind {
	switch strings.ToLower(token.Text) {
	case "when":
		return TriggerIntroductionWhen
	case "whenever":
		return TriggerIntroductionWhenever
	case "at":
		return TriggerIntroductionAt
	default:
		return TriggerIntroductionUnknown
	}
}

func parsePhaseStepTriggerClause(tokens []shared.Token) *PhaseStepTriggerClause {
	if clause, ok := parseStandaloneEndOfCombat(tokens); ok {
		return &clause
	}
	event, ok := cutSyntaxWords(tokens, "the", "beginning", "of")
	if !ok {
		return nil
	}
	for _, parse := range []func([]shared.Token) (PhaseStepTriggerClause, bool){
		parseAttachedControllerPhaseStep,
		parseTurnQualifiedPhaseStep,
		parseRelationPhaseStep,
	} {
		clause, ok := parse(event)
		if ok {
			clause.Span = shared.SpanOf(tokens)
			return &clause
		}
	}
	return nil
}

func parseStandaloneEndOfCombat(tokens []shared.Token) (PhaseStepTriggerClause, bool) {
	nameTokens := tokens
	quantifier := PhaseStepQuantifier{Kind: PhaseStepQuantifierNone}
	player := TriggerPlayerSelector{Kind: TriggerPlayerSelectorAny}
	if rest, ok := cutSyntaxWords(tokens, "the"); ok {
		nameTokens = rest
		quantifier = PhaseStepQuantifier{Kind: PhaseStepQuantifierSingle, Span: tokens[0].Span}
		player.Span = tokens[0].Span
	}
	name, ok := parsePhaseStepName(nameTokens, false)
	if !ok || name.Kind != PhaseStepNameEndOfCombat {
		return PhaseStepTriggerClause{}, false
	}
	return PhaseStepTriggerClause{
		Span:       shared.SpanOf(tokens),
		Quantifier: quantifier,
		Player:     player,
		Name:       name,
	}, true
}

func parseRelationPhaseStep(tokens []shared.Token) (PhaseStepTriggerClause, bool) {
	determiner, ok := parsePhaseStepDeterminer(tokens)
	if !ok {
		return PhaseStepTriggerClause{}, false
	}
	remainder := determiner.remainder
	next := false
	if rest, ok := cutSyntaxWords(remainder, "next"); ok {
		remainder = rest
		next = true
	}
	eachOf := determiner.quantifier.Kind == PhaseStepQuantifierEachOf
	name, ok := parsePhaseStepName(remainder, eachOf)
	first := false
	eachTurn := false
	if !ok && !next {
		// Recognize the "first <step> each turn" ordinal wording (Paradox Haze:
		// "enchanted player's first upkeep each turn"). Both the leading "first"
		// and the trailing "each turn" must be present; the step name is parsed
		// from what remains. parsePhaseStepName is tried first above so the
		// distinct "first main phase" step name still wins.
		stripped, cutFirst := cutSyntaxWords(remainder, "first")
		if cutFirst {
			if rest, cutEach := cutTokenSuffix(stripped, "each", "turn"); cutEach {
				first = true
				eachTurn = true
				name, ok = parsePhaseStepName(rest, eachOf)
				if ok && name.Kind != PhaseStepNameUpkeep {
					return PhaseStepTriggerClause{}, false
				}
			}
		}
	}
	if !ok {
		return PhaseStepTriggerClause{}, false
	}
	return PhaseStepTriggerClause{
		Quantifier: determiner.quantifier,
		Player:     determiner.player,
		Name:       name,
		Next:       next,
		First:      first,
		EachTurn:   eachTurn,
	}, true
}

func parseTurnQualifiedPhaseStep(tokens []shared.Token) (PhaseStepTriggerClause, bool) {
	on := syntaxWordIndex(tokens, "on")
	if on <= 0 {
		return PhaseStepTriggerClause{}, false
	}
	nameTokens := tokens[:on]
	article := false
	if len(nameTokens) > 0 && equalWord(nameTokens[0], "the") {
		article = true
		nameTokens = nameTokens[1:]
	}
	name, ok := parsePhaseStepName(nameTokens, false)
	if !ok ||
		name.Kind == PhaseStepNameCombat && article ||
		name.Kind == PhaseStepNameEndOfCombat && !article ||
		(name.Kind != PhaseStepNameCombat && name.Kind != PhaseStepNameEndOfCombat) {
		return PhaseStepTriggerClause{}, false
	}
	determiner, ok := parsePhaseStepDeterminer(tokens[on+1:])
	if !ok ||
		determiner.quantifier.Kind == PhaseStepQuantifierEachOf ||
		determiner.quantifier.Kind == PhaseStepQuantifierSingle &&
			determiner.player.Kind == TriggerPlayerSelectorAny ||
		!syntaxWordsEqual(determiner.remainder, "turn") {
		return PhaseStepTriggerClause{}, false
	}
	return PhaseStepTriggerClause{
		Quantifier: determiner.quantifier,
		Player:     determiner.player,
		Name:       name,
	}, true
}

func parseAttachedControllerPhaseStep(tokens []shared.Token) (PhaseStepTriggerClause, bool) {
	if len(tokens) < 6 || !equalWord(tokens[0], "the") {
		return PhaseStepTriggerClause{}, false
	}
	of := syntaxWordIndex(tokens, "of")
	if of <= 1 || of+3 >= len(tokens) || !equalWord(tokens[of+1], "enchanted") {
		return PhaseStepTriggerClause{}, false
	}
	name, ok := parsePhaseStepName(tokens[1:of], false)
	if !ok ||
		(name.Kind != PhaseStepNameUpkeep &&
			name.Kind != PhaseStepNameDrawStep &&
			name.Kind != PhaseStepNameEndStep) {
		return PhaseStepTriggerClause{}, false
	}
	subjectTokens := tokens[of+2 : len(tokens)-1]
	if len(subjectTokens) == 0 ||
		!strings.HasSuffix(strings.ToLower(subjectTokens[len(subjectTokens)-1].Text), "'s") &&
			!strings.HasSuffix(strings.ToLower(subjectTokens[len(subjectTokens)-1].Text), "’s") ||
		!equalWord(tokens[len(tokens)-1], "controller") {
		return PhaseStepTriggerClause{}, false
	}
	subject, ok := parseTriggerAttachedSubject(subjectTokens)
	if !ok {
		return PhaseStepTriggerClause{}, false
	}
	return PhaseStepTriggerClause{
		Quantifier: PhaseStepQuantifier{
			Kind: PhaseStepQuantifierSingle,
			Span: tokens[0].Span,
		},
		Player: TriggerPlayerSelector{
			Kind:            TriggerPlayerSelectorAttachedController,
			Span:            shared.SpanOf(tokens[of:]),
			AttachedSubject: subject,
		},
		Name: name,
	}, true
}

type phaseStepDeterminer struct {
	quantifier PhaseStepQuantifier
	player     TriggerPlayerSelector
	remainder  []shared.Token
}

func parsePhaseStepDeterminer(tokens []shared.Token) (phaseStepDeterminer, bool) {
	if len(tokens) == 0 {
		return phaseStepDeterminer{}, false
	}
	if rest, ok := cutSyntaxWords(tokens, "each", "of"); ok {
		parsed := parseTriggerPlayerSelector(rest)
		if !parsed.ok ||
			parsed.form != triggerPlayerSelectorPossessive ||
			parsed.player.Kind != TriggerPlayerSelectorYou {
			return phaseStepDeterminer{}, false
		}
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierEachOf, Span: shared.SpanOf(tokens[:2])},
			player:     parsed.player,
			remainder:  parsed.remainder,
		}, true
	}
	if rest, ok := cutSyntaxWords(tokens, "each"); ok && len(rest) > 0 {
		parsed := parseTriggerPlayerSelector(rest)
		if parsed.ok && parsed.form == triggerPlayerSelectorPossessive &&
			(parsed.player.Kind == TriggerPlayerSelectorAny ||
				parsed.player.Kind == TriggerPlayerSelectorOpponent) {
			return phaseStepDeterminer{
				quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierEach, Span: tokens[0].Span},
				player:     parsed.player,
				remainder:  parsed.remainder,
			}, true
		}
	}
	if parsed := parseTriggerPlayerSelector(tokens); parsed.ok &&
		parsed.form == triggerPlayerSelectorPossessive &&
		(parsed.player.Kind == TriggerPlayerSelectorYou ||
			parsed.player.Kind == TriggerPlayerSelectorSourceController ||
			parsed.player.Kind == TriggerPlayerSelectorEnchantedPlayer ||
			parsed.player.Kind == TriggerPlayerSelectorMonarch) {
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierSingle, Span: parsed.player.Span},
			player:     parsed.player,
			remainder:  parsed.remainder,
		}, true
	}
	switch {
	case equalWord(tokens[0], "the"):
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierSingle, Span: tokens[0].Span},
			player:     TriggerPlayerSelector{Kind: TriggerPlayerSelectorAny, Span: tokens[0].Span},
			remainder:  tokens[1:],
		}, true
	case equalWord(tokens[0], "each"):
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierEach, Span: tokens[0].Span},
			player:     TriggerPlayerSelector{Kind: TriggerPlayerSelectorAny, Span: tokens[0].Span},
			remainder:  tokens[1:],
		}, true
	default:
		return phaseStepDeterminer{}, false
	}
}

type triggerPlayerSelectorForm string

const (
	triggerPlayerSelectorFormUnknown triggerPlayerSelectorForm = ""
	triggerPlayerSelectorSubject     triggerPlayerSelectorForm = "triggerPlayerSelectorSubject"
	triggerPlayerSelectorPossessive  triggerPlayerSelectorForm = "triggerPlayerSelectorPossessive"
)

type triggerPlayerSelectorParse struct {
	player    TriggerPlayerSelector
	remainder []shared.Token
	form      triggerPlayerSelectorForm
	ok        bool
}

func parseTriggerPlayerSelector(tokens []shared.Token) triggerPlayerSelectorParse {
	switch {
	case len(tokens) >= 2 && equalWord(tokens[0], "an") && equalWord(tokens[1], "opponent"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorOpponent, Span: shared.SpanOf(tokens[:2])},
			remainder: tokens[2:],
			form:      triggerPlayerSelectorSubject,
			ok:        true,
		}
	case len(tokens) >= 2 && equalWord(tokens[0], "a") && equalWord(tokens[1], "player"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorAny, Span: shared.SpanOf(tokens[:2])},
			remainder: tokens[2:],
			form:      triggerPlayerSelectorSubject,
			ok:        true,
		}
	case len(tokens) >= 2 && equalWord(tokens[0], "its") && equalWord(tokens[1], "controller's"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorSourceController, Span: shared.SpanOf(tokens[:2])},
			remainder: tokens[2:],
			form:      triggerPlayerSelectorPossessive,
			ok:        true,
		}
	case len(tokens) >= 1 && equalWord(tokens[0], "you"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorYou, Span: tokens[0].Span},
			remainder: tokens[1:],
			form:      triggerPlayerSelectorSubject,
			ok:        true,
		}
	case len(tokens) >= 1 && equalWord(tokens[0], "your"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorYou, Span: tokens[0].Span},
			remainder: tokens[1:],
			form:      triggerPlayerSelectorPossessive,
			ok:        true,
		}
	case len(tokens) >= 1 && equalWord(tokens[0], "player's"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorAny, Span: tokens[0].Span},
			remainder: tokens[1:],
			form:      triggerPlayerSelectorPossessive,
			ok:        true,
		}
	case len(tokens) >= 2 && equalWord(tokens[0], "enchanted") && equalWord(tokens[1], "player's"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorEnchantedPlayer, Span: shared.SpanOf(tokens[:2])},
			remainder: tokens[2:],
			form:      triggerPlayerSelectorPossessive,
			ok:        true,
		}
	case len(tokens) >= 1 && equalWord(tokens[0], "opponent's"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorOpponent, Span: tokens[0].Span},
			remainder: tokens[1:],
			form:      triggerPlayerSelectorPossessive,
			ok:        true,
		}
	case len(tokens) >= 2 && equalWord(tokens[0], "the") && equalWord(tokens[1], "monarch's"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorMonarch, Span: shared.SpanOf(tokens[:2])},
			remainder: tokens[2:],
			form:      triggerPlayerSelectorPossessive,
			ok:        true,
		}
	default:
		return triggerPlayerSelectorParse{}
	}
}

func parsePlayerEventTriggerClause(tokens []shared.Token, introduction TriggerIntroductionKind, cardName string) *PlayerEventTriggerClause {
	parsedPlayer := parseTriggerPlayerSelector(tokens)
	if !parsedPlayer.ok || parsedPlayer.form != triggerPlayerSelectorSubject {
		return nil
	}
	action, rest, ok := parsePlayerEventAction(parsedPlayer.remainder, parsedPlayer.player.Kind)
	if !ok {
		return nil
	}
	modifiers, ok := parsePlayerEventModifiers(rest, action.Kind, parsedPlayer.player.Kind, cardName)
	if !ok ||
		(modifiers.occurrence.Kind == PlayerEventOccurrenceAny &&
			introduction != TriggerIntroductionWhenever &&
			modifiers.card.Kind != PlayerEventCardThis) {
		return nil
	}
	if action.Kind == PlayerEventActionCast &&
		modifiers.turnRelation != TriggerCastTurnRelationEventPlayerTurn {
		return nil
	}
	return &PlayerEventTriggerClause{
		Span:         shared.SpanOf(tokens),
		Player:       parsedPlayer.player,
		Action:       action,
		Card:         modifiers.card,
		Occurrence:   modifiers.occurrence,
		TurnRelation: modifiers.turnRelation,
	}
}

func parsePlayerEventAction(
	tokens []shared.Token,
	player TriggerPlayerSelectorKind,
) (PlayerEventAction, []shared.Token, bool) {
	if len(tokens) == 0 {
		return PlayerEventAction{}, nil, false
	}
	verbMatches := func(token shared.Token, secondPerson, thirdPerson string) bool {
		if player == TriggerPlayerSelectorYou {
			return equalWord(token, secondPerson)
		}
		return equalWord(token, thirdPerson)
	}
	if len(tokens) >= 3 &&
		verbMatches(tokens[0], "cycle", "cycles") &&
		equalWord(tokens[1], "or") &&
		verbMatches(tokens[2], "discard", "discards") {
		return PlayerEventAction{
			Kind: PlayerEventActionCycleOrDiscard,
			Span: shared.SpanOf(tokens[:3]),
		}, tokens[3:], true
	}
	if len(tokens) >= 3 &&
		verbMatches(tokens[0], "search", "searches") &&
		possessiveMatches(tokens[1], player) &&
		equalWord(tokens[2], "library") {
		return PlayerEventAction{
			Kind: PlayerEventActionSearchLibrary,
			Span: shared.SpanOf(tokens[:3]),
		}, tokens[3:], true
	}
	if len(tokens) >= 3 &&
		verbMatches(tokens[0], "commit", "commits") &&
		equalWord(tokens[1], "a") &&
		equalWord(tokens[2], "crime") {
		return PlayerEventAction{
			Kind: PlayerEventActionCommitCrime,
			Span: shared.SpanOf(tokens[:3]),
		}, tokens[3:], true
	}
	if len(tokens) >= 3 &&
		verbMatches(tokens[0], "become", "becomes") &&
		equalWord(tokens[1], "the") &&
		equalWord(tokens[2], "monarch") {
		return PlayerEventAction{
			Kind: PlayerEventActionBecomeMonarch,
			Span: shared.SpanOf(tokens[:3]),
		}, tokens[3:], true
	}
	if verbMatches(tokens[0], "play", "plays") {
		return PlayerEventAction{
			Kind: PlayerEventActionPlay,
			Span: shared.SpanOf(tokens[:1]),
		}, tokens[1:], true
	}
	if verbMatches(tokens[0], "cast", "casts") {
		return PlayerEventAction{
			Kind: PlayerEventActionCast,
			Span: shared.SpanOf(tokens[:1]),
		}, tokens[1:], true
	}
	for _, form := range []struct {
		kind       PlayerEventActionKind
		second     string
		third      string
		lifeObject bool
	}{
		{kind: PlayerEventActionDraw, second: "draw", third: "draws"},
		{kind: PlayerEventActionDiscard, second: "discard", third: "discards"},
		{kind: PlayerEventActionCycle, second: "cycle", third: "cycles"},
		{kind: PlayerEventActionScry, second: "scry", third: "scries"},
		{kind: PlayerEventActionSurveil, second: "surveil", third: "surveils"},
		{kind: PlayerEventActionGainLife, second: "gain", third: "gains", lifeObject: true},
		{kind: PlayerEventActionLoseLife, second: "lose", third: "loses", lifeObject: true},
	} {
		if !verbMatches(tokens[0], form.second, form.third) {
			continue
		}
		end := 1
		if form.lifeObject {
			if len(tokens) < 2 || !equalWord(tokens[1], "life") {
				return PlayerEventAction{}, nil, false
			}
			end = 2
		}
		return PlayerEventAction{Kind: form.kind, Span: shared.SpanOf(tokens[:end])}, tokens[end:], true
	}
	return PlayerEventAction{}, nil, false
}

// possessiveMatches reports whether token is the possessive determiner used for
// the selector's grammatical person: "your" for the controller-scoped "you"
// selector and "their" for the third-person "a player"/"an opponent" selectors.
func possessiveMatches(token shared.Token, player TriggerPlayerSelectorKind) bool {
	if player == TriggerPlayerSelectorYou {
		return equalWord(token, "your")
	}
	return equalWord(token, "their")
}

// playerEventModifiers bundles the optional modifiers that follow a player-event
// action: the card object, the turn-relative occurrence, and the active-turn
// timing relation folded in by a "during your turn" phrase.
type playerEventModifiers struct {
	card         PlayerEventCard
	occurrence   PlayerEventOccurrence
	turnRelation TriggerCastTurnRelation
}

func parsePlayerEventModifiers(
	tokens []shared.Token,
	action PlayerEventActionKind,
	player TriggerPlayerSelectorKind,
	cardName string,
) (playerEventModifiers, bool) {
	card := PlayerEventCard{Kind: PlayerEventCardNone}
	occurrence := PlayerEventOccurrence{Kind: PlayerEventOccurrenceAny}
	turnRelation := TriggerCastTurnRelationNone
	rest := tokens
	if playerEventActionHasCard(action) {
		parsed := parsePlayerEventCard(rest, action, player, cardName)
		if !parsed.ok {
			return playerEventModifiers{}, false
		}
		card = parsed.card
		occurrence = parsed.occurrence
		turnRelation = parsed.turnRelation
		rest = parsed.remainder
	}
	if next, relation, ok := cutPlayerEventFirstEachTurn(rest, player); ok {
		if occurrence.Kind != PlayerEventOccurrenceAny || !playerEventFirstEachTurnAllowed(action, player) {
			return playerEventModifiers{}, false
		}
		if relation != TriggerCastTurnRelationNone && !playerEventTurnRelationAllowed(action) {
			return playerEventModifiers{}, false
		}
		occurrence = PlayerEventOccurrence{
			Kind:    PlayerEventOccurrenceFirstEachTurn,
			Span:    shared.SpanOf(rest),
			Ordinal: 1,
		}
		turnRelation = relation
		rest = next
	}
	if action == PlayerEventActionDraw {
		if next, ok := cutExceptFirstInDrawStep(rest, player); ok {
			if occurrence.Kind != PlayerEventOccurrenceAny {
				return playerEventModifiers{}, false
			}
			occurrence = PlayerEventOccurrence{
				Kind: PlayerEventOccurrenceExceptFirstInDrawStep,
				Span: shared.SpanOf(rest),
			}
			rest = next
		}
	}
	if turnRelation == TriggerCastTurnRelationNone {
		if next, relation, ok := cutPlayerEventTurnRelation(rest); ok {
			if !playerEventTurnRelationAllowed(action) {
				return playerEventModifiers{}, false
			}
			turnRelation = relation
			rest = next
		}
	}
	if len(rest) != 0 {
		return playerEventModifiers{}, false
	}
	return playerEventModifiers{card: card, occurrence: occurrence, turnRelation: turnRelation}, true
}

// cutPlayerEventFirstEachTurn consumes the "for the first time" occurrence
// qualifier in either the plain "for the first time each turn" form or the
// active-turn "for the first time during each of your turns" / "... their turns"
// form, reporting the controller-relative turn relation the timing phrase folds
// in (none for the plain form).
func cutPlayerEventFirstEachTurn(
	tokens []shared.Token,
	player TriggerPlayerSelectorKind,
) ([]shared.Token, TriggerCastTurnRelation, bool) {
	rest, ok := cutSyntaxWords(tokens, "for", "the", "first", "time")
	if !ok {
		return tokens, TriggerCastTurnRelationNone, false
	}
	if next, ok := cutSyntaxWords(rest, "each", "turn"); ok {
		return next, TriggerCastTurnRelationNone, true
	}
	if next, ok := cutSyntaxWords(rest, "during", "each", "of"); ok {
		if len(next) >= 2 && possessiveMatches(next[0], player) && equalWord(next[1], "turns") {
			return next[2:], playerEventTurnRelationFor(player), true
		}
	}
	return tokens, TriggerCastTurnRelationNone, false
}

// cutPlayerEventTurnRelation strips a trailing "during your turn" / "during an
// opponent's turn" timing phrase, reporting the controller-relative turn the
// triggering event must occur on.
func cutPlayerEventTurnRelation(tokens []shared.Token) ([]shared.Token, TriggerCastTurnRelation, bool) {
	if next, ok := cutSyntaxWords(tokens, "during", "your", "turn"); ok {
		return next, TriggerCastTurnRelationYourTurn, true
	}
	if next, ok := cutSyntaxWords(tokens, "during", "an", "opponent's", "turn"); ok {
		return next, TriggerCastTurnRelationNotYourTurn, true
	}
	return tokens, TriggerCastTurnRelationNone, false
}

// playerEventTurnRelationFor reports the controller-relative turn restriction for
// a "during each of <player>'s turns" phrase: the controller's own turn for the
// "you" selector and a turn that isn't the controller's for an opponent.
func playerEventTurnRelationFor(player TriggerPlayerSelectorKind) TriggerCastTurnRelation {
	if player == TriggerPlayerSelectorYou {
		return TriggerCastTurnRelationYourTurn
	}
	return TriggerCastTurnRelationNotYourTurn
}

// playerEventTurnRelationAllowed restricts the active-turn timing qualifier to
// the life-change events, where "during your turn" / "during each of their
// turns" appears in printed templating (CR 603.2).
func playerEventTurnRelationAllowed(action PlayerEventActionKind) bool {
	switch action {
	case PlayerEventActionGainLife, PlayerEventActionLoseLife:
		return true
	default:
		return false
	}
}

// cutExceptFirstInDrawStep consumes the "except the first one they draw in each
// of their draw steps" suffix (using "you"/"your" for the controller-scoped
// selector), which restricts a card-draw trigger to draws other than the first
// of each draw step (Orcish Bowmasters).
func cutExceptFirstInDrawStep(
	tokens []shared.Token,
	player TriggerPlayerSelectorKind,
) ([]shared.Token, bool) {
	rest, ok := cutSyntaxWords(tokens, "except", "the", "first", "one")
	if !ok {
		return nil, false
	}
	if len(rest) == 0 || !subjectPronounMatches(rest[0], player) {
		return nil, false
	}
	rest, ok = cutSyntaxWords(rest[1:], "draw", "in", "each", "of")
	if !ok {
		return nil, false
	}
	if len(rest) == 0 || !possessiveMatches(rest[0], player) {
		return nil, false
	}
	return cutSyntaxWords(rest[1:], "draw", "steps")
}

// subjectPronounMatches reports whether token is the subject pronoun used for
// the selector's grammatical person: "you" for the controller-scoped "you"
// selector and "they" for the third-person "a player"/"an opponent" selectors.
func subjectPronounMatches(token shared.Token, player TriggerPlayerSelectorKind) bool {
	if player == TriggerPlayerSelectorYou {
		return equalWord(token, "you")
	}
	return equalWord(token, "they")
}

type playerEventCardParse struct {
	card         PlayerEventCard
	occurrence   PlayerEventOccurrence
	turnRelation TriggerCastTurnRelation
	remainder    []shared.Token
	ok           bool
}

func parsePlayerEventCard(
	tokens []shared.Token,
	action PlayerEventActionKind,
	player TriggerPlayerSelectorKind,
	cardName string,
) playerEventCardParse {
	occurrence := PlayerEventOccurrence{Kind: PlayerEventOccurrenceAny}
	if action == PlayerEventActionPlay {
		if rest, ok := cutPlayerEventExiledWithSource(tokens, cardName); ok {
			return playerEventCardParse{
				card: PlayerEventCard{
					Kind: PlayerEventCardExiledWithSource,
					Span: shared.SpanOf(tokens[:len(tokens)-len(rest)]),
				},
				occurrence: occurrence,
				remainder:  rest,
				ok:         true,
			}
		}
		if rest, ok := cutSyntaxWords(tokens, "a", "land"); ok {
			return playerEventCardParse{
				card: PlayerEventCard{
					Kind: PlayerEventCardLand,
					Span: shared.SpanOf(tokens[:2]),
				},
				occurrence: occurrence,
				remainder:  rest,
				ok:         true,
			}
		}
		return playerEventCardParse{}
	}
	if rest, filter, ok := cutPlayerEventCardNoun(tokens, action, "a", "card"); ok {
		return playerEventCardParse{
			card: PlayerEventCard{
				Kind:                PlayerEventCardSingle,
				Span:                shared.SpanOf(tokens[:len(tokens)-len(rest)]),
				RequiredTypes:       filter.required,
				ExcludedTypes:       filter.excluded,
				RequiredTypesAny:    filter.requiredAny,
				RequiredSubtypesAny: filter.subtypesAny,
			},
			occurrence: occurrence,
			remainder:  rest,
			ok:         true,
		}
	}
	if action == PlayerEventActionDiscard {
		if rest, filter, ok := cutPlayerEventCardNoun(tokens, action, "one", "or", "more", "cards"); ok {
			return playerEventCardParse{
				card: PlayerEventCard{
					Kind:                PlayerEventCardOneOrMore,
					Span:                shared.SpanOf(tokens[:len(tokens)-len(rest)]),
					RequiredTypes:       filter.required,
					ExcludedTypes:       filter.excluded,
					RequiredTypesAny:    filter.requiredAny,
					RequiredSubtypesAny: filter.subtypesAny,
				},
				occurrence: occurrence,
				remainder:  rest,
				ok:         true,
			}
		}
	}
	if rest, ok := cutSyntaxWords(tokens, "another", "card"); ok &&
		(action == PlayerEventActionDiscard ||
			action == PlayerEventActionCycle ||
			action == PlayerEventActionCycleOrDiscard) {
		return playerEventCardParse{
			card:       PlayerEventCard{Kind: PlayerEventCardAnother, Span: shared.SpanOf(tokens[:2])},
			occurrence: occurrence,
			remainder:  rest,
			ok:         true,
		}
	}
	if rest, ok := cutSyntaxWords(tokens, "this", "card"); ok &&
		(action == PlayerEventActionCycle || action == PlayerEventActionCycleOrDiscard) {
		return playerEventCardParse{
			card:       PlayerEventCard{Kind: PlayerEventCardThis, Span: shared.SpanOf(tokens[:2])},
			occurrence: occurrence,
			remainder:  rest,
			ok:         true,
		}
	}
	if (action != PlayerEventActionDraw && action != PlayerEventActionCast) || len(tokens) < 3 {
		return playerEventCardParse{}
	}
	possessive := "their"
	if player == TriggerPlayerSelectorYou {
		possessive = "your"
	}
	noun := "card"
	if action == PlayerEventActionCast {
		noun = "spell"
	}
	if !equalWord(tokens[0], possessive) ||
		!equalWord(tokens[2], noun) {
		return playerEventCardParse{}
	}
	ordinal, ok := parsePlayerEventOrdinal(tokens[1])
	if !ok {
		return playerEventCardParse{}
	}
	remainder := tokens[3:]
	turnRelation := TriggerCastTurnRelationNone
	var occurrenceEnd int
	if next, ok := cutSyntaxWords(remainder, "each", "turn"); ok {
		remainder = next
		occurrenceEnd = 5
	} else {
		relation := TriggerCastTurnRelationEventPlayerTurn
		if player == TriggerPlayerSelectorYou {
			relation = TriggerCastTurnRelationYourTurn
		}
		next, parsedRelation, ok := cutPlayerEventTurnRelationForPlayer(remainder, player, relation)
		if !ok {
			return playerEventCardParse{}
		}
		remainder = next
		turnRelation = parsedRelation
		occurrenceEnd = len(tokens) - len(remainder)
	}
	return playerEventCardParse{
		card: PlayerEventCard{
			Kind: PlayerEventCardSingle,
			Span: shared.SpanOf(tokens[:3]),
		},
		occurrence: PlayerEventOccurrence{
			Kind:    PlayerEventOccurrenceOrdinalEachTurn,
			Span:    shared.SpanOf(tokens[1:occurrenceEnd]),
			Ordinal: ordinal,
		},
		turnRelation: turnRelation,
		remainder:    remainder,
		ok:           true,
	}
}

func cutPlayerEventTurnRelationForPlayer(
	tokens []shared.Token,
	player TriggerPlayerSelectorKind,
	relation TriggerCastTurnRelation,
) ([]shared.Token, TriggerCastTurnRelation, bool) {
	possessive := "their"
	if player == TriggerPlayerSelectorYou {
		possessive = "your"
	}
	next, ok := cutSyntaxWords(tokens, "during", possessive, "turn")
	if !ok {
		return tokens, TriggerCastTurnRelationNone, false
	}
	return next, relation, true
}

// cutPlayerEventExiledWithSource recognizes the self-referential player-event
// card object "a card exiled with <this permanent>" (Prowl, Stoic Strategist),
// whose trailing name is the ability's own source. It returns the tokens after
// the object, or false when the object is absent or names a different
// permanent, so an unrelated "plays ..." trigger stays unrecognized.
func cutPlayerEventExiledWithSource(tokens []shared.Token, cardName string) ([]shared.Token, bool) {
	after, ok := cutSyntaxWords(tokens, "a", "card", "exiled", "with")
	if !ok || len(after) == 0 {
		return nil, false
	}
	for _, span := range collectSelfNameSpans(after, cardName, false) {
		if span.Start.Offset != after[0].Span.Start.Offset {
			continue
		}
		end := 0
		for end < len(after) && after[end].Span.Start.Offset < span.End.Offset {
			end++
		}
		return after[end:], true
	}
	return nil, false
}

// playerEventCardFilter holds the card-type filter parsed from a player-event
// card object, such as "a creature card" or "a noncreature, nonland card".
type playerEventCardFilter struct {
	required    []TriggerCardType
	excluded    []TriggerCardType
	requiredAny []TriggerCardType
	subtypesAny []TriggerSubtype
}

// cutPlayerEventCardNoun matches a card object that opens with the prefix words
// and closes with the noun, allowing an optional card-type filter such as "a
// creature card" or "discard a noncreature, nonland card" in between. The type
// filter is only recognized for discard, where card-type-filtered discard
// triggers occur (CR 603.2).
func cutPlayerEventCardNoun(
	tokens []shared.Token,
	action PlayerEventActionKind,
	words ...string,
) (rest []shared.Token, filter playerEventCardFilter, ok bool) {
	prefix := words[:len(words)-1]
	noun := words[len(words)-1]
	after, ok := cutPlayerEventCardPrefix(tokens, prefix)
	if !ok {
		return nil, playerEventCardFilter{}, false
	}
	if rest, ok := cutSyntaxWords(after, noun); ok {
		return rest, playerEventCardFilter{}, true
	}
	if action != PlayerEventActionDiscard {
		return nil, playerEventCardFilter{}, false
	}
	nounIndex := syntaxWordIndex(after, noun)
	if nounIndex <= 0 {
		return nil, playerEventCardFilter{}, false
	}
	filter, ok = parsePlayerEventCardTypes(after[:nounIndex])
	if !ok {
		return nil, playerEventCardFilter{}, false
	}
	return after[nounIndex+1:], filter, true
}

// cutPlayerEventCardPrefix consumes a card object's leading prefix words. When
// the prefix is the singular article "a", it also accepts "an" so a filtered
// card object such as "an artifact card" or "an Island, Pirate, or Vehicle
// card" matches the same way the unfiltered "a card" form does.
func cutPlayerEventCardPrefix(tokens []shared.Token, prefix []string) ([]shared.Token, bool) {
	if len(prefix) == 1 && prefix[0] == "a" {
		if len(tokens) == 0 || (!equalWord(tokens[0], "a") && !equalWord(tokens[0], "an")) {
			return nil, false
		}
		return tokens[1:], true
	}
	return cutSyntaxWords(tokens, prefix...)
}

// parsePlayerEventCardTypes reads one or more comma-separated card-type words,
// each optionally negated with the "non" prefix, into required and excluded
// type filters. A disjunctive union joined by "or" ("an Island, Pirate, or
// Vehicle card") instead lowers to an any-of subtype or card-type union. It
// fails closed on any unrecognized or duplicated word.
func parsePlayerEventCardTypes(tokens []shared.Token) (playerEventCardFilter, bool) {
	if len(tokens) == 0 {
		return playerEventCardFilter{}, false
	}
	if filter, ok := parsePlayerEventCardUnion(tokens); ok {
		return filter, true
	}
	var filter playerEventCardFilter
	for _, token := range tokens {
		if token.Kind == shared.Comma {
			continue
		}
		word := strings.ToLower(token.Text)
		if rest, negated := strings.CutPrefix(word, "non"); negated {
			cardType, typeOK := triggerCardType(rest)
			if !typeOK || cardType == TriggerCardTypeUnknown || slices.Contains(filter.excluded, cardType) {
				return playerEventCardFilter{}, false
			}
			filter.excluded = append(filter.excluded, cardType)
			continue
		}
		cardType, typeOK := triggerCardType(word)
		if !typeOK || cardType == TriggerCardTypeUnknown || slices.Contains(filter.required, cardType) {
			return playerEventCardFilter{}, false
		}
		filter.required = append(filter.required, cardType)
	}
	if len(filter.required) == 0 && len(filter.excluded) == 0 {
		return playerEventCardFilter{}, false
	}
	return filter, true
}

// parsePlayerEventCardUnion recognizes a disjunctive card-object union joined by
// "or" with optional Oxford comma, such as "an Island, Pirate, or Vehicle card"
// or "an artifact or creature card". Every member must resolve to the same
// dimension: all subtypes lower to an any-of subtype union and all card types
// to an any-of card-type union. A mixed union fails closed because the runtime
// selection conjoins the two dimensions rather than disjoining them.
func parsePlayerEventCardUnion(tokens []shared.Token) (playerEventCardFilter, bool) {
	members, hasOr, ok := splitPlayerEventCardUnion(tokens)
	if !ok || !hasOr || len(members) < 2 {
		return playerEventCardFilter{}, false
	}
	var subtypes []TriggerSubtype
	var cardTypes []TriggerCardType
	for _, member := range members {
		if sub, ok := recognizeSubtypePhrase(member); ok && looksLikeTriggerSubtype(member) {
			if slices.Contains(subtypes, sub) {
				return playerEventCardFilter{}, false
			}
			subtypes = append(subtypes, sub)
			continue
		}
		if cardType, ok := triggerCardType(member); ok && cardType != TriggerCardTypeUnknown {
			if slices.Contains(cardTypes, cardType) {
				return playerEventCardFilter{}, false
			}
			cardTypes = append(cardTypes, cardType)
			continue
		}
		return playerEventCardFilter{}, false
	}
	if len(subtypes) > 0 && len(cardTypes) > 0 {
		return playerEventCardFilter{}, false
	}
	if len(subtypes) > 0 {
		return playerEventCardFilter{subtypesAny: subtypes}, true
	}
	return playerEventCardFilter{requiredAny: cardTypes}, true
}

// splitPlayerEventCardUnion splits a card-object filter into its comma- and
// "or"-separated members, lowercasing and joining each member's words. It
// reports whether an "or" connector was present and fails closed on any
// non-word, non-comma token.
func splitPlayerEventCardUnion(tokens []shared.Token) (members []string, hasOr, ok bool) {
	var current []string
	flush := func() {
		if len(current) > 0 {
			members = append(members, strings.Join(current, " "))
			current = nil
		}
	}
	for _, token := range tokens {
		if token.Kind == shared.Comma {
			flush()
			continue
		}
		if token.Kind != shared.Word {
			return nil, false, false
		}
		word := strings.ToLower(token.Text)
		if word == "or" {
			hasOr = true
			flush()
			continue
		}
		current = append(current, word)
	}
	flush()
	return members, hasOr, true
}

func playerEventActionHasCard(action PlayerEventActionKind) bool {
	switch action {
	case PlayerEventActionDraw,
		PlayerEventActionDiscard,
		PlayerEventActionCycle,
		PlayerEventActionCycleOrDiscard,
		PlayerEventActionCast,
		PlayerEventActionPlay:
		return true
	default:
		return false
	}
}

func playerEventFirstEachTurnAllowed(action PlayerEventActionKind, player TriggerPlayerSelectorKind) bool {
	switch action {
	case PlayerEventActionDraw,
		PlayerEventActionScry,
		PlayerEventActionSurveil:
		return true
	case PlayerEventActionDiscard, PlayerEventActionCycle:
		return player == TriggerPlayerSelectorYou
	case PlayerEventActionGainLife, PlayerEventActionLoseLife:
		return player != TriggerPlayerSelectorAny
	default:
		return false
	}
}

func parsePlayerEventOrdinal(token shared.Token) (int, bool) {
	switch strings.ToLower(token.Text) {
	case "first":
		return 1, true
	case "second":
		return 2, true
	case "third":
		return 3, true
	case "fourth":
		return 4, true
	case "fifth":
		return 5, true
	default:
		return 0, false
	}
}

type phaseStepNameForm struct {
	kind   PhaseStepNameKind
	plural bool
	words  []string
}

var phaseStepNameForms = []phaseStepNameForm{
	{kind: PhaseStepNameUpkeep, words: []string{"upkeep"}},
	{kind: PhaseStepNameUpkeep, plural: true, words: []string{"upkeeps"}},
	{kind: PhaseStepNameDrawStep, words: []string{"draw", "step"}},
	{kind: PhaseStepNameDrawStep, plural: true, words: []string{"draw", "steps"}},
	{kind: PhaseStepNameEndStep, words: []string{"end", "step"}},
	{kind: PhaseStepNameEndStep, plural: true, words: []string{"end", "steps"}},
	{kind: PhaseStepNameCombat, words: []string{"combat"}},
	{kind: PhaseStepNameCombat, plural: true, words: []string{"combats"}},
	{kind: PhaseStepNameCombatStep, words: []string{"combat", "step"}},
	{kind: PhaseStepNameCombatStep, plural: true, words: []string{"combat", "steps"}},
	{kind: PhaseStepNameEndOfCombat, words: []string{"end", "of", "combat"}},
	{kind: PhaseStepNameEndOfCombatStep, words: []string{"end", "of", "combat", "step"}},
	{kind: PhaseStepNameEndOfCombatStep, plural: true, words: []string{"end", "of", "combat", "steps"}},
	{kind: PhaseStepNamePrecombatMainPhase, words: []string{"precombat", "main", "phase"}},
	{kind: PhaseStepNamePrecombatMainPhase, plural: true, words: []string{"precombat", "main", "phases"}},
	{kind: PhaseStepNamePostcombatMainPhase, words: []string{"postcombat", "main", "phase"}},
	{kind: PhaseStepNamePostcombatMainPhase, plural: true, words: []string{"postcombat", "main", "phases"}},
	{kind: PhaseStepNameFirstMainPhase, words: []string{"first", "main", "phase"}},
	{kind: PhaseStepNameFirstMainPhase, plural: true, words: []string{"first", "main", "phases"}},
	{kind: PhaseStepNameSecondMainPhase, words: []string{"second", "main", "phase"}},
	{kind: PhaseStepNameSecondMainPhase, plural: true, words: []string{"second", "main", "phases"}},
}

func parsePhaseStepName(tokens []shared.Token, plural bool) (PhaseStepName, bool) {
	for _, form := range phaseStepNameForms {
		if form.plural == plural && syntaxWordsEqual(tokens, form.words...) {
			return PhaseStepName{Kind: form.kind, Span: shared.SpanOf(tokens)}, true
		}
	}
	return PhaseStepName{}, false
}

func parseTriggerAttachedSubject(tokens []shared.Token) (TriggerAttachedSubject, bool) {
	subjectTokens := cloneTokens(tokens)
	last := &subjectTokens[len(subjectTokens)-1]
	last.Text = strings.TrimSuffix(strings.TrimSuffix(last.Text, "'s"), "’s")
	if last.Text == "" {
		return TriggerAttachedSubject{}, false
	}
	selection, ok := parseTriggerSelection(subjectTokens)
	if !ok {
		return TriggerAttachedSubject{}, false
	}
	return TriggerAttachedSubject{
		Span:      shared.SpanOf(tokens),
		Selection: selection,
	}, true
}

func cutSyntaxWords(tokens []shared.Token, words ...string) ([]shared.Token, bool) {
	if len(tokens) < len(words) || !syntaxWordsEqual(tokens[:len(words)], words...) {
		return nil, false
	}
	return tokens[len(words):], true
}

func syntaxWordsEqual(tokens []shared.Token, words ...string) bool {
	if len(tokens) != len(words) {
		return false
	}
	for i, word := range words {
		if !equalWord(tokens[i], word) {
			return false
		}
	}
	return true
}

func syntaxWordIndex(tokens []shared.Token, word string) int {
	for i, token := range tokens {
		if equalWord(token, word) {
			return i
		}
	}
	return -1
}

func equalWord(token shared.Token, word string) bool {
	return token.Kind == shared.Word && strings.EqualFold(token.Text, word)
}

func triggerBodyComma(tokens []shared.Token, cardName string) int {
	// The legendary pre-"of" short name carries no comma, so it never changes
	// which comma ends the trigger clause; the non-legendary name spans suffice.
	selfNameSpans := collectSelfNameSpans(tokens, cardName, false)
	comma := shared.TopLevelIndex(tokens, shared.Comma)
	for comma > 0 {
		if end, ok := spellListComma(tokens, comma); ok {
			next := shared.TopLevelIndex(tokens[end:], shared.Comma)
			if next < 0 {
				return -1
			}
			comma = end + next
			continue
		}
		if end, ok := selfNameComma(tokens, comma, selfNameSpans); ok {
			next := shared.TopLevelIndex(tokens[end:], shared.Comma)
			if next < 0 {
				return -1
			}
			comma = end + next
			continue
		}
		break
	}
	return comma
}

// selfNameComma reports whether the comma at the given index sits inside an
// occurrence of the card's own name, such as the internal comma of a legendary
// name like "Etali, Primal Storm", rather than separating the trigger condition
// from its body. When it does, it returns the index of the first token past the
// name so the caller resumes scanning for the real body comma. A comma counts as
// interior only when the matched name span extends in source on both sides of
// it, so a card whose name ends at the comma cannot swallow the body separator.
func selfNameComma(tokens []shared.Token, comma int, selfNameSpans []shared.Span) (int, bool) {
	commaSpan := tokens[comma].Span
	for _, span := range selfNameSpans {
		if span.Start.Offset >= commaSpan.Start.Offset || span.End.Offset <= commaSpan.End.Offset {
			continue
		}
		end := comma + 1
		for end < len(tokens) && tokens[end].Span.Start.Offset < span.End.Offset {
			end++
		}
		return end, true
	}
	return 0, false
}

// spellListComma reports whether the comma at the given index sits inside a
// homogeneous comma-separated spell-type/subtype list such as "noncreature,
// nonland card" or "Aura, Equipment, or Vehicle spell", rather than separating
// the trigger condition from its body. When it does, it returns the index just
// past the closing "spell"/"card" noun so the caller can resume scanning for the
// real body comma. Every list noun must share a category (all card types or all
// subtypes); a mixed list such as "instant, sorcery, or Wizard spell" is not
// expressible as a single union, so it fails closed and keeps the legacy first-
// item split. The category and terminator checks also keep effect text like
// "draw a card, counter target spell" from being mistaken for a list.
func spellListComma(tokens []shared.Token, comma int) (int, bool) {
	if comma <= 0 || comma+1 >= len(tokens) || !isSpellListNoun(tokens[comma-1]) {
		return 0, false
	}
	start := comma - 1
	for start > 0 && isListRunToken(tokens[start-1]) {
		start--
	}
	end := comma + 1
	for end < len(tokens) && isListRunToken(tokens[end]) {
		end++
	}
	if end >= len(tokens) || !isSpellOrCardNoun(tokens[end]) {
		return 0, false
	}
	if !homogeneousSpellList(tokens[start:end]) {
		return 0, false
	}
	return end + 1, true
}

// isListRunToken reports whether the token can appear in the interior of a
// spell-type list: a list noun, a comma, or an "or"/"and" connector.
func isListRunToken(token shared.Token) bool {
	return isSpellListNoun(token) || token.Kind == shared.Comma || isListConjunction(token)
}

// isSpellOrCardNoun reports whether the token closes a spell-type list with a
// "spell" or "card" noun.
func isSpellOrCardNoun(token shared.Token) bool {
	return isSpellNoun(token) || equalWord(token, "card") || equalWord(token, "cards")
}

// homogeneousSpellList reports whether every list noun in the run shares a
// category: all card types (optionally "non"-prefixed), all colors, or all card
// subtypes.
func homogeneousSpellList(tokens []shared.Token) bool {
	sawType, sawColor, sawSubtype := false, false, false
	for _, token := range tokens {
		if !isSpellListNoun(token) {
			continue
		}
		word := strings.TrimPrefix(strings.ToLower(token.Text), "non")
		switch {
		case isTriggerCardTypeWord(word):
			sawType = true
		case isColorWord(word):
			sawColor = true
		default:
			sawSubtype = true
		}
	}
	categories := 0
	for _, saw := range []bool{sawType, sawColor, sawSubtype} {
		if saw {
			categories++
		}
	}
	return categories == 1
}

// isSpellListNoun reports whether the token is a card-type, color, or card-
// subtype word (optionally "non"-prefixed) that can appear as a member of a
// spell-type list.
func isSpellListNoun(token shared.Token) bool {
	if token.Kind != shared.Word {
		return false
	}
	word := strings.ToLower(token.Text)
	word = strings.TrimPrefix(word, "non")
	if isTriggerCardTypeWord(word) || isColorWord(word) {
		return true
	}
	_, ok := recognizeSubtypePhrase(word)
	return ok
}

// isTriggerCardTypeWord reports whether the lowercase word names a recognized
// card type.
func isTriggerCardTypeWord(word string) bool {
	cardType, ok := triggerCardType(word)
	return ok && cardType != TriggerCardTypeUnknown
}

// isColorWord reports whether the lowercase word names a single color.
func isColorWord(word string) bool {
	_, ok := recognizeColorWord(word)
	return ok
}
