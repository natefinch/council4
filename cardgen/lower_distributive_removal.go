package cardgen

import (
	"github.com/natefinch/council4/cardgen/oracle/compiler"
	"github.com/natefinch/council4/cardgen/oracle/parser"
	"github.com/natefinch/council4/mtg/game"
)

// distributiveRemovalScopes maps a recognized clause scope to the runtime player
// group the primitive walks.
var distributiveRemovalScopes = map[parser.DistributiveRemovalScope]game.PlayerGroupReference{
	parser.DistributiveScopeEachPlayer:   game.AllPlayersReference(),
	parser.DistributiveScopeEachOpponent: game.OpponentsReference(),
}

// distributiveRemovalVerbs maps a recognized clause verb to the runtime removal.
var distributiveRemovalVerbs = map[parser.DistributiveRemovalVerb]game.DistributiveRemoval{
	parser.DistributiveVerbExile:   game.DistributiveRemovalExile,
	parser.DistributiveVerbDestroy: game.DistributiveRemovalDestroy,
}

// distributiveRemovalPrimitive builds the ForEachPlayer primitive for a
// recognized "For each <group>, <verb> up to one target <permanent> that player
// controls" clause, or reports false when effect is not that clause with the
// requested verb and duration.
//
// Callers name the verb and duration whose payoff they can pair, but not the
// scope: the group is read from the clause. That is the point of the collapse.
// "For each opponent, destroy up to one target creature that player controls"
// with a token payoff needs no new parser, lowering, or runtime code even though
// no card had that wording when this was written.
func distributiveRemovalPrimitive(effect compiler.CompiledEffect, verb parser.DistributiveRemovalVerb, untilSourceLeaves bool, selection game.Selection, linkedKey game.LinkedKey) (game.ForEachPlayer, bool) {
	removal := effect.DistributiveRemoval
	if removal == nil || removal.Verb != verb || removal.UntilSourceLeaves != untilSourceLeaves {
		return game.ForEachPlayer{}, false
	}
	scope, ok := distributiveRemovalScopes[removal.Scope]
	if !ok {
		return game.ForEachPlayer{}, false
	}
	action, ok := distributiveRemovalVerbs[removal.Verb]
	if !ok {
		return game.ForEachPlayer{}, false
	}
	return game.ForEachPlayer{
		Scope:     scope,
		Chooser:   game.ControllerReference(),
		Selection: selection,
		Removal:   action,
		LinkedKey: linkedKey,
		// A per-opponent walk republishes its linked set on every resolution, so
		// the clause that publishes the set also clears it. The all-players Saga
		// chapters keep theirs for a later chapter to consume.
		ReplaceLink: removal.Scope == parser.DistributiveScopeEachOpponent,
	}, true
}
