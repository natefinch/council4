package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseDoorUnlockTriggerEventClause recognizes the self-source trigger "you
// unlock this door" on a Room enchantment half (CR 715). Casting a Room half
// unlocks that door as it enters, and a locked door can later be unlocked as a
// special action; either way "this door" is the ability's own source, so the
// clause compiles to a self-source trigger with no subject selection.
func parseDoorUnlockTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	_ Atoms,
	_ string,
) *TriggerEventClause {
	if !syntaxWordsEqual(tokens, "you", "unlock", "this", "door") {
		return nil
	}
	return &TriggerEventClause{
		Span:    shared.SpanOf(tokens),
		Kind:    TriggerEventKindDoorUnlocked,
		Subject: TriggerEventSubject{Kind: TriggerEventSubjectSelf, Span: shared.SpanOf(tokens)},
	}
}
