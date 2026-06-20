package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseAttackBecameTargetUnionTriggerEventClause recognizes the event-union
// trigger "<subject> attacks or becomes the target of <a spell|a spell or
// ability>", e.g. "Whenever this creature attacks or becomes the target of a
// spell". Both verbs share one subject, so the trigger fires when either the
// attack event or the became-target event happens (CR 603.2). The became-target
// event is the primary clause because it carries the stack-object (spell vs.
// ability) filter; the attack event joins it through UnionKind.
func parseAttackBecameTargetUnionTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	index := syntaxWordsIndex(tokens, "attacks", "or", "becomes", "the", "target", "of")
	if index <= 0 {
		return nil
	}
	subject := parsePermanentEventSubject(tokens[:index], false, atoms)
	if !subject.ok || subject.oneOrMore {
		return nil
	}
	cause := tokens[index+6:]
	causeController := TriggerEventActorUnknown
	switch {
	case endsWithSyntaxWords(cause, "you", "control"):
		cause = cause[:len(cause)-2]
		causeController = TriggerEventActorYou
	case endsWithSyntaxWords(cause, "an", "opponent", "controls"):
		cause = cause[:len(cause)-3]
		causeController = TriggerEventActorOpponent
	default:
	}
	var stackObject TriggerEventStackObject
	switch {
	case syntaxWordsEqual(cause, "a", "spell"):
		stackObject = TriggerEventStackObject{Kind: TriggerEventStackObjectSpell, Span: shared.SpanOf(cause)}
	case syntaxWordsEqual(cause, "a", "spell", "or", "ability"):
		stackObject = TriggerEventStackObject{Kind: TriggerEventStackObjectAny, Span: shared.SpanOf(cause)}
	default:
		return nil
	}
	return &TriggerEventClause{
		Kind:            TriggerEventKindBecameTarget,
		UnionKind:       TriggerEventKindAttack,
		Subject:         subject.subject,
		Controller:      subject.controller,
		ExcludeSelf:     subject.excludeSelf,
		StackObject:     stackObject,
		CauseController: causeController,
	}
}
