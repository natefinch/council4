package parser

import (
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

// parseTokenCreatedTriggerEventClause recognizes "you create <token subject>"
// triggers, e.g. "Whenever you create a token" or "Whenever you create one or
// more tokens".
func parseTokenCreatedTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	remaining, ok := cutSyntaxWords(tokens, "you", "create")
	if !ok {
		return nil
	}
	subject := parsePermanentEventSubject(remaining, false, atoms)
	if !subject.ok || !subject.subject.Selection.TokenOnly {
		return nil
	}
	return &TriggerEventClause{
		Kind:        TriggerEventKindTokenCreated,
		Actor:       TriggerEventActor{Kind: TriggerEventActorYou, Span: shared.SpanOf(tokens[:2])},
		Subject:     subject.subject,
		Controller:  subject.controller,
		ExcludeSelf: subject.excludeSelf,
		OneOrMore:   subject.oneOrMore,
	}
}

// parseTokenCreateSacrificeUnionTriggerEventClause recognizes the event-union
// trigger "you create or sacrifice <token subject>" (and its mirror "you
// sacrifice or create ..."). Both verbs share a single token subject, so the
// trigger fires when either event happens.
func parseTokenCreateSacrificeUnionTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	atoms Atoms,
	_ string,
) *TriggerEventClause {
	remaining, ok := cutSyntaxWords(tokens, "you")
	if !ok {
		return nil
	}
	var primary, secondary TriggerEventKind
	switch {
	case len(remaining) >= 3 && syntaxWordsEqual(remaining[:3], "create", "or", "sacrifice"):
		primary, secondary = TriggerEventKindTokenCreated, TriggerEventKindSacrificed
	case len(remaining) >= 3 && syntaxWordsEqual(remaining[:3], "sacrifice", "or", "create"):
		primary, secondary = TriggerEventKindSacrificed, TriggerEventKindTokenCreated
	default:
		return nil
	}
	subject := parsePermanentEventSubject(remaining[3:], false, atoms)
	if !subject.ok || !subject.subject.Selection.TokenOnly {
		return nil
	}
	return &TriggerEventClause{
		Kind:        primary,
		UnionKind:   secondary,
		Actor:       TriggerEventActor{Kind: TriggerEventActorYou, Span: tokens[0].Span},
		Subject:     subject.subject,
		Controller:  subject.controller,
		ExcludeSelf: subject.excludeSelf,
		OneOrMore:   subject.oneOrMore,
	}
}
