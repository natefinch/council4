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
//
// The "blocks or becomes blocked" order also accepts a trailing "by <creature
// selection>" naming the other combat participant, e.g. "Whenever this creature
// blocks or becomes blocked by a creature" or "...by a nonblack creature". The
// filter restricts both constituent events to combats whose other creature
// matches the selection: for the block event the creature this source blocks,
// and for the became-blocked event the creature blocking the source. Both events
// carry that other creature as the trigger's related permanent, so the selection
// joins the primary (block) clause as its RelatedSelection.
func parseBlockBecameBlockedUnionTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	if index := syntaxWordsIndex(tokens, "blocks", "or", "becomes", "blocked"); index > 0 {
		return blockBecameBlockedUnionClause(tokens[:index], tokens[index+4:], atoms)
	}
	if index := syntaxWordsIndex(tokens, "becomes", "blocked", "or", "blocks"); index > 0 {
		if index+4 != len(tokens) {
			return nil
		}
		return blockBecameBlockedUnionClause(tokens[:index], nil, atoms)
	}
	return nil
}

func blockBecameBlockedUnionClause(prefix, trailing []shared.Token, atoms Atoms) *TriggerEventClause {
	subject := parsePermanentEventSubject(prefix, false, atoms)
	if !subject.ok || subject.oneOrMore {
		return nil
	}
	clause := &TriggerEventClause{
		Kind:        TriggerEventKindBlock,
		UnionKind:   TriggerEventKindBecameBlocked,
		Subject:     subject.subject,
		Controller:  subject.controller,
		ExcludeSelf: subject.excludeSelf,
	}
	if len(trailing) == 0 {
		return clause
	}
	rest, ok := cutSyntaxWords(trailing, "by")
	if !ok {
		return nil
	}
	related, ok := parseRelatedSelectionPhrase(rest)
	if !ok || !selectionHasType(related, TriggerCardTypeCreature) {
		return nil
	}
	clause.RelatedSelection = related
	return clause
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
