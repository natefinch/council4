package parser

import "github.com/natefinch/council4/cardgen/oracle/shared"

// parseCompletedDungeonTriggerEventClause recognizes player-scoped dungeon
// completion events. Completion is authoritative runtime state emitted when a
// final room ability leaves the stack, not inferred from room text or position.
func parseCompletedDungeonTriggerEventClause(
	tokens []shared.Token,
	_ TriggerIntroductionKind,
	_ Atoms,
	_ string,
) *TriggerEventClause {
	var player TriggerPlayerSelector
	switch {
	case syntaxWordsEqual(tokens, "you", "complete", "a", "dungeon"):
		player = playerSelectorFromKind(TriggerPlayerSelectorYou, tokens[0].Span)
	case syntaxWordsEqual(tokens, "an", "opponent", "completes", "a", "dungeon"):
		player = playerSelectorFromKind(TriggerPlayerSelectorOpponent, shared.SpanOf(tokens[:2]))
	case syntaxWordsEqual(tokens, "a", "player", "completes", "a", "dungeon"):
		player = playerSelectorFromKind(TriggerPlayerSelectorAny, shared.SpanOf(tokens[:2]))
	default:
		return nil
	}
	return &TriggerEventClause{
		Kind:   TriggerEventKindCompletedDungeon,
		Player: player,
	}
}
