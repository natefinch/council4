package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
	"github.com/natefinch/council4/mtg/game/types"
)

// parseAnimateTargetEffect recognizes the one-shot continuous target-animation
// "[Until end of turn,] target land becomes a N/N [<color>...] <subtype>...
// creature [with <keyword>...] [until end of turn] [that's still a land]."
// (Animate Land, Vivify, Hydroform, Kamahl, Soilshaper, Lifespark Spellbomb; CR
// 613). It is the targeted broadening of parseAnimateSelfEffect: the affected
// permanent is a single "target land" left in the sentence for the target
// machinery to extract, rather than the ability's own source.
//
// Exactly one "until end of turn" must be present, either as a leading clause
// ("Until end of turn, target land becomes ...") or as a trailing phrase
// ("... becomes a 3/3 creature until end of turn."); both or neither fail
// closed, which also rejects the permanent ("lasts indefinitely") form. The
// "still a land" confirmation may appear inline as a trailing relative clause
// ("...creature that's still a land.") or, for the trailing-duration form, as a
// following "It's still a land." sentence folded on by
// foldAnimateTargetStillSentence. The base power/toughness is a literal N/N; the
// colors set the land's color set; the named subtypes are added creature types;
// and a "with" clause grants supported keyword(s). Any richer shape — an X/X
// amount, a non-land subject, an "a copy"/"in addition" tail, a quoted granted
// ability, or an unsupported keyword — fails closed so those cards stay
// unsupported.
func parseAnimateTargetEffect(sentence Sentence, tokens []shared.Token, atoms Atoms) ([]EffectSyntax, bool) {
	body := semanticEffectTokens(tokens)
	if len(body) == 0 || body[len(body)-1].Kind != shared.Period {
		return nil, false
	}
	inner := body[:len(body)-1]

	remaining, leadingDuration := stripLeadingDurationClause(inner, atoms)
	leadingUntilEndOfTurn := leadingDuration == EffectDurationUntilEndOfTurn

	work, inlineStill := trimInlineStillALand(remaining)
	work, countedType, dynamicPT := trimAnimateWhereXClause(work)
	work, trailingUntilEndOfTurn := becomeCopyTrimUntilEndOfTurn(work)
	if leadingUntilEndOfTurn == trailingUntilEndOfTurn {
		return nil, false
	}

	becomesIndex := -1
	for i := range work {
		if equalWord(work[i], "becomes") {
			becomesIndex = i
			break
		}
	}
	if becomesIndex < 1 || !equalWord(work[0], "target") {
		return nil, false
	}
	if !animateTargetSubjectIsLand(work[1:becomesIndex]) {
		return nil, false
	}

	cursor := becomesIndex + 1
	if cursor >= len(work) || (!equalWord(work[cursor], "a") && !equalWord(work[cursor], "an")) {
		return nil, false
	}
	cursor++
	pt, ok := animateTargetPowerToughness(work, cursor, dynamicPT)
	if !ok {
		return nil, false
	}
	cursor = pt.Next

	colors, cursor := parseAnimateSelfColorRun(work, cursor)

	characteristics, cursor, ok := parseAnimateSelfCharacteristicRun(work, cursor, atoms)
	if !ok || !characteristics.HasCreature || characteristics.AddArtifact {
		return nil, false
	}

	keywords, everyCreatureType, ok := parseAnimateSelfRiders(work[cursor:], 0)
	if !ok || everyCreatureType {
		return nil, false
	}

	// The inline relative clause is only valid alongside the leading-duration
	// form; the trailing-duration form's confirmation is a separate sentence.
	if inlineStill && !leadingUntilEndOfTurn {
		return nil, false
	}

	var dynamicPTPayload *AnimateDynamicPowerToughness
	if dynamicPT {
		dynamicPTPayload = &AnimateDynamicPowerToughness{ControlledType: countedType}
	}

	effect := EffectSyntax{
		Kind:       EffectAnimateTarget,
		Context:    EffectContextController,
		Span:       sentence.Span,
		ClauseSpan: sentence.Span,
		Text:       sentence.Text,
		Tokens:     append([]shared.Token(nil), body...),
		Duration:   EffectDurationUntilEndOfTurn,
		AnimateTarget: &AnimateSelfSyntax{
			Power:                 pt.Power,
			Toughness:             pt.Toughness,
			Colors:                colors,
			Subtypes:              characteristics.Subtypes,
			EveryCreatureType:     false,
			Keywords:              keywords,
			DynamicPowerToughness: dynamicPTPayload,
		},
	}
	return []EffectSyntax{effect}, true
}

