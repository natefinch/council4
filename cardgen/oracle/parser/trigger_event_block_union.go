package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseBlockBecameBlockedUnionTriggerEventClause recognizes the event-union
// trigger "<subject> blocks or becomes blocked" (and the mirrored "<subject>
// becomes blocked or blocks"), e.g. "Whenever this creature blocks or becomes
// blocked". Both verbs share one subject, so the trigger fires when either the
// blocker-declared event or the became-blocked event happens (CR 603.2). This is
// the combat trigger that the Bushido keyword expands to. The block event is the
// primary clause; the became-blocked event joins it through UnionKind.
func parseBlockBecameBlockedUnionTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if index := syntaxWordsIndex(tokens, "blocks", "or", "becomes", "blocked"); index > 0 {
		if index+4 != len(tokens) {
			return nil
		}
		return blockBecameBlockedUnionClause(tokens[:index], atoms)
	}
	if index := syntaxWordsIndex(tokens, "becomes", "blocked", "or", "blocks"); index > 0 {
		if index+4 != len(tokens) {
			return nil
		}
		return blockBecameBlockedUnionClause(tokens[:index], atoms)
	}
	return nil
}

func blockBecameBlockedUnionClause(prefix []shared.Token, atoms Atoms) *TriggerEventClause {
	subject := parsePermanentEventSubject(prefix, false, atoms)
	if !subject.ok || subject.oneOrMore {
		return nil
	}
	return &TriggerEventClause{
		Kind:        TriggerEventKindBlock,
		UnionKind:   TriggerEventKindBecameBlocked,
		Subject:     subject.subject,
		Controller:  subject.controller,
		ExcludeSelf: subject.excludeSelf,
	}
}

// parseAttackBlockUnionTriggerEventClause recognizes the event-union trigger
// "<subject> attacks or blocks" (and the mirrored "<subject> blocks or
// attacks"), e.g. "Whenever this creature attacks or blocks". Both verbs share
// one subject, so the trigger fires when either the attacker-declared event or
// the blocker-declared event happens. The attack event is the primary clause;
// the block event joins it through UnionKind.
func parseAttackBlockUnionTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if index := syntaxWordsIndex(tokens, "attacks", "or", "blocks"); index > 0 {
		if index+3 != len(tokens) {
			return nil
		}
		return attackBlockUnionClause(tokens[:index], atoms)
	}
	if index := syntaxWordsIndex(tokens, "blocks", "or", "attacks"); index > 0 {
		if index+3 != len(tokens) {
			return nil
		}
		return attackBlockUnionClause(tokens[:index], atoms)
	}
	return nil
}

func attackBlockUnionClause(prefix []shared.Token, atoms Atoms) *TriggerEventClause {
	subject := parsePermanentEventSubject(prefix, false, atoms)
	if !subject.ok || subject.oneOrMore {
		return nil
	}
	return &TriggerEventClause{
		Kind:        TriggerEventKindAttack,
		UnionKind:   TriggerEventKindBlock,
		Subject:     subject.subject,
		Controller:  subject.controller,
		ExcludeSelf: subject.excludeSelf,
	}
}
