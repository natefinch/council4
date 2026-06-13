package oracle

import "strings"

func parseTriggerClause(source string, tokens []Token) *TriggerClause {
	end := triggerBodyComma(tokens)
	if end < 0 {
		end = len(tokens)
	}
	clauseTokens := tokens[:end]
	if len(clauseTokens) == 0 {
		return nil
	}
	clause := &TriggerClause{
		Span:   spanOf(clauseTokens),
		Text:   sliceSpan(source, spanOf(clauseTokens)),
		Tokens: cloneTokens(clauseTokens),
		Introduction: TriggerIntroduction{
			Kind: triggerIntroductionKind(clauseTokens[0]),
			Span: clauseTokens[0].Span,
		},
	}
	if len(clauseTokens) == 1 {
		return clause
	}
	clause.Event = phraseFromTokens(source, clauseTokens[1:])
	switch clause.Introduction.Kind {
	case TriggerIntroductionAt:
		clause.PhaseStep = parsePhaseStepTriggerClause(clauseTokens[1:])
	case TriggerIntroductionWhen, TriggerIntroductionWhenever:
		clause.PlayerEvent = parsePlayerEventTriggerClause(clauseTokens[1:], clause.Introduction.Kind)
	default:
	}
	return clause
}

func triggerIntroductionKind(token Token) TriggerIntroductionKind {
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

func parsePhaseStepTriggerClause(tokens []Token) *PhaseStepTriggerClause {
	if clause, ok := parseStandaloneEndOfCombat(tokens); ok {
		return &clause
	}
	event, ok := cutSyntaxWords(tokens, "the", "beginning", "of")
	if !ok {
		return nil
	}
	for _, parse := range []func([]Token) (PhaseStepTriggerClause, bool){
		parseAttachedControllerPhaseStep,
		parseTurnQualifiedPhaseStep,
		parseRelationPhaseStep,
	} {
		clause, ok := parse(event)
		if ok {
			clause.Span = spanOf(tokens)
			return &clause
		}
	}
	return nil
}

func parseStandaloneEndOfCombat(tokens []Token) (PhaseStepTriggerClause, bool) {
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
		Span:       spanOf(tokens),
		Quantifier: quantifier,
		Player:     player,
		Name:       name,
	}, true
}

func parseRelationPhaseStep(tokens []Token) (PhaseStepTriggerClause, bool) {
	determiner, ok := parsePhaseStepDeterminer(tokens)
	if !ok {
		return PhaseStepTriggerClause{}, false
	}
	name, ok := parsePhaseStepName(
		determiner.remainder,
		determiner.quantifier.Kind == PhaseStepQuantifierEachOf,
	)
	if !ok {
		return PhaseStepTriggerClause{}, false
	}
	return PhaseStepTriggerClause{
		Quantifier: determiner.quantifier,
		Player:     determiner.player,
		Name:       name,
	}, true
}

func parseTurnQualifiedPhaseStep(tokens []Token) (PhaseStepTriggerClause, bool) {
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

func parseAttachedControllerPhaseStep(tokens []Token) (PhaseStepTriggerClause, bool) {
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
			Span:            spanOf(tokens[of:]),
			AttachedSubject: subject,
		},
		Name: name,
	}, true
}

type phaseStepDeterminer struct {
	quantifier PhaseStepQuantifier
	player     TriggerPlayerSelector
	remainder  []Token
}