// animateTargetPowerToughness reads the animated base power/toughness at index,
// dispatching on whether the sentence carried a "where X is ..." clause. With
// the clause it matches the variable "X/X" form (reporting zero placeholders
// bound by the clause); without it, the literal "N/N" form. Coupling the two
// means an unbound "X/X" (no clause) and a bound clause without "X/X" both fail
// closed.
func animateTargetPowerToughness(tokens []shared.Token, index int, dynamic bool) (powerToughness, bool) {
	if dynamic {
		return parseVariableXAnimatePowerToughness(tokens, index)
	}
	return parsePowerToughness(tokens, index)
}

// parseVariableXAnimatePowerToughness matches a variable "X/X" base
// power/toughness at index (Destiny Spinner's "becomes an X/X ... creature").
// The reported power/toughness are zero; the actual value is bound by the
// caller's "where X is the number of <type> you control" clause and locked at
// resolution.
func parseVariableXAnimatePowerToughness(tokens []shared.Token, index int) (powerToughness, bool) {
	if index+2 >= len(tokens) ||
		!equalWord(tokens[index], "X") ||
		tokens[index+1].Kind != shared.Slash ||
		!equalWord(tokens[index+2], "X") {
		return powerToughness{}, false
	}
	return powerToughness{Power: 0, Toughness: 0, Next: index + 3}, true
}

// trimAnimateWhereXClause removes a trailing ", where X is the number of <type>
// you control" clause that binds a variable X/X target animation (Destiny
// Spinner) and reports the counted controlled permanent type. The clause must be
// the exact supported shape running to the end of the animation sentence; any
// other "where X is ..." wording leaves the tokens unchanged and reports false,
// so a fixed N/N animation is untouched and an unsupported dynamic clause fails
// closed downstream (the X/X base P/T is then left unbound).
func trimAnimateWhereXClause(work []shared.Token) ([]shared.Token, types.Card, bool) {
	whereIdx := -1
	for i := range work {
		if equalWord(work[i], "where") {
			whereIdx = i
			break
		}
	}
	if whereIdx < 1 {
		return work, "", false
	}
	clause := work[whereIdx:]
	if len(clause) != 9 ||
		!equalWord(clause[1], "X") || !equalWord(clause[2], "is") ||
		!equalWord(clause[3], "the") || !equalWord(clause[4], "number") ||
		!equalWord(clause[5], "of") || !equalWord(clause[7], "you") ||
		!equalWord(clause[8], "control") {
		return work, "", false
	}
	counted, ok := groupEntersTappedPermanentType(clause[6].Text)
	if !ok {
		return work, "", false
	}
	head := work[:whereIdx]
	if len(head) > 0 && head[len(head)-1].Kind == shared.Comma {
		head = head[:len(head)-1]
	}
	return head, counted, true
}

// animateTargetSubjectIsLand reports whether the words between "target" and
// "becomes" name a single land target ("land" or "land you control"). Any other
// noun phrase, connector, or verb fails closed so a compound or non-land subject
// is not silently animated.
func animateTargetSubjectIsLand(subject []shared.Token) bool {
	switch len(subject) {
	case 1:
		return equalWord(subject[0], "land")
	case 3:
		return equalWord(subject[0], "land") &&
			equalWord(subject[1], "you") && equalWord(subject[2], "control")
	default:
		return false
	}
}

