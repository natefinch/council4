package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

func parseFightTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	for _, verb := range []string{"fight", "fights"} {
		prefix, ok := stripTokenSuffix(tokens, verb)
		if !ok {
			continue
		}
		subject := parsePermanentEventSubject(prefix, verb == "fight", atoms)
		if !subject.ok {
			return nil
		}
		return &TriggerEventClause{
			Kind:        TriggerEventKindFight,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
		}
	}
	return nil
}

func parseFightBecameBlockedUnionTriggerEventClause(
	tokens []shared.Token,
	intro TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if intro != TriggerIntroductionWhenever {
		return nil
	}
	for _, form := range []struct {
		words  []string
		plural bool
	}{
		{words: []string{"fight", "or", "become", "blocked"}, plural: true},
		{words: []string{"fights", "or", "becomes", "blocked"}},
		{words: []string{"become", "blocked", "or", "fight"}, plural: true},
		{words: []string{"becomes", "blocked", "or", "fights"}},
	} {
		prefix, ok := stripTokenSuffix(tokens, form.words...)
		if !ok {
			continue
		}
		subject := parsePermanentEventSubject(prefix, form.plural, atoms)
		if !subject.ok {
			return nil
		}
		return &TriggerEventClause{
			Kind:        TriggerEventKindFight,
			UnionKind:   TriggerEventKindBecameBlocked,
			Subject:     subject.subject,
			Controller:  subject.controller,
			ExcludeSelf: subject.excludeSelf,
			OneOrMore:   subject.oneOrMore,
		}
	}
	return nil
}