func parsePhaseStepDeterminer(tokens []Token) (phaseStepDeterminer, bool) {
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
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierEachOf, Span: spanOf(tokens[:2])},
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
			parsed.player.Kind == TriggerPlayerSelectorSourceController) {
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

type triggerPlayerSelectorForm uint8

const (
	triggerPlayerSelectorFormUnknown triggerPlayerSelectorForm = iota
	triggerPlayerSelectorSubject
	triggerPlayerSelectorPossessive
)

type triggerPlayerSelectorParse struct {
	player    TriggerPlayerSelector
	remainder []Token
	form      triggerPlayerSelectorForm
	ok        bool
}

func parseTriggerPlayerSelector(tokens []Token) triggerPlayerSelectorParse {
	switch {
	case len(tokens) >= 2 && equalWord(tokens[0], "an") && equalWord(tokens[1], "opponent"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorOpponent, Span: spanOf(tokens[:2])},
			remainder: tokens[2:],
			form:      triggerPlayerSelectorSubject,
			ok:        true,
		}
	case len(tokens) >= 2 && equalWord(tokens[0], "a") && equalWord(tokens[1], "player"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorAny, Span: spanOf(tokens[:2])},
			remainder: tokens[2:],
			form:      triggerPlayerSelectorSubject,
			ok:        true,
		}
	case len(tokens) >= 2 && equalWord(tokens[0], "its") && equalWord(tokens[1], "controller's"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorSourceController, Span: spanOf(tokens[:2])},
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
	case len(tokens) >= 1 && equalWord(tokens[0], "opponent's"):
		return triggerPlayerSelectorParse{
			player:    TriggerPlayerSelector{Kind: TriggerPlayerSelectorOpponent, Span: tokens[0].Span},
			remainder: tokens[1:],
			form:      triggerPlayerSelectorPossessive,
			ok:        true,
		}
	default:
		return triggerPlayerSelectorParse{}
	}
}

func parsePlayerEventTriggerClause(tokens []Token, introduction TriggerIntroductionKind) *PlayerEventTriggerClause {
	parsedPlayer := parseTriggerPlayerSelector(tokens)
	if !parsedPlayer.ok || parsedPlayer.form != triggerPlayerSelectorSubject {
		return nil
	}
	action, rest, ok := parsePlayerEventAction(parsedPlayer.remainder, parsedPlayer.player.Kind)
	if !ok {
		return nil
	}
	card, occurrence, ok := parsePlayerEventModifiers(rest, action.Kind, parsedPlayer.player.Kind)
	if !ok ||
		occurrence.Kind == PlayerEventOccurrenceAny && introduction != TriggerIntroductionWhenever {
		return nil
	}
	return &PlayerEventTriggerClause{
		Span:       spanOf(tokens),
		Player:     parsedPlayer.player,
		Action:     action,
		Card:       card,
		Occurrence: occurrence,
	}
}

func parsePlayerEventAction(
	tokens []Token,
	player TriggerPlayerSelectorKind,
) (PlayerEventAction, []Token, bool) {
	if len(tokens) == 0 {
		return PlayerEventAction{}, nil, false
	}
	verbMatches := func(token Token, secondPerson, thirdPerson string) bool {
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
			Span: spanOf(tokens[:3]),
		}, tokens[3:], true
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
		return PlayerEventAction{Kind: form.kind, Span: spanOf(tokens[:end])}, tokens[end:], true
	}
	return PlayerEventAction{}, nil, false
}

func parsePlayerEventModifiers(
	tokens []Token,
	action PlayerEventActionKind,
	player TriggerPlayerSelectorKind,
) (PlayerEventCard, PlayerEventOccurrence, bool) {
	card := PlayerEventCard{Kind: PlayerEventCardNone}
	occurrence := PlayerEventOccurrence{Kind: PlayerEventOccurrenceAny}
	rest := tokens
	if playerEventActionHasCard(action) {
		parsed := parsePlayerEventCard(rest, action, player)
		if !parsed.ok {
			return PlayerEventCard{}, PlayerEventOccurrence{}, false
		}
		card = parsed.card
		occurrence = parsed.occurrence
		rest = parsed.remainder
	}
	if next, ok := cutSyntaxWords(rest, "for", "the", "first", "time", "each", "turn"); ok {
		if occurrence.Kind != PlayerEventOccurrenceAny || !playerEventFirstEachTurnAllowed(action, player) {
			return PlayerEventCard{}, PlayerEventOccurrence{}, false
		}
		occurrence = PlayerEventOccurrence{
			Kind:    PlayerEventOccurrenceFirstEachTurn,
			Span:    spanOf(rest),
			Ordinal: 1,
		}
		rest = next
	}
	if len(rest) != 0 {
		return PlayerEventCard{}, PlayerEventOccurrence{}, false
	}
	return card, occurrence, true
}

type playerEventCardParse struct {
	card       PlayerEventCard
	occurrence PlayerEventOccurrence
	remainder  []Token
	ok         bool
}

