package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parsePlayerAndControlledPermanentsSubject recognizes the subject of a resolving
// keyword grant that couples the controller with the permanents they control:
// "You and permanents you control gain hexproof until end of turn." (Dawn's
// Truce) and its gift-gated sibling "permanents you control also gain
// indestructible until end of turn.". The generic static-subject scan gives up on
// both shapes — the leading "you and" hides the group and the interposed "also"
// adverb separates the group from its verb — so recognize them here.
//
// It returns the controlled-permanents group as an EffectStaticSubjectSyntax
// whose Span covers exactly the "permanents you control" tokens (so exact
// reconstruction round-trips the group body), a playerAnd flag reporting the
// leading "you and" coupling (which drives the separate controller-scoped rule
// grant at lowering), and ok. Only the exact controlled-permanents group is
// mapped; every other group, and a bare group carrying neither the "you and"
// coupling nor the "also" adverb, fails closed so the ordinary recognizers keep
// ownership of the plain "permanents you control gain ..." form.
func parsePlayerAndControlledPermanentsSubject(tokens []shared.Token, atoms Atoms) (subject EffectStaticSubjectSyntax, playerAnd bool, ok bool) {
	rest := tokens
	// Strip a leading condition clause ("If the gift was promised, ...") the way
	// exact clause reconstruction does, so a gift-gated rider's group subject is
	// recognized; the condition itself is parsed separately at the ability level
	// and re-applied to the lowered instruction as a sequence gate.
	if len(rest) > 0 {
		if intro, _ := conditionIntroAt(rest, 0); intro != ConditionIntroUnknown {
			end := conditionClauseEnd(rest, 0)
			if end < len(rest) && rest[end].Kind == shared.Comma {
				rest = rest[end+1:]
			}
		}
	}
	if len(rest) >= 2 && equalWord(rest[0], "you") && equalWord(rest[1], "and") {
		playerAnd = true
		rest = rest[2:]
	}
	verbIndex := -1
	for i := range rest {
		if staticGroupVerb(rest[i]) || staticGroupVerbSingular(rest[i]) {
			verbIndex = i
			break
		}
	}
	if verbIndex <= 0 {
		return EffectStaticSubjectSyntax{}, false, false
	}
	groupTokens := rest[:verbIndex]
	also := false
	if len(groupTokens) > 0 && equalWord(groupTokens[len(groupTokens)-1], "also") {
		also = true
		groupTokens = groupTokens[:len(groupTokens)-1]
	}
	// Require at least one compound marker so the plain "permanents you control
	// gain ..." form (which the generic scan already recognizes) is never
	// hijacked here.
	if !playerAnd && !also {
		return EffectStaticSubjectSyntax{}, false, false
	}
	subject, ok = coordinatedGroupSubjectWords(groupTokens, atoms)
	if !ok || subject.Kind != EffectStaticSubjectControlledPermanents {
		return EffectStaticSubjectSyntax{}, false, false
	}
	subject.Span = shared.SpanOf(groupTokens)
	return subject, playerAnd, true
}
