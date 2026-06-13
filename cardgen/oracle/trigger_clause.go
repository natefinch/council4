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
	if clause.Introduction.Kind == TriggerIntroductionAt {
		clause.PhaseStep = parsePhaseStepTriggerClause(clauseTokens[1:])
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
	player := PhaseStepPlayerRelation{Kind: PhaseStepPlayerRelationAny}
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
			determiner.player.Kind == PhaseStepPlayerRelationAny ||
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
	subject, ok := parsePhaseStepAttachedSubject(subjectTokens)
	if !ok {
		return PhaseStepTriggerClause{}, false
	}
	return PhaseStepTriggerClause{
		Quantifier: PhaseStepQuantifier{
			Kind: PhaseStepQuantifierSingle,
			Span: tokens[0].Span,
		},
		Player: PhaseStepPlayerRelation{
			Kind:            PhaseStepPlayerRelationAttachedController,
			Span:            spanOf(tokens[of:]),
			AttachedSubject: subject,
		},
		Name: name,
	}, true
}

type phaseStepDeterminer struct {
	quantifier PhaseStepQuantifier
	player     PhaseStepPlayerRelation
	remainder  []Token
}

func parsePhaseStepDeterminer(tokens []Token) (phaseStepDeterminer, bool) {
	if len(tokens) == 0 {
		return phaseStepDeterminer{}, false
	}
	if len(tokens) >= 3 &&
		equalWord(tokens[0], "each") &&
		equalWord(tokens[1], "of") &&
		equalWord(tokens[2], "your") {
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierEachOf, Span: spanOf(tokens[:2])},
			player:     PhaseStepPlayerRelation{Kind: PhaseStepPlayerRelationYou, Span: tokens[2].Span},
			remainder:  tokens[3:],
		}, true
	}
	if len(tokens) >= 2 && equalWord(tokens[0], "each") {
		switch {
		case equalWord(tokens[1], "player's"):
			return phaseStepDeterminer{
				quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierEach, Span: tokens[0].Span},
				player:     PhaseStepPlayerRelation{Kind: PhaseStepPlayerRelationAny, Span: tokens[1].Span},
				remainder:  tokens[2:],
			}, true
		case equalWord(tokens[1], "opponent's"):
			return phaseStepDeterminer{
				quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierEach, Span: tokens[0].Span},
				player:     PhaseStepPlayerRelation{Kind: PhaseStepPlayerRelationOpponent, Span: tokens[1].Span},
				remainder:  tokens[2:],
			}, true
		}
	}
	if len(tokens) >= 2 &&
		equalWord(tokens[0], "its") &&
		equalWord(tokens[1], "controller's") {
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierSingle, Span: spanOf(tokens[:2])},
			player:     PhaseStepPlayerRelation{Kind: PhaseStepPlayerRelationSourceController, Span: spanOf(tokens[:2])},
			remainder:  tokens[2:],
		}, true
	}
	switch {
	case equalWord(tokens[0], "your"):
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierSingle, Span: tokens[0].Span},
			player:     PhaseStepPlayerRelation{Kind: PhaseStepPlayerRelationYou, Span: tokens[0].Span},
			remainder:  tokens[1:],
		}, true
	case equalWord(tokens[0], "the"):
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierSingle, Span: tokens[0].Span},
			player:     PhaseStepPlayerRelation{Kind: PhaseStepPlayerRelationAny, Span: tokens[0].Span},
			remainder:  tokens[1:],
		}, true
	case equalWord(tokens[0], "each"):
		return phaseStepDeterminer{
			quantifier: PhaseStepQuantifier{Kind: PhaseStepQuantifierEach, Span: tokens[0].Span},
			player:     PhaseStepPlayerRelation{Kind: PhaseStepPlayerRelationAny, Span: tokens[0].Span},
			remainder:  tokens[1:],
		}, true
	default:
		return phaseStepDeterminer{}, false
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

func parsePhaseStepAttachedSubject(tokens []Token) (PhaseStepAttachedSubject, bool) {
	subjectTokens := cloneTokens(tokens)
	last := &subjectTokens[len(subjectTokens)-1]
	last.Text = strings.TrimSuffix(strings.TrimSuffix(last.Text, "'s"), "’s")
	if last.Text == "" {
		return PhaseStepAttachedSubject{}, false
	}
	parsed := parseCombatPermanentSelection("a "+strings.ToLower(joinedSourceText(subjectTokens)), false)
	if !parsed.ok || parsed.excludeSelf {
		return PhaseStepAttachedSubject{}, false
	}
	parsed.selection.Controller = parsed.controller
	return PhaseStepAttachedSubject{
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