// trimInlineStillALand removes a trailing inline "that's still a land" /
// "that is still a land" / "it's still a land" relative clause from the body,
// returning the trimmed tokens and whether the clause was present. The clause
// carries no new semantics — the type layer adds the creature type rather than
// setting it, so the targeted land keeps its land type — but must be consumed so
// it does not leave the sentence partially recognized.
func trimInlineStillALand(tokens []shared.Token) ([]shared.Token, bool) {
	if width, ok := inlineStillALandWidth(tokens); ok {
		return tokens[:len(tokens)-width], true
	}
	return tokens, false
}

// inlineStillALandWidth reports the token width of a trailing inline "still a
// land" confirmation clause, or false when the tokens do not end with one.
func inlineStillALandWidth(tokens []shared.Token) (int, bool) {
	words := normalizedWords(tokens)
	stillLand := []string{"still", "a", "land"}
	if len(words) < len(stillLand)+1 {
		return 0, false
	}
	tail := words[len(words)-len(stillLand):]
	if tail[0] != stillLand[0] || tail[1] != stillLand[1] || tail[2] != stillLand[2] {
		return 0, false
	}
	lead := words[len(words)-len(stillLand)-1]
	switch lead {
	case "that's", "it's":
		return len(stillLand) + 1, true
	}
	if len(words) >= len(stillLand)+2 {
		two := words[len(words)-len(stillLand)-2 : len(words)-len(stillLand)]
		if (two[0] == "that" || two[0] == "it") && two[1] == "is" {
			return len(stillLand) + 2, true
		}
	}
	return 0, false
}

// abilityHasAnimateTarget reports whether the ability carries a recognized
// EffectAnimateTarget effect.
func abilityHasAnimateTarget(ability *Ability) bool {
	for i := range ability.Sentences {
		for j := range ability.Sentences[i].Effects {
			if ability.Sentences[i].Effects[j].Kind == EffectAnimateTarget {
				return true
			}
		}
	}
	return false
}

// stripAnimateTargetSemantics clears the residual reference, keyword, and
// condition semantics the general scans re-derive for an ability whose resolving
// content is an EffectAnimateTarget. The animation clause names a keyword ("with
// flying") and a "target land" the general scans would otherwise surface as
// ability-level keywords or references, over-counting the ability and failing
// the lowering coverage gate. It mirrors stripAnimateSelfSemantics.
func stripAnimateTargetSemantics(abilities []Ability) {
	for i := range abilities {
		ability := &abilities[i]
		if !abilityHasAnimateTarget(ability) {
			continue
		}
		ability.SemanticKeywords = nil
		ability.ConditionBoundaries = nil
		ability.EventHistoryConditions = nil
		ability.ConditionClauses = nil
		ability.ConditionSegments = nil
		ability.TriggerConditionSegments = nil
		ability.StaticDeclarations = nil
	}
}

// foldAnimateTargetStillSentence extends an EffectAnimateTarget effect's span to
// cover the immediately following "It's still a land." confirmation sentence
// (Vivify, Hydroform, Kamahl, Soilshaper). That confirmation carries no new
// semantics — the type layer adds rather than sets — but its tokens would
// otherwise be left uncovered and fail the lowering coverage gate. Folding the
// span onto the recognized effect accounts for those tokens without adding any
// resolving behavior. The inline relative-clause form has no separate sentence
// and is unaffected.
func foldAnimateTargetStillSentence(ability *Ability) {
	for i := range ability.Sentences {
		if !sentenceHasAnimateTarget(&ability.Sentences[i]) || i+1 >= len(ability.Sentences) {
			continue
		}
		next := &ability.Sentences[i+1]
		if len(next.Effects) != 0 || !isStillSourceTypeSentence(next.Tokens) {
			continue
		}
		sentence := &ability.Sentences[i]
		for e := range sentence.Effects {
			if sentence.Effects[e].Kind != EffectAnimateTarget {
				continue
			}
			sentence.Effects[e].Span.End = next.Span.End
			sentence.Effects[e].ClauseSpan.End = next.Span.End
		}
	}
}

// sentenceHasAnimateTarget reports whether the sentence carries an
// EffectAnimateTarget effect.
func sentenceHasAnimateTarget(sentence *Sentence) bool {
	for j := range sentence.Effects {
		if sentence.Effects[j].Kind == EffectAnimateTarget {
			return true
		}
	}
	return false
}