func parsePlayerEventCard(
	tokens []Token,
	action PlayerEventActionKind,
	player TriggerPlayerSelectorKind,
) playerEventCardParse {
	occurrence := PlayerEventOccurrence{Kind: PlayerEventOccurrenceAny}
	if rest, ok := cutSyntaxWords(tokens, "a", "card"); ok {
		return playerEventCardParse{
			card:       PlayerEventCard{Kind: PlayerEventCardSingle, Span: spanOf(tokens[:2])},
			occurrence: occurrence,
			remainder:  rest,
			ok:         true,
		}
	}
	if rest, ok := cutSyntaxWords(tokens, "one", "or", "more", "cards"); ok &&
		action == PlayerEventActionDiscard {
		return playerEventCardParse{
			card:       PlayerEventCard{Kind: PlayerEventCardOneOrMore, Span: spanOf(tokens[:4])},
			occurrence: occurrence,
			remainder:  rest,
			ok:         true,
		}
	}
	if rest, ok := cutSyntaxWords(tokens, "another", "card"); ok &&
		(action == PlayerEventActionDiscard ||
			action == PlayerEventActionCycle ||
			action == PlayerEventActionCycleOrDiscard) {
		return playerEventCardParse{
			card:       PlayerEventCard{Kind: PlayerEventCardAnother, Span: spanOf(tokens[:2])},
			occurrence: occurrence,
			remainder:  rest,
			ok:         true,
		}
	}
	if action != PlayerEventActionDraw || len(tokens) < 5 {
		return playerEventCardParse{}
	}
	possessive := "their"
	if player == TriggerPlayerSelectorYou {
		possessive = "your"
	}
	if !equalWord(tokens[0], possessive) ||
		!equalWord(tokens[2], "card") ||
		!equalWord(tokens[3], "each") ||
		!equalWord(tokens[4], "turn") {
		return playerEventCardParse{}
	}
	ordinal, ok := parsePlayerEventOrdinal(tokens[1])
	if !ok {
		return playerEventCardParse{}
	}
	return playerEventCardParse{
		card: PlayerEventCard{
			Kind: PlayerEventCardSingle,
			Span: spanOf(tokens[:3]),
		},
		occurrence: PlayerEventOccurrence{
			Kind:    PlayerEventOccurrenceOrdinalEachTurn,
			Span:    spanOf(tokens[1:5]),
			Ordinal: ordinal,
		},
		remainder: tokens[5:],
		ok:        true,
	}
}

func playerEventActionHasCard(action PlayerEventActionKind) bool {
	switch action {
	case PlayerEventActionDraw,
		PlayerEventActionDiscard,
		PlayerEventActionCycle,
		PlayerEventActionCycleOrDiscard:
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
	case PlayerEventActionGainLife, PlayerEventActionLoseLife:
		return player != TriggerPlayerSelectorAny
	default:
		return false
	}
}

func parsePlayerEventOrdinal(token Token) (int, bool) {
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

func parsePhaseStepName(tokens []Token, plural bool) (PhaseStepName, bool) {
	for _, form := range phaseStepNameForms {
		if form.plural == plural && syntaxWordsEqual(tokens, form.words...) {
			return PhaseStepName{Kind: form.kind, Span: spanOf(tokens)}, true
		}
	}
	return PhaseStepName{}, false
}

func parseTriggerAttachedSubject(tokens []Token) (TriggerAttachedSubject, bool) {
	subjectTokens := cloneTokens(tokens)
	last := &subjectTokens[len(subjectTokens)-1]
	last.Text = strings.TrimSuffix(strings.TrimSuffix(last.Text, "'s"), "’s")
	if last.Text == "" {
		return TriggerAttachedSubject{}, false
	}
	parsed := parseCombatPermanentSelection("a "+strings.ToLower(joinedSourceText(subjectTokens)), false)
	if !parsed.ok || parsed.excludeSelf {
		return TriggerAttachedSubject{}, false
	}
	parsed.selection.Controller = parsed.controller
	return TriggerAttachedSubject{
		Span:      spanOf(tokens),
		Selection: parsed.selection,
	}, true
}

func cutSyntaxWords(tokens []Token, words ...string) ([]Token, bool) {
	if len(tokens) < len(words) || !syntaxWordsEqual(tokens[:len(words)], words...) {
		return nil, false
	}
	return tokens[len(words):], true
}

func syntaxWordsEqual(tokens []Token, words ...string) bool {
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

func syntaxWordIndex(tokens []Token, word string) int {
	for i, token := range tokens {
		if equalWord(token, word) {
			return i
		}
	}
	return -1
}
